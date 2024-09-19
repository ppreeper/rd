package main

func (a *DAPP) InitializeDatabase() error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS file_hashes (
		path TEXT NOT NULL,
		size UINT64,
		digest TEXT
		);`

	_, err := a.DB.Exec(createTableSQL)
	return err
}

func (a *DAPP) StoreFileInDatabase(path string, size uint64, hash string) error {
	storeFileSQL := `INSERT INTO file_hashes (path,size,digest) VALUES (?,?,?);`
	_, err := a.DB.Exec(storeFileSQL, path, size, hash)
	return err
}

func (a *DAPP) StoreHashInDatabase(hash string, rowid int64) error {
	storeHashSQL := `UPDATE file_hashes SET digest = ? WHERE rowid = ?;`
	_, err := a.DB.Exec(storeHashSQL, hash, rowid)
	return err
}

type Files struct {
	Rowid int64
	Path  string
}

func (a *DAPP) GetFilesFromDatabase() ([]Files, error) {
	fileList := []Files{}
	SELECT_FILES := `SELECT rowid,path FROM file_hashes ORDER BY size, path`
	err := a.DB.Select(&fileList, SELECT_FILES)
	return fileList, err
}

type Duplicates struct {
	Digest string
	Path   string
}

func (a *DAPP) GetDuplicateFilesFromDatabase(matchCount int) ([]Duplicates, error) {
	fileList := []Duplicates{}
	DUPLICATES := `select digest,path from file_hashes where digest in (select digest from file_hashes group by digest having count(digest) > ?) order by digest, path`
	err := a.DB.Select(&fileList, DUPLICATES, matchCount)
	return fileList, err
}
