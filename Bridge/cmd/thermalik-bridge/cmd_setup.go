package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"thermalik-bridge/internal/config"
	"thermalik-bridge/internal/svcmgr"
	"thermalik-bridge/internal/version"
)

func cmdSetup(_ []string) {
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		log.Fatalf("resolve config path: %v", err)
	}

	fmt.Printf("\n✦ Thermalik Bridge Setup %s\n", version.Version)
	fmt.Println(strings.Repeat("─", 48))
	fmt.Printf("Config: %s\n\n", cfgPath)

	existing, loadErr := config.Load(cfgPath)
	cfg := existing

	if loadErr == nil && len(existing.Printers) > 0 {
		fmt.Printf("Existing config found with %d printer(s).\n", len(existing.Printers))
		if confirm("Add another printer?") {
			printer := promptPrinter()
			cfg.Printers = append(cfg.Printers, printer)
		}
	} else {
		printer := promptPrinter()
		cfg.Printers = append(cfg.Printers, printer)
	}

	origin := promptAllowedOrigin()
	if origin != "" {
		cfg.Server.AllowedOrigins = appendUniqueOrigin(cfg.Server.AllowedOrigins, origin)
	}

	if err := saveConfig(cfgPath, cfg); err != nil {
		log.Fatalf("save config: %v", err)
	}

	fmt.Printf("✓ Config saved to %s\n", cfgPath)

	installAndStart(cfgPath)

	fmt.Println("\nTest with:")
	fmt.Println("  curl http://127.0.0.1:9100/health")
	fmt.Println("  curl http://127.0.0.1:9100/printers")
}

func promptPrinter() config.PrinterConfig {
	scanner := bufio.NewScanner(os.Stdin)

	id := prompt(scanner, "Printer ID (e.g. caja-1)", "caja-principal")
	name := prompt(scanner, "Printer display name (e.g. Caja Principal)", "Caja Principal")
	transport := promptChoice(scanner, "Transport", []string{"network", "usb", "serial"}, "network")

	printer := config.PrinterConfig{
		ID:        id,
		Name:      name,
		Transport: transport,
		Default:   true,
		Capabilities: map[string]any{
			"paper_width_mm": 58,
			"encodings":      []string{"cp437", "cp850"},
			"job_tracking":   true,
		},
	}

	switch transport {
	case "network":
		address := prompt(scanner, "IP address:port (e.g. 192.168.1.50:9100)", "192.168.1.50:9100")
		timeout := prompt(scanner, "Connection timeout", "5s")
		printer.Network = &config.NetworkPrinterConfig{
			Address: address,
			Timeout: timeout,
		}
	case "usb":
		queue := prompt(scanner, "Print queue name (from `lpstat -p`)", "")
		printer.USB = &config.USBPrinterConfig{
			Queue: queue,
		}
	case "serial":
		port := prompt(scanner, "Serial port (e.g. /dev/ttyUSB0 or COM3)", "")
		printer.Serial = &config.SerialPrinterConfig{
			Port: port,
			Baud: 9600,
		}
	}

	widthStr := promptChoice(scanner, "Paper width (mm)", []string{"58", "80"}, "58")
	if widthStr == "80" {
		printer.Capabilities["paper_width_mm"] = 80
	}

	return printer
}

func installAndStart(cfgPath string) {
	// Install the service if not already registered.
	if err := svcmgr.Install(); err != nil {
		// Service may already be installed — attempt to restart instead.
		fmt.Printf("Note: install returned: %v\n", err)
	} else {
		fmt.Println("✓ Service installed")
	}

	if err := svcmgr.Start(); err != nil {
		fmt.Printf("Warning: service start returned: %v\n", err)
		fmt.Printf("You can start manually with: thermalik-bridge run -config %s\n", cfgPath)
		return
	}

	fmt.Println("✓ Service started — listening on http://127.0.0.1:9100")
}

func saveConfig(path string, cfg config.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func prompt(scanner *bufio.Scanner, label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("? %s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("? %s: ", label)
	}

	scanner.Scan()
	value := strings.TrimSpace(scanner.Text())

	if value == "" {
		return defaultValue
	}

	return value
}

func promptChoice(scanner *bufio.Scanner, label string, choices []string, defaultValue string) string {
	fmt.Printf("? %s [%s] (%s): ", label, defaultValue, strings.Join(choices, "/"))
	scanner.Scan()
	value := strings.TrimSpace(scanner.Text())

	if value == "" {
		return defaultValue
	}

	for _, c := range choices {
		if strings.EqualFold(value, c) {
			return c
		}
	}

	fmt.Printf("Invalid choice %q, using default %q\n", value, defaultValue)

	return defaultValue
}

func promptAllowedOrigin() string {
	scanner := bufio.NewScanner(os.Stdin)
	return prompt(scanner, "Allowed origin to add (e.g. https://my-web-app.com, Enter to skip)", "")
}

func appendUniqueOrigin(origins []string, origin string) []string {
	originLower := strings.ToLower(origin)

	for _, current := range origins {
		if strings.ToLower(strings.TrimSpace(current)) == originLower {
			return origins
		}
	}

	return append(origins, origin)
}

func confirm(label string) bool {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("? %s [y/N]: ", label)
	scanner.Scan()
	value := strings.TrimSpace(strings.ToLower(scanner.Text()))

	return value == "y" || value == "yes"
}
