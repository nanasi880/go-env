package main

import (
	"os"

	"github.com/nanasi880/go-env/internal/cmd/gocmd"
	"github.com/nanasi880/go-env/internal/cmd/goenv"
)

func main() {
	switch os.Args[0] {
	case "go", "gofmt":
		gocmd.Main()
	default:
		goenv.Main()
	}
}
