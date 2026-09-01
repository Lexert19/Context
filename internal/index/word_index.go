package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func Tokenize(content string) []string {
	var tokens []string
	f := strings.FieldsFunc(content, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
	for _, t := range f {
		t = strings.ToLower(t)
		if len(t) > 1 && !isNumeric(t) {
			tokens = append(tokens, t)
		}
	}
	return tokens
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func BuildWordIndex(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tokens := Tokenize(string(content))
	set := make(map[string]bool)
	for _, t := range tokens {
		set[t] = true
	}
	words := make([]string, 0, len(set))
	for w := range set {
		words = append(words, w)
	}
	return words, nil
}

func SaveWordIndex(filePaths []string) error {
	wordSet := make(map[string]bool)
	for _, path := range filePaths {
		words, err := BuildWordIndex(path)
		if err != nil {
			continue
		}
		for _, w := range words {
			wordSet[w] = true
		}
	}

	wordsList := make([]string, 0, len(wordSet))
	for w := range wordSet {
		wordsList = append(wordsList, w)
	}
	sort.Strings(wordsList)

	outputPath := filepath.Join(".idea", "word_index.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for word index: %w", err)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create word index file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(wordsList); err != nil {
		return fmt.Errorf("failed to write word index JSON: %w", err)
	}

	fmt.Printf("Saved word index (%d unique words) to %s\n", len(wordsList), outputPath)
	return nil
}
