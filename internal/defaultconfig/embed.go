package defaultconfig

import (
	"bytes"
	_ "embed"
)

// Config contains the embedded default .plumber.yaml configuration.
// The file is copied here by the build process (make build).
// Source of truth: .plumber.yaml in the repository root.
//
//go:embed default.yaml
var Config []byte

// buildHeader is the comment added by 'make build' that should be stripped for user output
var buildHeader = []byte("# DO NOT EDIT - Generated from .plumber.yaml by 'make build'\n")

// Get returns the embedded default configuration as a byte slice,
// with the build-time header comment stripped for clean user output.
func Get() []byte {
	// Strip the build header if present
	if bytes.HasPrefix(Config, buildHeader) {
		return Config[len(buildHeader):]
	}
	return Config
}

// GetString returns the embedded default configuration as a string,
// with the build-time header comment stripped for clean user output.
func GetString() string {
	return string(Get())
}

// GetRaw returns the raw embedded configuration including the build header.
// Use this only for debugging or internal purposes.
func GetRaw() []byte {
	return Config
}
