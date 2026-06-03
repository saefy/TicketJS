package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"thermalik-bridge/internal/bridge"
	"thermalik-bridge/internal/domain"
	"thermalik-bridge/internal/store"
)

type BridgeService interface {
	Health() map[string]any
	ListPrinters() []domain.Printer
	EnqueuePrint(context.Context, string, []byte, map[string]any) (*domain.Job, error)
	GetJob(context.Context, string) (*domain.Job, error)
}

type printRequest struct {
	Printer string         `json:"printer"`
	Bytes   []int          `json:"bytes"`
	Render  map[string]any `json:"render,omitempty"`
}

func NewHandler(service BridgeService, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeMethodNotAllowed(writer, http.MethodGet)
			return
		}

		writeJSON(writer, http.StatusOK, service.Health())
	})
	mux.HandleFunc("/printers", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeMethodNotAllowed(writer, http.MethodGet)
			return
		}

		writeJSON(writer, http.StatusOK, map[string]any{
			"data": service.ListPrinters(),
		})
	})
	mux.HandleFunc("/print", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writeMethodNotAllowed(writer, http.MethodPost)
			return
		}

		var payload printRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{
				"message": "Invalid JSON body.",
			})
			return
		}

		if payload.Printer == "" {
			writeJSON(writer, http.StatusBadRequest, map[string]string{
				"message": "The printer field is required.",
			})
			return
		}

		if len(payload.Bytes) == 0 {
			writeJSON(writer, http.StatusBadRequest, map[string]string{
				"message": "The bytes field must contain at least one byte.",
			})
			return
		}

		bytes, err := normalizeBytes(payload.Bytes)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
			return
		}

		job, err := service.EnqueuePrint(request.Context(), payload.Printer, bytes, payload.Render)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if errors.Is(err, bridge.ErrPrinterNotFound) {
				statusCode = http.StatusNotFound
			}

			writeJSON(writer, statusCode, map[string]string{
				"message": err.Error(),
			})
			return
		}

		writeJSON(writer, http.StatusAccepted, presentJob(job))
	})
	mux.HandleFunc("/jobs/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writeMethodNotAllowed(writer, http.MethodGet)
			return
		}

		jobID := strings.TrimPrefix(request.URL.Path, "/jobs/")
		if jobID == "" {
			writeJSON(writer, http.StatusBadRequest, map[string]string{
				"message": "The job ID is required.",
			})
			return
		}

		job, err := service.GetJob(request.Context(), jobID)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if errors.Is(err, store.ErrJobNotFound) {
				statusCode = http.StatusNotFound
			}

			writeJSON(writer, statusCode, map[string]string{
				"message": err.Error(),
			})
			return
		}

		writeJSON(writer, http.StatusOK, presentJob(job))
	})

	return withCORS(mux, allowedOrigins)
}

// withCORS restricts cross-origin access to loopback origins and explicitly
// allowed origins. It also handles Chrome's Private Network Access (PNA)
// preflight so that requests from public HTTPS pages to 127.0.0.1 are not
// blocked by the browser.
func withCORS(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.ToLower(o)] = struct{}{}
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")

		if origin != "" {
			originLower := strings.ToLower(origin)
			_, isAllowed := allowed[originLower]
			if !isAllowed && !isLoopbackOrigin(origin) {
				http.Error(writer, "Forbidden", http.StatusForbidden)
				return
			}

			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

			// Chrome Private Network Access: respond to the PNA preflight header
			// so requests from public HTTPS origins to 127.0.0.1 are permitted.
			if request.Header.Get("Access-Control-Request-Private-Network") == "true" {
				writer.Header().Set("Access-Control-Allow-Private-Network", "true")
			}

			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return
			}
		}

		next.ServeHTTP(writer, request)
	})
}

// isLoopbackOrigin reports whether origin is a loopback address (localhost or 127.x.x.x).
func isLoopbackOrigin(origin string) bool {
	// Strip scheme.
	host := origin
	if idx := strings.Index(host, "://"); idx != -1 {
		host = host[idx+3:]
	}

	// Strip port.
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	return host == "localhost" || strings.HasPrefix(host, "127.")
}

func writeMethodNotAllowed(writer http.ResponseWriter, allowedMethod string) {
	writer.Header().Set("Allow", allowedMethod)
	writeJSON(writer, http.StatusMethodNotAllowed, map[string]string{
		"message": "Method not allowed.",
	})
}

func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	_ = json.NewEncoder(writer).Encode(payload)
}

func normalizeBytes(values []int) ([]byte, error) {
	bytes := make([]byte, len(values))

	for index, value := range values {
		if value < 0 || value > 255 {
			return nil, errors.New("Each byte must be between 0 and 255.")
		}

		bytes[index] = byte(value)
	}

	return bytes, nil
}

func presentJob(job *domain.Job) map[string]any {
	response := map[string]any{
		"job_id":       job.ID,
		"status":       job.Status,
		"printer":      job.Printer,
		"created_at":   job.CreatedAt,
		"updated_at":   job.UpdatedAt,
		"completed_at": job.CompletedAt,
	}

	if job.Error != "" {
		response["error"] = job.Error
	}

	if job.Render != nil {
		response["render"] = job.Render
	}

	return response
}
