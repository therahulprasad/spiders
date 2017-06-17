package db

import "database/sql"
import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/therahulprasad/spiderman/crawler/config"
	"log"
)

var db *sql.DB = nil

// Database related functions
func Setup(configuration config.Configuration, resume bool) {
	connect_db(configuration)
	if !resume {
		create_tables()
	}
}

func connect_db(configuration config.Configuration) {
	var err error

	// Create new SQLite database
	db, err = sql.Open("sqlite3", configuration.Directory + "/db.sqlite3")
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Creates base tables in sqlite database
func create_tables() {
	if db != nil {
		// Create config table
		sql_config_table := `
			CREATE TABLE config (
			    id    INTEGER PRIMARY KEY AUTOINCREMENT,
			    name  STRING  UNIQUE,
			    value TEXT
			);`
		_, err := db.Exec(sql_config_table)
		if err != nil {
			log.Fatal("Could not create config table")
		}

		// Create Queue database
		sql_queue_table := `
			CREATE TABLE queue (
			    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			    link       STRING   UNIQUE,
			    added_on   DATETIME,
			    status     STRING   DEFAULT waiting,
			    crawled_on DATETIME,
			    parent_id  INTEGER,
			    matches    INTEGER,
			    md5	       STRING
			);`
		_, err = db.Exec(sql_queue_table)
		if err != nil {
			log.Fatal("Could not create Queue table")
		}
	} else {
		log.Fatal("Database not initialized")
	}
}