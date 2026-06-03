package transport

import (
	"context"
	"fmt"

	"thermalik-bridge/internal/config"
)

type StubDriver struct {
	Transport string
}

func (d StubDriver) Send(_ context.Context, _ config.PrinterConfig, _ []byte) error {
	return fmt.Errorf("%s transport is not implemented yet", d.Transport)
}
