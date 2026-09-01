package review

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Append(w *bufio.Writer, target, reviewDir string) error {
	content, err := resolveContent(target, reviewDir)
	if err != nil {
		return fmt.Errorf("failed to read review: %w", err)
	}

	if content == "" {
		return nil
	}

	w.WriteString("========== PREVIOUS REVIEWS ==========\n\n")
	w.WriteString(content)
	w.WriteString("\n=====================================\n\n")
	return nil
}

func resolveContent(target, reviewDir string) (string, error) {
	if target == "" {
		return readAllFromDir(reviewDir)
	}

	if strings.Contains(target, "/") || strings.Contains(target, "\\") {
		return readFile(target)
	}

	return readFile(filepath.Join(reviewDir, target))
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readAllFromDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var sb strings.Builder
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("=== REVIEW: %s ===\n", entry.Name()))
		sb.Write(data)
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}
