package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Use a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "dbtest")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // clean up

	dbPath := filepath.Join(tempDir, "test.db")

	// Call InitDB with the temporary path
	InitDB(dbPath)

	// Check if file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}

	// Verify DB variable is set
	if DB == nil {
		t.Fatal("Global DB variable is nil")
	}

	// Verify table exists by querying sqlite_master
	var tableName string
	err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='canvas_spaces'").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("Table 'canvas_spaces' not created")
		} else {
			t.Errorf("Error checking table existence: %v", err)
		}
	}

	// Close DB to allow cleanup
	DB.Close()
}
