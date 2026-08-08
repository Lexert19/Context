package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getBaseBranch() string {
	branches := []string{"main", "master", "origin/main", "origin/master"}
	for _, branch := range branches {
		cmd := exec.Command("git", "rev-parse", "--verify", branch)
		if err := cmd.Run(); err == nil {
			return branch
		}
	}
	return ""
}

func normalizeExtensions(rawExts []string) []string {
	var result []string
	for _, ext := range rawExts {
		parts := strings.Fields(ext)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				if !strings.HasPrefix(part, ".") {
					part = "." + part
				}
				result = append(result, part)
			}
		}
	}
	return result
}

func getOutputPath(customOutput string) string {
	if customOutput != "" {
		return customOutput
	}
	info, err := os.Stat(".idea")
	if err == nil && info.IsDir() {
		return filepath.Join(".idea", "output.txt")
	}
	return "output.txt"
}

func matchesExtension(filename string, extensions []string) bool {
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

func appendFileContent(w io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func runReviewMode(out *bufio.Writer, extensions []string) error {
	baseBranch := getBaseBranch()
	if baseBranch == "" {
		return fmt.Errorf("could not find 'origin/main', 'origin/master', 'main', or 'master' branch")
	}

	diffTarget := fmt.Sprintf("%s...HEAD", baseBranch)
	fmt.Printf("Review mode: comparing with %s...\n", diffTarget)

	nameCmd := exec.Command("git", "--no-pager", "diff", "--name-only", "--diff-filter=d", diffTarget)
	outBytes, err := nameCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run git diff name-only: %w", err)
	}

	lines := strings.Split(string(outBytes), "\n")
	out.WriteString("========== EDITED FILES FULL CONTENT ==========\n\n")

	processed := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		path := filepath.Clean(line)
		if !matchesExtension(filepath.Base(path), extensions) {
			continue
		}

		out.WriteString(fmt.Sprintf("--- FILE: \"%s\" ---\n", filepath.ToSlash(path)))
		if err := appendFileContent(out, path); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
		}
		out.WriteString("\n\n")
		processed++
	}

	diffCmd := exec.Command("git", "--no-pager", "diff", diffTarget)
	diffOut, err := diffCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run git diff: %w", err)
	}

	out.WriteString("========== SUMMARY: GIT DIFF ==========\n\n")
	out.Write(diffOut)
	out.WriteString("\n")

	fmt.Printf("Processed %d files and added git diff summary.\n", processed)
	return nil
}

func main() {
	var searchDir string
	var customOutput string
	var pathHeader bool
	var listOnly bool
	var reviewMode bool

	flag.StringVar(&searchDir, "d", ".", "Search directory")
	flag.StringVar(&searchDir, "dir", ".", "Search directory")

	flag.StringVar(&customOutput, "o", "", "Output file path")
	flag.StringVar(&customOutput, "output", "", "Output file path")

	flag.BoolVar(&pathHeader, "p", false, "Include file path header before content")
	flag.BoolVar(&pathHeader, "path-header", false, "Include file path header before content")

	flag.BoolVar(&listOnly, "l", false, "List file paths only without dumping file content")
	flag.BoolVar(&listOnly, "list-only", false, "List file paths only without dumping file content")

	flag.BoolVar(&reviewMode, "r", false, "Run in git review mode comparing HEAD against main/master")
	flag.BoolVar(&reviewMode, "review", false, "Run in git review mode comparing HEAD against main/master")

	flag.Parse()

	rawExtensions := flag.Args()

	if len(rawExtensions) == 0 && !reviewMode {
		fmt.Fprintln(os.Stderr, "Error: At least one file extension or --review flag must be specified.")
		flag.Usage()
		os.Exit(1)
	}

	extensions := normalizeExtensions(rawExtensions)
	outputPath := getOutputPath(customOutput)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error creating output directory: %v\n", err)
		os.Exit(1)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	bufWriter := bufio.NewWriterSize(outFile, 64*1024)
	defer bufWriter.Flush()

	if reviewMode {
		if err := runReviewMode(bufWriter, extensions); err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	count := 0
	err = filepath.WalkDir(searchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if matchesExtension(d.Name(), extensions) {
			normalizedPath := filepath.ToSlash(path)

			if listOnly {
				bufWriter.WriteString(normalizedPath + "\n")
				count++
			} else {
				if pathHeader {
					fmt.Fprintf(bufWriter, "\"%s\"\n", normalizedPath)
				}
				if err := appendFileContent(bufWriter, path); err != nil {
					fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
				} else {
					bufWriter.WriteString("\n")
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error walking directory: %v\n", err)
		os.Exit(1)
	}

	if listOnly {
		fmt.Printf("Saved list of %d files to %s.\n", count, outputPath)
	} else {
		extStr := strings.Join(extensions, ", ")
		if extStr == "" {
			extStr = "all files"
		}
		fmt.Printf("Saved content of %d files (%s) to %s.\n", count, extStr, outputPath)
	}
}
