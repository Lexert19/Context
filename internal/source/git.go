package source

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type GitSource struct {
	extensions []string
	BaseBranch string
}

func NewGitSource(exts []string) *GitSource {
	return &GitSource{extensions: exts, BaseBranch: detectBaseBranch()}
}

func (g *GitSource) Header() string {
	return "========== EDITED FILES FULL CONTENT ==========\n\n"
}

func (g *GitSource) Collect() ([]string, error) {
	if g.BaseBranch == "" {
		return nil, fmt.Errorf("could not find base branch (main/master)")
	}
	out, err := exec.Command("git", "--no-pager", "diff", "--name-only", "--diff-filter=d", g.BaseBranch+"...HEAD").Output()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && MatchExtension(filepath.Base(line), g.extensions) {
			files = append(files, filepath.Clean(line))
		}
	}
	return files, nil
}

func detectBaseBranch() string {
	for _, b := range []string{"main", "master", "origin/main", "origin/master"} {
		if err := exec.Command("git", "rev-parse", "--verify", b).Run(); err == nil {
			return b
		}
	}
	return ""
}
