package utils

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGetCallerDir_ReturnsDirectoryOfCallerBasedOnSkip(t *testing.T) {
	t.Run("returns test file directory when skip is 0", func(t *testing.T) {
		dir := GetCallerDir(0)

		if dir == "" {
			t.Error("expected non-empty directory")
		}

		if !strings.HasSuffix(dir, "utils") {
			t.Errorf("expected directory to end with 'utils', got '%s'", dir)
		}
	})

	t.Run("always returns an absolute path", func(t *testing.T) {
		dir := GetCallerDir(0)

		if !filepath.IsAbs(dir) {
			t.Errorf("expected absolute path, got '%s'", dir)
		}
	})

	t.Run("returns caller directory when called through helper with skip 1", func(t *testing.T) {
		dir := helperGetCallerDir()

		if !strings.HasSuffix(dir, "utils") {
			t.Errorf("expected directory to end with 'utils', got '%s'", dir)
		}
	})
}

func helperGetCallerDir() string {
	return GetCallerDir(1)
}

func TestGetCallerDir_ReturnsConsistentResultsOnMultipleCalls(t *testing.T) {
	dir1 := GetCallerDir(0)
	dir2 := GetCallerDir(0)

	if dir1 != dir2 {
		t.Errorf("expected consistent results, got '%s' and '%s'", dir1, dir2)
	}
}
