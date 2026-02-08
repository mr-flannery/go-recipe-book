package templates

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

var funcMap = template.FuncMap{
	"dict": func(values ...any) map[string]any {
		if len(values)%2 != 0 {
			panic("dict requires even number of arguments")
		}
		m := make(map[string]any, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				panic("dict keys must be strings")
			}
			m[key] = values[i+1]
		}
		return m
	},
	"joinTagNames": func(tags []models.Tag) string {
		names := make([]string, len(tags))
		for i, t := range tags {
			names[i] = t.Name
		}
		return strings.Join(names, ",")
	},
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

	return template.New("").Funcs(funcMap).ParseFiles(files...)
}

var Templates = template.Must(loadTemplatesRecursive(utils.GetCallerDir(0)))
