package main

import (
	"git.0x0001f346.de/andreas/ablage/app"
	"git.0x0001f346.de/andreas/ablage/config"
	"git.0x0001f346.de/andreas/ablage/filesystem"
)

func main() {
	config.Init()
	filesystem.Init()
	app.Init()
}
