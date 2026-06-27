package database

import (
	"database/sql"
	"log"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	dbOnce sync.Once
	db     *sql.DB
)

// GetDB returns the singleton SQLite connection.
// The database file is created at local_data/jobs.db relative to the working directory.
func GetDB() *sql.DB {
	dbOnce.Do(func() {
		var err error
		db, err = sql.Open("sqlite", "local_data/jobs.db")
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
	return err
}
