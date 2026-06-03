package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesDefaultsWhenConfigMissing(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Address != DefaultAddress {
		t.Fatalf("expected default address %q, got %q", DefaultAddress, cfg.Server.Address)
	}

	if cfg.Storage.JobStorePath == "" {
		t.Fatal("expected default job store path")
	}
}

func TestLoadValidatesDuplicatePrinterIDs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
  "server": { "address": "127.0.0.1:9100" },
  "storage": { "job_store_path": "` + filepath.ToSlash(filepath.Join(dir, "jobs.json")) + `" },
  "printers": [
    { "id": "main", "name": "Main", "transport": "network", "network": { "address": "127.0.0.1:9100" } },
    { "id": "main", "name": "Backup", "transport": "network", "network": { "address": "127.0.0.1:9101" } }
  ]
}`)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected duplicate printer validation error")
	}
}

func TestLoadRequiresUSBQueueForUSBPrinters(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := []byte(`{
  "server": { "address": "127.0.0.1:9100" },
  "storage": { "job_store_path": "` + filepath.ToSlash(filepath.Join(dir, "jobs.json")) + `" },
  "printers": [
    { "id": "main", "name": "Main", "transport": "usb", "usb": { "vendor_id": "04b8", "product_id": "0e15" } }
  ]
}`)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected usb queue validation error")
	}
}
