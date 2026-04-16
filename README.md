# rd

Remove empty directories.

## Installation

```bash
go install github.com/ppreeper/rd@latest
```

Or build from source:

```bash
go build -o rd .
```

## Usage

```bash
rd [flags] [paths...]
```

If no paths are given, `rd` scans the current directory.

## Flags

- `-n`, `--dry-run`: count directories that would be removed
- `-v`, `--verbose`: print the final removed count

## Examples

```bash
rd .
rd -n ./src ./testdata
rd -v /tmp/project
```
