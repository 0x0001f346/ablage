package main

import (
	"fmt"
	"os"

	"git.0x0001f346.de/andreas/ablage/app"
	"git.0x0001f346.de/andreas/ablage/config"
	"git.0x0001f346.de/andreas/ablage/filesystem"
)

func main() {
	err := config.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] %v\n", err)
		os.Exit(1)
	}

	err = filesystem.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] %v\n", err)
		os.Exit(1)
	}

	err = app.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] %v\n", err)
		os.Exit(1)
	}
}
