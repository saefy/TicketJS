package main

import (
	"fmt"
	"log"

	"thermalik-bridge/internal/svcmgr"
	"thermalik-bridge/internal/updater"
)

func cmdUpdate(_ []string) {
	fmt.Println("Checking for updates...")

	info, err := updater.CheckLatest()
	if err != nil {
		log.Fatalf("check for updates: %v", err)
	}

	if !info.Available {
		fmt.Println("thermalik-bridge is already up to date.")
		return
	}

	fmt.Printf("Update available: %s\n", info.Version)
	fmt.Println("Downloading and verifying...")

	if err := updater.Apply(info); err != nil {
		log.Fatalf("apply update: %v", err)
	}

	fmt.Printf("Updated to %s. Restarting service...\n", info.Version)

	if err := svcmgr.Restart(); err != nil {
		// Non-fatal: binary was replaced but service restart failed.
		// The user can restart manually.
		fmt.Printf("Warning: service restart failed (%v). Run 'thermalik-bridge restart' manually.\n", err)
		return
	}

	fmt.Println("Update complete. thermalik-bridge is running the new version.")
}
