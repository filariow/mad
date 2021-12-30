package main

import (
	"os"

	"github.com/filariow/mad/internal/app"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "com.filariow.mad"

func main() {
	os.Exit(run())
}

func run() int {
	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	a := app.New()
	return a.Run(os.Args)
}
