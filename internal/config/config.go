package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	SearchDir    string
	OutputFile   string
	FromPath     string
	ReviewMode   bool
	ChangedMode  bool
	ListOnly     bool
	PathHeader   bool
	IndexWords   bool
	ReviewArg    string
	AppendReview bool
	Extensions   []string
}

func Parse() (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("mycontext", flag.ExitOnError)

	fs.Usage = printHelp(fs)

	bindString(fs, &cfg.SearchDir, ".", "d", "dir")
	bindString(fs, &cfg.OutputFile, "", "o", "output")
	bindString(fs, &cfg.FromPath, "", "F", "from-file", "selected")
	bindBool(fs, &cfg.ListOnly, false, "l", "list-only")
	bindBool(fs, &cfg.PathHeader, false, "p", "path-header")
	bindBool(fs, &cfg.ReviewMode, false, "r", "review")
	bindBool(fs, &cfg.ChangedMode, false, "c", "changed")
	bindBool(fs, &cfg.IndexWords, false, "index-words")

	fs.Func("append-review", "", func(s string) error {
		cfg.AppendReview = true
		cfg.ReviewArg = s
		return nil
	})

	_ = fs.Parse(os.Args[1:])

	if cfg.FromPath == "true" || (containsArg("-F", "--from-file", "--selected") && cfg.FromPath == "") {
		cfg.FromPath = ".idea/selectedFiles.txt"
	}

	for _, ext := range fs.Args() {
		for _, part := range strings.Fields(ext) {
			part = strings.TrimSpace(part)
			if part != "" {
				if !strings.HasPrefix(part, ".") {
					part = "." + part
				}
				cfg.Extensions = append(cfg.Extensions, part)
			}
		}
	}

	if cfg.OutputFile == "" {
		cfg.OutputFile = defaultOutputPath()
	}

	return cfg, nil
}

func printHelp(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(os.Stderr, `mycontext - Context aggregator for LLM prompts

USAGE:
  mycontext [options] [extensions...]

MODES (select one):
  <extensions...>          Scan directory for specific extensions (e.g. .go .ts)
  -r, --review             Git review mode: dump changed files + git diff against main/master
  -c, --changed            Git changed mode: dump content of changed files only (no diff)
  -F, --from-file [path]   Read file list from file (default: .idea/selectedFiles.txt)

OPTIONS:
  -d, --dir <path>         Search directory (default: ".")
  -o, --output <path>      Output file path (default: .idea/output.txt or output.txt)
  -l, --list-only          List matching file paths only, omit contents
  -p, --path-header        Include path header before file content
  --append-review [file]   Append review notes from .idea/review (or specified file)
  --index-words            Build unique words index into .idea/word_index.json
  -h, --help               Show this help message

EXAMPLES:
  mycontext .go .sql
  mycontext -r
  mycontext -F
  mycontext --index-words .go
`)
	}
}

func defaultOutputPath() string {
	if info, err := os.Stat(".idea"); err == nil && info.IsDir() {
		return filepath.Join(".idea", "output.txt")
	}
	return "output.txt"
}

func bindString(fs *flag.FlagSet, target *string, def string, names ...string) {
	for _, name := range names {
		fs.StringVar(target, name, def, "")
	}
}

func bindBool(fs *flag.FlagSet, target *bool, def bool, names ...string) {
	for _, name := range names {
		fs.BoolVar(target, name, def, "")
	}
}

func containsArg(names ...string) bool {
	for _, arg := range os.Args[1:] {
		for _, name := range names {
			if arg == name || strings.HasPrefix(arg, name+"=") {
				return true
			}
		}
	}
	return false
}
