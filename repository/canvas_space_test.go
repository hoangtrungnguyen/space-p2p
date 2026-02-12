package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbFile := "test_space.db"
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS canvas_spaces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		canvas_space_id TEXT UNIQUE NOT NULL,
		room_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(dbFile)
	})

	return db
}

func TestSQLiteCanvasSpaceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteCanvasSpaceRepository(db)

	canvasSpaceID := "cs-123"
	roomName := "room-123"

	created, err := repo.Create(canvasSpaceID, roomName)
	if err != nil {
		t.Fatalf("Failed to create canvas space: %v", err)
	}

	if created.CanvasSpaceID != canvasSpaceID {
		t.Errorf("Expected CanvasSpaceID %s, got %s", canvasSpaceID, created.CanvasSpaceID)
	}
	if created.RoomName != roomName {
		t.Errorf("Expected RoomName %s, got %s", roomName, created.RoomName)
	}
	if created.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestSQLiteCanvasSpaceRepository_GetByCanvasSpaceID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteCanvasSpaceRepository(db)

	canvasSpaceID := "cs-456"
	roomName := "room-456"
	_, err := repo.Create(canvasSpaceID, roomName)
	if err != nil {
		t.Fatalf("Failed to create canvas space: %v", err)
	}

	t.Run("ExistingID", func(t *testing.T) {
		got, err := repo.GetByCanvasSpaceID(canvasSpaceID)
		if err != nil {
			t.Fatalf("Failed to get canvas space: %v", err)
		}
		if got == nil {
			t.Fatal("Expected canvas space, got nil")
		}
		if got.CanvasSpaceID != canvasSpaceID {
			t.Errorf("Expected CanvasSpaceID %s, got %s", canvasSpaceID, got.CanvasSpaceID)
		}
	})

	t.Run("NonExistingID", func(t *testing.T) {
		got, err := repo.GetByCanvasSpaceID("non-existent")
		if err != nil {
			t.Fatalf("Failed to get canvas space: %v", err)
		}
		if got != nil {
			t.Error("Expected nil, got canvas space")
		}
	})
}

func TestSQLiteCanvasSpaceRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteCanvasSpaceRepository(db)

	repo.Create("cs-1", "room-1")
	repo.Create("cs-2", "room-2")

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all canvas spaces: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("Expected 2 canvas spaces, got %d", len(all))
	}
}

func TestSQLiteCanvasSpaceRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSQLiteCanvasSpaceRepository(db)

	canvasSpaceID := "cs-del"
	repo.Create(canvasSpaceID, "room-del")

	err := repo.Delete(canvasSpaceID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	got, err := repo.GetByCanvasSpaceID(canvasSpaceID)
	if err != nil {
		t.Fatalf("Failed to get after delete: %v", err)
	}
	if got != nil {
		t.Error("Expected nil after delete, got object")
	}
}
