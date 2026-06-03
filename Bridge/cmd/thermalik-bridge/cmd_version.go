package main

import (
	"fmt"

	"thermalik-bridge/internal/version"
)

func cmdVersion() {
	fmt.Println(version.String())
}
