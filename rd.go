package main

import (
	"fmt"
	"os"

	// _ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
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

	var deleteCount int
	app := &cli.App{
		Name:  "fdup",
		Usage: "Find duplicate files",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Add Delete commands to duplicate files",
				Value:   false,
				Count:   &deleteCount,
			},
			&cli.IntFlag{
				Name:    "matchcount",
				Aliases: []string{"m"},
				Usage:   "Minimum count of duplicate files to show",
				Value:   1,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.Int("matchcount") < 1 {
				return fmt.Errorf("matchcount must be greater than 0")
			}
			a.Dirs = cCtx.Args().Slice()
			if len(a.Dirs) == 0 {
				a.Dirs = append(a.Dirs, ".")
			}
			a.Delete = cCtx.Bool("delete")
			a.Matchcount = cCtx.Int("matchcount")

			// fmt.Println("Filewalker")
			if err := a.Filewalker(); err != nil {
				fmt.Printf("filewalker error: %v\n", err)
				return err
			}
			// err = a.FindEmpty()
			// if err != nil {
			// 	fmt.Printf("checkduplicate error: %v\n", err)
			// 	return err
			// }

			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error running app: %v\n", err)
	}
}
