package transport

import (
	"context"
	"errors"
	"strings"
	"testing"

	"thermalik-bridge/internal/config"
)

func TestUSBDriverSendUsesRawLPQueue(t *testing.T) {
	t.Parallel()

	var gotName string
	var gotArgs []string
	var gotStdin []byte

	driver := USBDriver{
		run: func(_ context.Context, name string, args []string, stdin []byte) ([]byte, error) {
			gotName = name
			gotArgs = append([]string(nil), args...)
			gotStdin = append([]byte(nil), stdin...)

			return nil, nil
		},
	}

	payload := []byte{0x1b, 0x40, 0x0a}
	err := driver.Send(context.Background(), config.PrinterConfig{
		ID:        "usb-main",
		Name:      "USB Main",
		Transport: "usb",
		USB: &config.USBPrinterConfig{
			Queue: "EPSON_TM_T20III",
		},
	}, payload)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotName != "lp" {
		t.Fatalf("expected command lp, got %q", gotName)
	}

	expectedArgs := []string{"-d", "EPSON_TM_T20III", "-o", "raw"}
	if strings.Join(gotArgs, " ") != strings.Join(expectedArgs, " ") {
		t.Fatalf("expected args %v, got %v", expectedArgs, gotArgs)
	}

	if string(gotStdin) != string(payload) {
		t.Fatalf("expected payload %v, got %v", payload, gotStdin)
	}
}

func TestUSBDriverSendRequiresQueue(t *testing.T) {
	t.Parallel()

	driver := NewUSBDriver()

	err := driver.Send(context.Background(), config.PrinterConfig{
		ID:        "usb-main",
		Name:      "USB Main",
		Transport: "usb",
		USB:       &config.USBPrinterConfig{},
	}, []byte{0x1b, 0x40})
	if err == nil {
		t.Fatal("expected missing queue error")
	}
}

func TestUSBDriverSendSurfacesCommandOutput(t *testing.T) {
	t.Parallel()

	driver := USBDriver{
		run: func(_ context.Context, _ string, _ []string, _ []byte) ([]byte, error) {
			return []byte("lp: Unknown destination"), errors.New("exit status 1")
		},
	}

	err := driver.Send(context.Background(), config.PrinterConfig{
		ID:        "usb-main",
		Name:      "USB Main",
		Transport: "usb",
		USB: &config.USBPrinterConfig{
			Queue: "missing-printer",
		},
	}, []byte{0x1b, 0x40})
	if err == nil {
		t.Fatal("expected send error")
	}

	if !strings.Contains(err.Error(), "Unknown destination") {
		t.Fatalf("expected lp output in error, got %v", err)
	}
}
