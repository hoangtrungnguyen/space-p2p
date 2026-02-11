package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	// Ensure data directory exists
	dataPath := "data"
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Fatal(err)
	}

	dbPath := filepath.Join(dataPath, "space.db")
	var err error
	DB, err = sql.Open("sqlite", dbPath)
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
