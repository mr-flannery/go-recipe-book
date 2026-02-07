package utils

import (
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
