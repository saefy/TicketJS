package transport

import (
	"context"
	"errors"
	"net"
	"time"

	"thermalik-bridge/internal/config"
)

type NetworkDriver struct{}

func (NetworkDriver) Send(ctx context.Context, printer config.PrinterConfig, payload []byte) error {
	if printer.Network == nil || printer.Network.Address == "" {
		return errors.New("network printer address is required")
	}

	timeout := 5 * time.Second
	if printer.Network.Timeout != "" {
		parsedTimeout, err := time.ParseDuration(printer.Network.Timeout)
		if err != nil {
			return err
		}
		timeout = parsedTimeout
	}

	dialer := &net.Dialer{Timeout: timeout}
	connection, err := dialer.DialContext(ctx, "tcp", printer.Network.Address)
	if err != nil {
		return err
	}
	defer connection.Close()

	if err := connection.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}

	_, err = connection.Write(payload)

	return err
}
