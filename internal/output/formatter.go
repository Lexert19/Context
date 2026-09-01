package output

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Formatter interface {
	Write(out *bufio.Writer, path string) error
}

func ResolveFormatter(listOnly, pathHeader bool) Formatter {
	if listOnly {
		return &ListFormatter{}
	}
	return &ContentFormatter{PathHeader: pathHeader}
}

type ListFormatter struct{}

func (l *ListFormatter) Write(out *bufio.Writer, path string) error {
	_, err := out.WriteString(filepath.ToSlash(path) + "\n")
	return err
}

type ContentFormatter struct {
	PathHeader bool
}

func (c *ContentFormatter) Write(out *bufio.Writer, path string) error {
	normalized := filepath.ToSlash(path)
	out.WriteString(fmt.Sprintf("--- FILE: \"%s\" ---\n", normalized))
	if c.PathHeader {
		out.WriteString(fmt.Sprintf("\"%s\"\n", normalized))
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(out, f); err != nil {
		return err
	}
	out.WriteString("\n\n")
	return nil
}
