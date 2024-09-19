package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// _ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
	_ "modernc.org/sqlite"
)

func main() {
	// Initialize Application
	a := DAPP{}
	fdb := ":memory:"
	// fdb := ".filehashes.db"
	var err error

	os.Remove(fdb)
	a.DB, err = sqlx.Open("sqlite", fdb)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	defer a.DB.Close()
	defer os.Remove(fdb)

	// Initialize Database
	if err := a.InitializeDatabase(); err != nil {
		fmt.Printf("database initialization failed: %v\n", err)
		return
	}

	var verboseCount int
	app := &cli.App{
		Name:  "rd",
		Usage: "remove empty directories",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "increase verbosity",
				Value:   false,
				Count:   &verboseCount,
			},
		},
		Action: func(cCtx *cli.Context) error {
			a.Dirs = cCtx.Args().Slice()
			if len(a.Dirs) == 0 {
				a.Dirs = append(a.Dirs, ".")
			}
			a.Verbose = cCtx.Bool("verbose")

			// fmt.Println("Filewalker")
			if err := a.Filewalker(); err != nil {
				fmt.Printf("filewalker error: %v\n", err)
				return err
			}

			err = a.TrimDirs()
			if err != nil {
				fmt.Printf("trim dirs error: %v\n", err)
				return err
			}

			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error running app: %v\n", err)
	}
}

type DAPP struct {
	Dirs    []string
	Verbose bool
	DB      *sqlx.DB
}

func (a *DAPP) Filewalker() error {
	for _, dir := range a.Dirs {
		bar := progressbar.NewOptions(-1,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("files"),
			progressbar.OptionSetElapsedTime(false),
			progressbar.OptionSetPredictTime(false),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("scanning files"),
		)
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("error: %w", err)
			}
			if d.IsDir() {
				// store in database
				slashCount := strings.Count(path, string(os.PathSeparator))
				err = a.StoreDirInDatabase(path, uint64(slashCount))
				if err != nil {
					return fmt.Errorf("error: %w", err)
				}
				return nil
			}
			bar.Add(1)
			return nil
		})
		if err != nil {
			return err
		}
		bar.Finish()
	}
	return nil
}

func (a *DAPP) TrimDirs() error {
	dirList, err := a.GetFilesFromDatabase()
	if err != nil {
		return fmt.Errorf("error retrieving directory list: %w", err)
	}

	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("directories"),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("removing empty directories"),
	)
	for _, d := range dirList {
		// fmt.Println(i, r.Path)
		if err := os.Remove(d.Path); err != nil {
			if a.Verbose {
				fmt.Println("error removing directory:", err)
			}
			continue
		}
		bar.Add(1)
	}
	bar.Finish()

	return nil
}

func (a *DAPP) InitializeDatabase() error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS dir_list (
		path TEXT NOT NULL,
		size UINT64
		);`

	_, err := a.DB.Exec(createTableSQL)
	return err
}

func (a *DAPP) StoreDirInDatabase(path string, size uint64) error {
	storeFileSQL := `INSERT INTO dir_list (path,size) VALUES (?,?);`
	_, err := a.DB.Exec(storeFileSQL, path, size)
	return err
}

type Directories struct {
	Path string
}

func (a *DAPP) GetFilesFromDatabase() ([]Directories, error) {
	dirList := []Directories{}
	SELECT_DIRS := `SELECT path FROM dir_list ORDER BY size desc, path`
	err := a.DB.Select(&dirList, SELECT_DIRS)
	return dirList, err
}
