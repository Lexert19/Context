package source

import (
	"io/fs"
	"path/filepath"
)

type DirSource struct {
	root       string
	extensions []string
}

func NewDirSource(root string, exts []string) *DirSource {
	return &DirSource{root: root, extensions: exts}
}

func (d *DirSource) Header() string { return "" }

func (d *DirSource) Collect() ([]string, error) {
	var files []string
	err := filepath.WalkDir(d.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || (entry.IsDir() && entry.Name() == ".git") {
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.IsDir() && MatchExtension(entry.Name(), d.extensions) {
			files = append(files, filepath.Clean(path))
		}
		return nil
	})
	return files, err
}
