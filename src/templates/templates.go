package templates

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
)

// getPackageDir returns the directory containing this source file
func getPackageDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	return filepath.Dir(filename)
}

func loadTemplatesRecursive(root string) (*template.Template, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return template.ParseFiles(files...)
}

var Templates = template.Must(loadTemplatesRecursive(getPackageDir()))
