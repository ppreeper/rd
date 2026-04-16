package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "rd",
		Usage: "remove empty directories",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "count directories that would be removed",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "increase verbosity",
				Value:   false,
			},
		},
		Action: func(_ context.Context, cCtx *cli.Command) error {
			dirs := cCtx.Args().Slice()
			if len(dirs) == 0 {
				dirs = []string{"."}
			}

			progress := newProgressReporter(cCtx.Bool("dry-run"))
			defer progress.Stop()

			removed, err := pruneEmptyDirs(dirs, cCtx.Bool("dry-run"), progress)
			if err != nil {
				return err
			}
			if cCtx.Bool("dry-run") {
				fmt.Printf("would remove %d directories\n", removed)
			} else if cCtx.Bool("verbose") {
				fmt.Printf("removed %d directories\n", removed)
			}
			return nil
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("error running app: %v\n", err)
	}
}

func pruneEmptyDirs(roots []string, dryRun bool, p *progressReporter) (int, error) {
	total := 0
	for _, root := range roots {
		_, removed, err := pruneDir(root, dryRun, p)
		if err != nil {
			return total, err
		}
		total += removed
	}
	return total, nil
}

func pruneDir(path string, dryRun bool, p *progressReporter) (bool, int, error) {
	p.Scanned()
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, 0, err
	}

	hasRemaining := false
	removedCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			hasRemaining = true
			continue
		}

		childRemoved, childCount, err := pruneDir(filepath.Join(path, entry.Name()), dryRun, p)
		if err != nil {
			return false, removedCount, err
		}
		removedCount += childCount
		if !childRemoved {
			hasRemaining = true
		}
	}

	if hasRemaining {
		return false, removedCount, nil
	}

	if !dryRun {
		if err := os.Remove(path); err != nil {
			return false, removedCount, err
		}
	}
	p.Removed()

	return true, removedCount + 1, nil
}

type progressReporter struct {
	dryRun  bool
	scanned atomic.Int64
	removed atomic.Int64
	done    chan struct{}
}

func newProgressReporter(dryRun bool) *progressReporter {
	p := &progressReporter{dryRun: dryRun, done: make(chan struct{})}
	go p.run()
	return p
}

func (p *progressReporter) Scanned() { p.scanned.Add(1) }
func (p *progressReporter) Removed() { p.removed.Add(1) }
func (p *progressReporter) Stop()    { close(p.done) }

func (p *progressReporter) run() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	frames := []string{"|", "/", "-", "\\"}
	i := 0
	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "\r%s scanned=%d removed=%d", frames[i%len(frames)], p.scanned.Load(), p.removed.Load())
			if p.dryRun {
				fmt.Fprint(os.Stderr, " dry-run")
			}
			i++
		case <-p.done:
			fmt.Fprintf(os.Stderr, "\rscanned=%d removed=%d", p.scanned.Load(), p.removed.Load())
			if p.dryRun {
				fmt.Fprint(os.Stderr, " dry-run")
			}
			fmt.Fprintln(os.Stderr)
			return
		}
	}
}
