package postgres

import (
	"os"
	"strings"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func isShortMode() bool {
	for _, arg := range os.Args {
		if arg == "-short" || arg == "-test.short" ||
			strings.HasPrefix(arg, "-short=") || strings.HasPrefix(arg, "-test.short=") {
			return true
		}
	}
	return false
}

func TestMain(m *testing.M) {
	if isShortMode() {
		os.Exit(m.Run())
	}

	testutil.SetupSharedTestDatabase()
	code := m.Run()
	testutil.TeardownSharedTestDatabase()
	os.Exit(code)
}
