package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetCallerDir returns the directory containing the source file of the caller.
// The skip parameter specifies how many stack frames to skip:
// - skip=0: returns the directory of the file calling GetCallerDir
// - skip=1: returns the directory of the file that called the function calling GetCallerDir
// etc.
func GetCallerDir(skip int) string {
	_, filename, _, ok := runtime.Caller(skip + 1)
	if !ok {
		panic("failed to get caller file path")
	}
	return filepath.Dir(filename)
}

// GetBasePath returns the application base path.
// In Docker/production, this is set via APP_BASE_PATH environment variable.
// In local development, it falls back to deriving the path from runtime.Caller.
func GetBasePath() string {
	if basePath := os.Getenv("APP_BASE_PATH"); basePath != "" {
		return basePath
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller file path")
	}
	// This file is at src/utils/paths.go, so go up two levels to get project root
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}
