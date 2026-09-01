package source

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type ListSource struct {
	path       string
	extensions []string
}

func NewListSource(path string, exts []string) *ListSource {
	return &ListSource{path: path, extensions: exts}
}

func (l *ListSource) Header() string { return "" }

func (l *ListSource) Collect() ([]string, error) {
	f, err := os.Open(l.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var files []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.Trim(strings.TrimSpace(s.Text()), `"'`+"`")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		clean := filepath.Clean(line)
		if MatchExtension(filepath.Base(clean), l.extensions) {
			files = append(files, clean)
		}
	}
	return files, s.Err()
}
