package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
	// Ensure data directory exists if it's a file path
	dir := filepath.Dir(dataSourceName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}

	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS canvas_spaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		canvas_space_id TEXT UNIQUE NOT NULL,
		room_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}
