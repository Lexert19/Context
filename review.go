package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type appendReviewFlag struct {
	value string
	set   bool
}

func (a *appendReviewFlag) String() string {
	return a.value
}

func (a *appendReviewFlag) Set(s string) error {
	a.value = s
	a.set = true
	return nil
}

func getAllReviews(reviewDir string) (string, error) {
	entries, err := os.ReadDir(reviewDir)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".txt") {
			continue
		}
		path := filepath.Join(reviewDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		result.WriteString(fmt.Sprintf("=== REVIEW: %s ===\n", name))
		result.Write(content)
		result.WriteString("\n\n")
	}
	return result.String(), nil
}

func getReviewFile(reviewDir, filename string) (string, error) {
	fullPath := filepath.Join(reviewDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func readReviewFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func appendReviewIfNeeded(w *bufio.Writer, appendReviewFlag appendReviewFlag, reviewDir string) error {
	if !appendReviewFlag.set {
		return nil
	}

	var reviewText string
	var err error
	value := appendReviewFlag.value

	if value == "" {
		reviewText, err = getAllReviews(reviewDir)
	} else if strings.Contains(value, "/") || strings.Contains(value, "\\") {
		reviewText, err = readReviewFile(value)
	} else {
		reviewText, err = getReviewFile(reviewDir, value)
	}

	if err != nil {
		return fmt.Errorf("could not read review: %w", err)
	}

	if reviewText != "" {
		w.WriteString("========== PREVIOUS REVIEWS ==========\n\n")
		w.WriteString(reviewText)
		w.WriteString("\n=====================================\n\n")
	}
	return nil
}
