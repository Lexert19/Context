package output

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type WordStat struct {
	Word   string  `json:"word"`
	Count  int     `json:"count"`
	Weight float64 `json:"weight"`
}

type FileIndex struct {
	TotalTokens  int        `json:"total_tokens"`
	UniqueTokens int        `json:"unique_tokens"`
	Words        []WordStat `json:"words"`
}

func BuildWordIndex(filePaths []string, targetPath string) error {
	index := make(map[string]FileIndex)

	for _, path := range filePaths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		tokens := tokenize(string(content))
		total := len(tokens)
		if total == 0 {
			continue
		}

		counts := make(map[string]int)
		for _, token := range tokens {
			counts[token]++
		}

		var stats []WordStat
		for word, count := range counts {
			weight := math.Round((float64(count)/float64(total))*10000) / 10000
			stats = append(stats, WordStat{
				Word:   word,
				Count:  count,
				Weight: weight,
			})
		}

		sort.Slice(stats, func(i, j int) bool {
			if stats[i].Count == stats[j].Count {
				return stats[i].Word < stats[j].Word
			}
			return stats[i].Count > stats[j].Count
		})

		normalizedPath := filepath.ToSlash(path)
		index[normalizedPath] = FileIndex{
			TotalTokens:  total,
			UniqueTokens: len(stats),
			Words:        stats,
		}
	}

	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(index)
}

func tokenize(content string) []string {
	var tokens []string

	rawFields := strings.FieldsFunc(content, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for _, field := range rawFields {
		subWords := splitCamelCase(field)
		for _, word := range subWords {
			word = strings.ToLower(word)
			if len(word) > 1 && !isNumber(word) {
				tokens = append(tokens, word)
			}
		}
	}

	return tokens
}

func splitCamelCase(s string) []string {
	var words []string
	var current []rune
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if unicode.IsUpper(r) {
			if len(current) > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) {
					words = append(words, string(current))
					current = nil
				} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					words = append(words, string(current))
					current = nil
				}
			}
		}

		current = append(current, r)
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}

func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
