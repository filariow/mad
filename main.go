package main

import (
	"os"

	"github.com/filariow/mad/internal/app"
)

func main() {
	os.Exit(run())
}

func run() int {
	a := app.New()
	return a.Run(os.Args)
}
