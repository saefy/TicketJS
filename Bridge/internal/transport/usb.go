package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"thermalik-bridge/internal/config"
)

type commandRunner func(context.Context, string, []string, []byte) ([]byte, error)

type USBDriver struct {
	run commandRunner
}

func NewUSBDriver() USBDriver {
	return USBDriver{
		run: runCommand,
	}
}

func (d USBDriver) Send(ctx context.Context, printer config.PrinterConfig, payload []byte) error {
	if printer.USB == nil || strings.TrimSpace(printer.USB.Queue) == "" {
		return errors.New("usb printer queue is required")
	}

	output, err := d.run(ctx, "lp", []string{"-d", printer.USB.Queue, "-o", "raw"}, payload)
	if err == nil {
		return nil
	}

	if errors.Is(err, exec.ErrNotFound) {
		return errors.New("usb printing requires the lp command to be available on the host system")
	}

	trimmedOutput := strings.TrimSpace(string(output))
	if trimmedOutput != "" {
		return fmt.Errorf("send usb print job: %w: %s", err, trimmedOutput)
	}

	return fmt.Errorf("send usb print job: %w", err)
}

func runCommand(ctx context.Context, name string, args []string, stdin []byte) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = bytes.NewReader(stdin)

	return command.CombinedOutput()
}
