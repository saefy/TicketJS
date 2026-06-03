package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"thermalik-bridge/internal/bridge"
	"thermalik-bridge/internal/config"
	"thermalik-bridge/internal/store"
	"thermalik-bridge/internal/transport"
)

type fakeDriver struct{}

func (fakeDriver) Send(_ context.Context, _ config.PrinterConfig, _ []byte) error {
	return nil
}

func TestPrintersEndpointReturnsConfiguredPrinters(t *testing.T) {
	t.Parallel()

	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodGet, "/printers", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 printer, got %d", len(payload.Data))
	}
}

func TestPrintLifecycleQueuesAndCompletesJob(t *testing.T) {
	t.Parallel()

	handler := testHandler(t)
	request := httptest.NewRequest(http.MethodPost, "/print", strings.NewReader(`{"printer":"caja-principal","bytes":[27,64,10]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	jobID, _ := created["job_id"].(string)
	if jobID == "" {
		t.Fatal("expected job_id in response")
	}

	var body map[string]any
	for range 25 {
		jobRequest := httptest.NewRequest(http.MethodGet, "/jobs/"+jobID, nil)
		jobRecorder := httptest.NewRecorder()
		handler.ServeHTTP(jobRecorder, jobRequest)

		if jobRecorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", jobRecorder.Code, jobRecorder.Body.String())
		}

		if err := json.Unmarshal(jobRecorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("Unmarshal returned error: %v", err)
		}

		if body["status"] == "completed" {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("expected completed status, got %v", body["status"])
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	dir := t.TempDir()
	jobStore, err := store.NewFileJobStore(filepath.Join(dir, "jobs.json"))
	if err != nil {
		t.Fatalf("NewFileJobStore returned error: %v", err)
	}

	manager := transport.NewManager()
	manager.Register("fake", fakeDriver{})

	service, err := bridge.NewService(ctx, config.Config{
		Server: config.ServerConfig{
			Address: "127.0.0.1:9100",
		},
		Storage: config.StorageConfig{
			JobStorePath: filepath.Join(dir, "jobs.json"),
		},
		Printers: []config.PrinterConfig{
			{
				ID:        "caja-principal",
				Name:      "Caja principal",
				Transport: "fake",
				Default:   true,
				Capabilities: map[string]any{
					"encodings": []string{"cp437"},
				},
				Network: nil,
			},
		},
	}, jobStore, manager)
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	return NewHandler(service, nil)
}
