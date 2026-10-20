package output

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
)

type WordStat struct {
	Word   string  `json:"word"`
	Count  int     `json:"count"`
	Weight float64 `json:"weight"`
}

type FileIndexJSON struct {
	TotalTokens  int        `json:"total_tokens"`
	UniqueTokens int        `json:"unique_tokens"`
	Words        []WordStat `json:"words"`
}

func BuildWordIndexProto(filePaths []string, targetPath string) error {
	files := make([]string, 0, len(filePaths))
	fileId := make(map[string]int)

	idx := make(map[string]*PostingList)

	for _, path := range filePaths {
		content, _ := os.ReadFile(path)
		tokens := tokenize(string(content))
		if len(tokens) == 0 {
			continue
		}

		p := filepath.ToSlash(path)
		if _, ok := fileId[p]; !ok {
			fileId[p] = len(files)
			files = append(files, p)
		}
		id := fileId[p]

		counts := map[string]int{}
		for _, t := range tokens {
			counts[t]++
		}

		for w, c := range counts {
			pl, ok := idx[w]
			if !ok {
				pl = &PostingList{}
				idx[w] = pl
			}
			pl.Postings = append(pl.Postings, &Posting{
				FileId: uint32(id),
				Count:  uint32(c),
			})
		}
	}

	out := &InvertedIndex{
		Files: files,
		Index: idx,
	}

	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(out)
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
	return os.WriteFile(targetPath, data, 0644)
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
