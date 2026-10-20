package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"mycontext/internal/config"
	"mycontext/internal/output"
	"mycontext/internal/review"
	"mycontext/internal/source"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	_ = os.MkdirAll(filepath.Dir(cfg.OutputFile), 0755)
	outFile, err := os.Create(cfg.OutputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal creating file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	buf := bufio.NewWriterSize(outFile, 64*1024)
	defer func() {
		if err := buf.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "flush: %v\n", err)
			os.Exit(1)
		}
	}()

	if cfg.AppendReview {
		_ = review.Append(buf, cfg.ReviewArg, ".idea/review")
	}

	src := source.Resolve(cfg)
	files, err := src.Collect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal collecting files: %v\n", err)
		os.Exit(1)
	}

	if header := src.Header(); header != "" {
		buf.WriteString(header)
	}

	formatter := output.ResolveFormatter(cfg.ListOnly, cfg.PathHeader)
	var processed []string

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			if err := formatter.Write(buf, file); err == nil {
				processed = append(processed, file)
			}
		}
	}

	if cfg.ReviewMode {
		if gitSrc, ok := src.(*source.GitSource); ok && gitSrc.BaseBranch != "" {
			diff, _ := exec.Command("git", "--no-pager", "diff", gitSrc.BaseBranch+"...HEAD").Output()
			buf.WriteString("========== SUMMARY: GIT DIFF ==========\n\n")
			buf.Write(diff)
		}
	}

	if cfg.IndexWords {
		targetIndex := filepath.Join(".idea", "word_index.json")
		if err := output.BuildWordIndexProto(processed, targetIndex); err != nil {
			fmt.Fprintf(os.Stderr, "Error building word index: %v\n", err)
		} else {
			fmt.Printf("Saved word index for %d files to %s\n", len(processed), targetIndex)
		}
	}

	fmt.Printf("Dumped %d files to %s\n", len(processed), cfg.OutputFile)
}
