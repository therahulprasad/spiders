package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/therahulprasad/spiders/crawler/config"
)

func Setup(configuration config.Configuration, resume bool) error {
	if err := connectDB(configuration); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if !resume {
		if err := createTables(); err != nil {
			return fmt.Errorf("failed to create tables: %w", err)
		}
	}

	return prepareStatements()
}

func prepareStatements() error {
	var err error

	pushStmt, err = db.Prepare("INSERT OR IGNORE INTO queue(link, added_on, status, parent_id) VALUES(?,?,?,?)")
	if err != nil {
		return fmt.Errorf("failed to prepare push statement: %w", err)
	}

	pushMultiStmt, err = db.Prepare("INSERT OR IGNORE INTO queue(link, added_on, status, parent_id) VALUES(?,?,?,?)")
	if err != nil {
		return fmt.Errorf("failed to prepare push multi statement: %w", err)
	}

	sqlPop := `UPDATE queue SET status=? WHERE id IN (
				    SELECT id FROM queue WHERE status = ? ORDER BY id ASC LIMIT 1
				)`
	popStmt, err = db.Prepare(sqlPop)
	if err != nil {
		return fmt.Errorf("failed to prepare pop statement: %w", err)
	}

	sqlUpdate := "UPDATE queue SET status=?, matches=?, md5=? WHERE id=?"
	updateStmt, err = db.Prepare(sqlUpdate)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}

	return nil
}

func connectDB(configuration config.Configuration) error {
	var err error
	dbPath := configuration.Directory + "/db.sqlite3"
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database %s: %w", dbPath, err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func createTables() error {
	if db == nil {
		return errors.New("database not initialized")
	}

	sqlConfigTable := `
		CREATE TABLE IF NOT EXISTS config (
		    id    INTEGER PRIMARY KEY AUTOINCREMENT,
		    name  TEXT UNIQUE NOT NULL,
		    value TEXT
		);`

	if _, err := db.Exec(sqlConfigTable); err != nil {
		return fmt.Errorf("failed to create config table: %w", err)
	}

	sqlQueueTable := `
		CREATE TABLE IF NOT EXISTS queue (
		    id         INTEGER PRIMARY KEY AUTOINCREMENT,
		    link       TEXT UNIQUE NOT NULL,
		    added_on   DATETIME NOT NULL,
		    status     TEXT DEFAULT 'waiting',
		    crawled_on DATETIME,
		    parent_id  INTEGER,
		    matches    INTEGER DEFAULT 0,
		    md5        TEXT
		);`

	if _, err := db.Exec(sqlQueueTable); err != nil {
		return fmt.Errorf("failed to create queue table: %w", err)
	}

	return nil
}
