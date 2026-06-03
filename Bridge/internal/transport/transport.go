package transport

import (
	"context"
	"fmt"
	"strings"

	"thermalik-bridge/internal/config"
)

type Driver interface {
	Send(context.Context, config.PrinterConfig, []byte) error
}

type Manager struct {
	drivers map[string]Driver
}

func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]Driver),
	}
}

func NewDefaultManager() *Manager {
	manager := NewManager()
	manager.Register("network", NetworkDriver{})
	manager.Register("serial", StubDriver{Transport: "serial"})
	manager.Register("usb", NewUSBDriver())

	return manager
}

func (m *Manager) Register(name string, driver Driver) {
	m.drivers[strings.ToLower(name)] = driver
}

func (m *Manager) Send(ctx context.Context, printer config.PrinterConfig, payload []byte) error {
	driver, ok := m.drivers[strings.ToLower(printer.Transport)]
	if !ok {
		return fmt.Errorf("transport %q is not registered", printer.Transport)
	}

	return driver.Send(ctx, printer, payload)
}
