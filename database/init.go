package database

import (
	"database/sql"
	"log"
	"sync"

	commonconst "github.com/vanshsinhaa/jobscanner/common_const"
	_ "modernc.org/sqlite"
)

var (
	dbOnce sync.Once
	db     *sql.DB
)

// GetDB returns the singleton SQLite connection.
func GetDB() *sql.DB {
	dbOnce.Do(func() {
		var err error
		db, err = sql.Open("sqlite", commonconst.DBPath())
		if err != nil {
			log.Fatal("failed to open SQLite db:", err)
		}
		if err = initSchema(db); err != nil {
			log.Fatal("failed to initialize schema:", err)
		}
	})
	return db
}

func initSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS jobs (
		id           INTEGER  PRIMARY KEY AUTOINCREMENT,
		company      TEXT     NOT NULL,
		job_id       TEXT     NOT NULL UNIQUE,
		title        TEXT     NOT NULL,
		location     TEXT,
		posted_on    TEXT,
		external_url TEXT     NOT NULL,
		role_type    TEXT     NOT NULL DEFAULT 'general',
		inserted_on  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS target_companies (
		name TEXT PRIMARY KEY
	)`)
	return err
}
