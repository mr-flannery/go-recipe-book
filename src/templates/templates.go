package templates

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/utils"
)

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

var Templates = template.Must(loadTemplatesRecursive(utils.GetCallerDir(0)))
