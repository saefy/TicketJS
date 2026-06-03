package bridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"thermalik-bridge/internal/config"
	"thermalik-bridge/internal/domain"
	"thermalik-bridge/internal/store"
	"thermalik-bridge/internal/transport"
)

var ErrPrinterNotFound = errors.New("printer not found")

type Service struct {
	store     store.JobStore
	transport *transport.Manager
	printers  map[string]config.PrinterConfig
	queues    map[string]chan *domain.Job
}

func NewService(ctx context.Context, cfg config.Config, jobStore store.JobStore, manager *transport.Manager) (*Service, error) {
	service := &Service{
		store:     jobStore,
		transport: manager,
		printers:  make(map[string]config.PrinterConfig, len(cfg.Printers)),
		queues:    make(map[string]chan *domain.Job, len(cfg.Printers)),
	}

	for _, printer := range cfg.Printers {
		service.printers[printer.ID] = printer
		queue := make(chan *domain.Job, 64)
		service.queues[printer.ID] = queue
		go service.runPrinterWorker(ctx, printer, queue)
	}

	return service, nil
}

func (s *Service) Health() map[string]any {
	return map[string]any{
		"status":              "ok",
		"configured_printers": len(s.printers),
	}
}

func (s *Service) ListPrinters() []domain.Printer {
	printers := make([]domain.Printer, 0, len(s.printers))

	for _, printer := range s.printers {
		printers = append(printers, domain.Printer{
			ID:           printer.ID,
			Name:         printer.Name,
			Transport:    printer.Transport,
			Status:       printerStatus(printer),
			Default:      printer.Default,
			Capabilities: cloneMap(printer.Capabilities),
			Connection:   printerConnection(printer),
		})
	}

	return printers
}

func (s *Service) EnqueuePrint(ctx context.Context, printerID string, bytes []byte, render map[string]any) (*domain.Job, error) {
	printer, ok := s.printers[printerID]
	if !ok {
		return nil, ErrPrinterNotFound
	}

	jobID, err := generateJobID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	job := &domain.Job{
		ID:        jobID,
		Printer:   printer.ID,
		Status:    domain.JobStatusQueued,
		Bytes:     append([]byte(nil), bytes...),
		Render:    cloneMap(render),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Save(ctx, job); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.queues[printer.ID] <- job:
		return job, nil
	}
}

func (s *Service) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return s.store.Get(ctx, jobID)
}

func (s *Service) runPrinterWorker(ctx context.Context, printer config.PrinterConfig, queue <-chan *domain.Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-queue:
			if job == nil {
				return
			}

			s.processJob(ctx, printer, job)
		}
	}
}

func (s *Service) processJob(ctx context.Context, printer config.PrinterConfig, job *domain.Job) {
	job.Status = domain.JobStatusProcessing
	job.UpdatedAt = time.Now().UTC()
	_ = s.store.Save(ctx, job)

	err := s.transport.Send(ctx, printer, job.Bytes)

	job.UpdatedAt = time.Now().UTC()
	if err != nil {
		job.Status = domain.JobStatusFailed
		job.Error = err.Error()
	} else {
		job.Status = domain.JobStatusCompleted
		completedAt := time.Now().UTC()
		job.CompletedAt = &completedAt
	}

	_ = s.store.Save(ctx, job)
}

func generateJobID() (string, error) {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate job id: %w", err)
	}

	return "job_" + hex.EncodeToString(buffer), nil
}

func printerStatus(printer config.PrinterConfig) string {
	if strings.EqualFold(printer.Transport, "network") || strings.EqualFold(printer.Transport, "usb") {
		return "ready"
	}

	return "stub"
}

func printerConnection(printer config.PrinterConfig) map[string]any {
	switch printer.Transport {
	case "network":
		if printer.Network == nil {
			return nil
		}

		return map[string]any{
			"address": printer.Network.Address,
			"timeout": printer.Network.Timeout,
		}
	case "serial":
		if printer.Serial == nil {
			return nil
		}

		return map[string]any{
			"port": printer.Serial.Port,
			"baud": printer.Serial.Baud,
		}
	case "usb":
		if printer.USB == nil {
			return nil
		}

		return map[string]any{
			"queue":      printer.USB.Queue,
			"vendor_id":  printer.USB.VendorID,
			"product_id": printer.USB.ProductID,
		}
	default:
		return nil
	}
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}

	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}

	return cloned
}
