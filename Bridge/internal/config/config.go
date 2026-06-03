package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultAddress = "127.0.0.1:9100"

type Config struct {
	Server   ServerConfig    `json:"server"`
	Storage  StorageConfig   `json:"storage"`
	Printers []PrinterConfig `json:"printers"`
}

type ServerConfig struct {
	Address        string   `json:"address"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
}

type StorageConfig struct {
	JobStorePath string `json:"job_store_path"`
}

type PrinterConfig struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Transport    string                `json:"transport"`
	Default      bool                  `json:"default"`
	Capabilities map[string]any        `json:"capabilities,omitempty"`
	Network      *NetworkPrinterConfig `json:"network,omitempty"`
	Serial       *SerialPrinterConfig  `json:"serial,omitempty"`
	USB          *USBPrinterConfig     `json:"usb,omitempty"`
}

type NetworkPrinterConfig struct {
	Address string `json:"address"`
	Timeout string `json:"timeout,omitempty"`
}

type SerialPrinterConfig struct {
	Port string `json:"port"`
	Baud int    `json:"baud,omitempty"`
}

type USBPrinterConfig struct {
	Queue     string `json:"queue,omitempty"`
	VendorID  string `json:"vendor_id"`
	ProductID string `json:"product_id"`
}

func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "thermalik-bridge", "config.json"), nil
}

func DefaultDataDir() (string, error) {
	dataDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dataDir, "thermalik-bridge"), nil
}

func DefaultConfig() (Config, error) {
	dataDir, err := DefaultDataDir()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Server: ServerConfig{
			Address: DefaultAddress,
		},
		Storage: StorageConfig{
			JobStorePath: filepath.Join(dataDir, "jobs.json"),
		},
		Printers: []PrinterConfig{},
	}, nil
}

func Load(path string) (Config, error) {
	cfg, err := DefaultConfig()
	if err != nil {
		return Config{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}

		return Config{}, err
	}

	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode %s: %w", path, err)
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Address == "" {
		c.Server.Address = DefaultAddress
	}
}

func (c Config) Validate() error {
	if c.Server.Address == "" {
		return errors.New("server.address is required")
	}

	if c.Storage.JobStorePath == "" {
		return errors.New("storage.job_store_path is required")
	}

	seenIDs := make(map[string]struct{}, len(c.Printers))
	defaultCount := 0

	for _, printer := range c.Printers {
		if printer.ID == "" {
			return errors.New("printers[].id is required")
		}

		if _, exists := seenIDs[printer.ID]; exists {
			return fmt.Errorf("duplicate printer id %q", printer.ID)
		}

		seenIDs[printer.ID] = struct{}{}

		if printer.Name == "" {
			return fmt.Errorf("printer %q must define name", printer.ID)
		}

		switch printer.Transport {
		case "network":
			if printer.Network == nil || printer.Network.Address == "" {
				return fmt.Errorf("printer %q must define network.address", printer.ID)
			}
		case "serial":
		case "usb":
			if printer.USB == nil || printer.USB.Queue == "" {
				return fmt.Errorf("printer %q must define usb.queue", printer.ID)
			}
		default:
			return fmt.Errorf("printer %q uses unsupported transport %q", printer.ID, printer.Transport)
		}

		if printer.Default {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		return errors.New("only one printer may be marked as default")
	}

	return nil
}
