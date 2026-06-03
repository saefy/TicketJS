package main

import (
	"fmt"
	"log"
	"os"

	"thermalik-bridge/internal/svcmgr"
)

func cmdServiceControl(action string, _ []string) {
	var err error

	switch action {
	case "start":
		fmt.Println("Starting thermalik-bridge service...")
		err = svcmgr.Start()
	case "stop":
		fmt.Println("Stopping thermalik-bridge service...")
		err = svcmgr.Stop()
	case "restart":
		fmt.Println("Restarting thermalik-bridge service...")
		err = svcmgr.Restart()
	case "status":
		status, statusErr := svcmgr.Status()
		if statusErr != nil {
			log.Fatalf("status: %v", statusErr)
		}
		fmt.Printf("thermalik-bridge: %s\n", status)
		return
	case "uninstall":
		fmt.Println("Uninstalling thermalik-bridge service...")
		err = svcmgr.Uninstall()
	default:
		fmt.Fprintf(os.Stderr, "unknown service action: %s\n", action)
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("%s: %v", action, err)
	}

	fmt.Printf("thermalik-bridge %s: ok\n", action)
}
