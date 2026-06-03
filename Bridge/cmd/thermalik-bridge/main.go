package main

import (
	"fmt"
	"os"
)

const usage = `Usage: thermalik-bridge <command> [flags]

Commands:
  setup      Interactive wizard: configure printers and install as a system service
  run        Run the bridge in the foreground (used internally by the service manager)
  start      Start the installed system service
  stop       Stop the installed system service
  restart    Restart the installed system service
  status     Show whether the system service is running
  update     Check for updates and apply if available
  uninstall  Remove the system service
  version    Print version information

Run 'thermalik-bridge <command> -help' for command-specific flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "run":
		cmdRun(args)
	case "setup":
		cmdSetup(args)
	case "start":
		cmdServiceControl("start", args)
	case "stop":
		cmdServiceControl("stop", args)
	case "restart":
		cmdServiceControl("restart", args)
	case "status":
		cmdServiceControl("status", args)
	case "uninstall":
		cmdServiceControl("uninstall", args)
	case "update":
		cmdUpdate(args)
	case "version", "-version", "--version", "-v":
		cmdVersion()
	case "-help", "--help", "-h", "help":
		fmt.Fprint(os.Stderr, usage)
	default:
		fmt.Fprintf(os.Stderr, "thermalik-bridge: unknown command %q\n\n%s", cmd, usage)
		os.Exit(1)
	}
}
