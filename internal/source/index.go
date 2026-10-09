package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"

	"mycontext/internal/output"
)

type IndexSource struct {
	words      []string
	indexPath  string
	extensions []string
}

func NewIndexSource(words []string, indexPath string, exts []string) *IndexSource {
	return &IndexSource{words: words, indexPath: indexPath, extensions: exts}
}

func (s *IndexSource) Header() string {
	return fmt.Sprintf("========== FILES MATCHING INDEX: %s ==========\n\n",
		strings.Join(s.words, ", "))
}

func (s *IndexSource) Collect() ([]string, error) {
	data, err := os.ReadFile(s.indexPath)
	if err != nil {
		return nil, fmt.Errorf("read index %q: %w", s.indexPath, err)
	}

	idx := &output.InvertedIndex{}
	if err := proto.Unmarshal(data, idx); err != nil {
		return nil, fmt.Errorf("parse index %q: %w", s.indexPath, err)
	}

	score := make(map[uint32]int)
	matched := make(map[uint32]map[string]bool)

	for _, w := range s.words {
		w = strings.ToLower(strings.TrimSpace(w))
		if w == "" {
			continue
		}
		pl, ok := idx.Index[w]
		if !ok {
			continue
		}
		for _, p := range pl.Postings {
			score[p.FileId]++
			if matched[p.FileId] == nil {
				matched[p.FileId] = map[string]bool{}
			}
			matched[p.FileId][w] = true
		}
	}

	type hit struct {
		path  string
		score int
	}
	var hits []hit
	for id, sc := range score {
		if sc == 0 || int(id) >= len(idx.Files) {
			continue
		}
		path := filepath.Clean(idx.Files[id])
		if !MatchExtension(filepath.Base(path), s.extensions) {
			continue
		}
		hits = append(hits, hit{path: path, score: sc})
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].path < hits[j].path
	})

	files := make([]string, 0, len(hits))
	for _, h := range hits {
		files = append(files, h.path)
	}
	return files, nil
}
