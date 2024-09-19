package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/schollz/progressbar/v3"
)

type DAPP struct {
	Dirs       []string
	Delete     bool
	Matchcount int
	DB         *sqlx.DB
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
				store in database
				return nil
			}
			// if d.Name() == ".filehashes.db" {
			// 	return nil
			// }

			// f, err := d.Info()
			// if err != nil {
			// 	return fmt.Errorf("error: %w", err)
			// }

			// // hash file
			// hash, err := hashFile(path)
			// if err != nil {
			// 	return fmt.Errorf("error calculating hash: %w", err)
			// }

			// err = a.StoreFileInDatabase(path, uint64(f.Size()), string(hash[:]))
			// if err != nil {
			// 	return fmt.Errorf("error: %w", err)
			// }
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


func (a *DAPP) FindDuplicate() error {
	dupList, err := a.GetDuplicateFilesFromDatabase(a.Matchcount)
	if err != nil {
		return fmt.Errorf("error retrieving duplicate files: %w", err)
	}

	for i, r := range dupList {
		if i == 0 {
			if a.Delete {
				fmt.Println("#rm -vf \"" + r.Path + "\"")
			} else {
				fmt.Println(r.Path)
			}
			continue
		}
		if r.Digest != dupList[i-1].Digest {
			fmt.Println()
		}
		if a.Delete {
			fmt.Println("#rm -vf \"" + r.Path + "\"")
		} else {
			fmt.Println(r.Path)
		}

	}

	return nil
}
