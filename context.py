#!/usr/bin/env python3
import argparse
import subprocess
import sys
from pathlib import Path
from typing import List, Generator

def get_base_branch() -> str | None:
    for branch in ['origin/main', 'origin/master', 'main', 'master']:
        cmd = ['git', 'rev-parse', '--verify', branch]
        if subprocess.run(cmd, capture_output=True).returncode == 0:
            return branch
    return None

def normalize_extensions(exts: List[str]) -> List[str]:
    result = []
    for ext in exts:
        for part in ext.split():
            part = part.strip()
            if part:
                result.append(part if part.startswith('.') else f'.{part}')
    return result

def get_output_path(custom_output: Path | None) -> Path:
    if custom_output:
        return custom_output
    idea_dir = Path('.idea')
    return idea_dir / 'output.txt' if idea_dir.is_dir() else Path('output.txt')

def find_files(search_dir: Path, extensions: List[str]) -> Generator[Path, None, None]:
    for path in search_dir.rglob('*'):
        if path.is_file():
            if not extensions or any(path.name.endswith(ext) for ext in extensions):
                yield path

def run_review_mode(out_f, extensions: List[str]) -> None:
    base_branch = get_base_branch()
    if not base_branch:
        sys.exit("Error: Could not find 'origin/main', 'origin/master', 'main', or 'master' branch.")

    diff_target = f"{base_branch}...HEAD"
    print(f"Review mode: comparing with {diff_target}...")

    name_cmd = ['git', '--no-pager', 'diff', '--name-only', '--diff-filter=d', diff_target]
    res = subprocess.run(name_cmd, capture_output=True, text=True, check=True)
    changed_files = [Path(f) for f in res.stdout.splitlines() if f]

    out_f.write("========== EDITED FILES FULL CONTENT ==========\n\n")
    processed = 0

    for path in changed_files:
        if extensions and not any(path.name.endswith(ext) for ext in extensions):
            continue
        try:
            out_f.write(f'--- FILE: "{path}" ---\n')
            out_f.write(path.read_text(encoding='utf-8'))
            out_f.write('\n\n')
            processed += 1
        except Exception as e:
            print(f"Error reading {path}: {e}", file=sys.stderr)

    diff_cmd = ['git', '--no-pager', 'diff', diff_target]
    diff_res = subprocess.run(diff_cmd, capture_output=True, text=True, check=True)
    out_f.write("========== SUMMARY: GIT DIFF ==========\n\n")
    out_f.write(diff_res.stdout)
    out_f.write("\n")

    print(f"Processed {processed} files and added git diff summary.")

def main():
    parser = argparse.ArgumentParser(
        prog="context",
        description="Unified tool to list file paths, export file contents, or generate git review context."
    )
    parser.add_argument('extensions', nargs='*', help="File extensions to process (e.g., py js .txt)")
    parser.add_argument('-d', '--dir', type=Path, default=Path('.'), help="Search directory (default: current directory)")
    parser.add_argument('-o', '--output', type=Path, help="Output file path (default: output.txt or .idea/output.txt)")
    parser.add_argument('-p', '--path-header', action='store_true', help="Include file path header before content")
    parser.add_argument('-l', '--list-only', action='store_true', help="List file paths only without dumping file content")
    parser.add_argument('-r', '--review', action='store_true', help="Run in git review mode comparing HEAD against main/master")

    args = parser.parse_args()

    if not args.extensions and not args.review:
        parser.error("At least one file extension or --review flag must be specified.")

    extensions = normalize_extensions(args.extensions)
    output_path = get_output_path(args.output)

    try:
        with open(output_path, 'w', encoding='utf-8') as out_f:
            if args.review:
                run_review_mode(out_f, extensions)
            elif args.list_only:
                count = 0
                for file_path in find_files(args.dir, extensions):
                    out_f.write(f"{file_path}\n")
                    count += 1
                print(f"Saved list of {count} files to {output_path}.")
            else:
                count = 0
                for file_path in find_files(args.dir, extensions):
                    try:
                        if args.path_header:
                            out_f.write(f'"{file_path}"\n')
                        out_f.write(file_path.read_text(encoding='utf-8'))
                        out_f.write('\n')
                        count += 1
                    except Exception as e:
                        print(f"Error reading {file_path}: {e}", file=sys.stderr)
                
                ext_str = ', '.join(extensions)
                print(f"Saved content of {count} files ({ext_str}) to {output_path}.")

    except Exception as e:
        sys.exit(f"Fatal error: {e}")

if __name__ == "__main__":
    main()