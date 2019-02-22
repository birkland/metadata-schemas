package jsonschema

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Load scans the provided list of files and directories
// recursively and produces a Map of schema id to parsed instance
// documents
func Load(filepaths []string) (Map, error) {
	var m Map = make(map[string]Instance)
	for _, path := range filepaths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not stat file %s", path)
		}

		err = load(m, path, info)
		if err != nil {
			return nil, errors.Wrapf(err, "error loading from %s", path)
		}
	}

	return m, nil
}

func load(m Map, path string, info os.FileInfo) error {
	if info.IsDir() {
		return loadDir(m, path)
	}

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "could not open file at %s", path)
	}
	defer file.Close()

	return m.Add(file)
}

func loadDir(m Map, path string) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if p != path {
			return load(m, p, info)
		}
		return nil
	})
}
