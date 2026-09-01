package source

import (
	"mycontext/internal/config"
	"strings"
)

type Source interface {
	Collect() ([]string, error)
	Header() string
}

func MatchExtension(filename string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}
	for _, ext := range extensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func Resolve(cfg *config.Config) Source {
	strategies := []struct {
		condition bool
		create    func() Source
	}{
		{cfg.ReviewMode || cfg.ChangedMode, func() Source { return NewGitSource(cfg.Extensions) }},
		{cfg.FromPath != "", func() Source { return NewListSource(cfg.FromPath, cfg.Extensions) }},
		{true, func() Source { return NewDirSource(cfg.SearchDir, cfg.Extensions) }},
	}

	for _, s := range strategies {
		if s.condition {
			return s.create()
		}
	}
	return nil
}
