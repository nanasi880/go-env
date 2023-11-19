package config

import (
	"os"
	"runtime"
)

var supportedOS = map[string]bool{
	"darwin": true,
}
var defaultLocationByOS = map[string]string{
	"darwin": "/usr/local/go",
}

const (
	BinaryDirectoryName = "bin"
	BinaryName          = "goenv"
	GoRootDirectoryName = "goroot"
)

func LoadLocation() string {
	mustSupportedOS()
	location := os.Getenv("GOENV_LOCATION")
	if location != "" {
		return location
	}
	return defaultLocationByOS[runtime.GOOS]
}

func mustSupportedOS() {
	if supportedOS[runtime.GOOS] == false {
		panic("not supported OS: " + runtime.GOOS)
	}
}
