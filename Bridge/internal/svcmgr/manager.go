// Package svcmgr wraps github.com/kardianos/service to manage thermalik-bridge
// as a system service (LaunchAgent on macOS, systemd on Linux, Windows Service on Windows).
package svcmgr

import (
	"fmt"

	"github.com/kardianos/service"
)

const (
	serviceName        = "thermalik-bridge"
	serviceDisplayName = "Thermalik Bridge"
	serviceDescription = "Local companion service for thermal printer communication."
)

// program satisfies the service.Interface required by kardianos/service.
// The actual bridge logic is started externally via the `run` subcommand,
// which the service manager invokes as the service executable.
type program struct{}

func (p *program) Start(_ service.Service) error { return nil }
func (p *program) Stop(_ service.Service) error  { return nil }

func newService() (service.Service, error) {
	cfg := &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
		// The service runs the same binary with the `run` subcommand.
		Arguments: []string{"run"},
		Option: service.KeyValue{
			// macOS: run as a user LaunchAgent, not a system daemon.
			"UserService": true,
		},
	}

	return service.New(&program{}, cfg)
}

// Install registers thermalik-bridge as a system service.
func Install() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	if err := svc.Install(); err != nil {
		return fmt.Errorf("install service: %w", err)
	}

	return nil
}

// Uninstall removes the system service registration.
func Uninstall() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	if err := svc.Uninstall(); err != nil {
		return fmt.Errorf("uninstall service: %w", err)
	}

	return nil
}

// Start starts the installed system service.
func Start() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	if err := svc.Start(); err != nil {
		return fmt.Errorf("start service: %w", err)
	}

	return nil
}

// Stop stops the running system service.
func Stop() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	if err := svc.Stop(); err != nil {
		return fmt.Errorf("stop service: %w", err)
	}

	return nil
}

// Restart stops and starts the system service.
func Restart() error {
	svc, err := newService()
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	if err := svc.Restart(); err != nil {
		return fmt.Errorf("restart service: %w", err)
	}

	return nil
}

// Status returns a human-readable status string for the service.
func Status() (string, error) {
	svc, err := newService()
	if err != nil {
		return "", fmt.Errorf("create service: %w", err)
	}

	status, err := svc.Status()
	if err != nil {
		return "", fmt.Errorf("query status: %w", err)
	}

	switch status {
	case service.StatusRunning:
		return "running", nil
	case service.StatusStopped:
		return "stopped", nil
	default:
		return "unknown", nil
	}
}
