package repository

import (
	"database/sql"
	"errors"
	"time"
)

type CanvasSpace struct {
	ID            int64     `json:"id"`
	CanvasSpaceID string    `json:"canvas_space_id"`
	RoomName      string    `json:"room_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type CanvasSpaceRepository interface {
	Create(canvasSpaceID, roomName string) (*CanvasSpace, error)
	GetByCanvasSpaceID(id string) (*CanvasSpace, error)
	GetAll() ([]CanvasSpace, error)
	Delete(id string) error
}

type SQLiteCanvasSpaceRepository struct {
	DB *sql.DB
}

func NewSQLiteCanvasSpaceRepository(db *sql.DB) *SQLiteCanvasSpaceRepository {
	return &SQLiteCanvasSpaceRepository{DB: db}
}

func (r *SQLiteCanvasSpaceRepository) Create(canvasSpaceID, roomName string) (*CanvasSpace, error) {
	query := `INSERT INTO canvas_spaces (canvas_space_id, room_name) VALUES (?, ?)`
	result, err := r.DB.Exec(query, canvasSpaceID, roomName)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &CanvasSpace{
		ID:            id,
		CanvasSpaceID: canvasSpaceID,
		RoomName:      roomName,
		CreatedAt:     time.Now(), // Approximate, for return value
	}, nil
}

func (r *SQLiteCanvasSpaceRepository) GetByCanvasSpaceID(id string) (*CanvasSpace, error) {
	query := `SELECT id, canvas_space_id, room_name, created_at FROM canvas_spaces WHERE canvas_space_id = ?`
	row := r.DB.QueryRow(query, id)

	var cs CanvasSpace
	err := row.Scan(&cs.ID, &cs.CanvasSpaceID, &cs.RoomName, &cs.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &cs, nil
}

func (r *SQLiteCanvasSpaceRepository) GetAll() ([]CanvasSpace, error) {
	query := `SELECT id, canvas_space_id, room_name, created_at FROM canvas_spaces`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spaces []CanvasSpace
	for rows.Next() {
		var cs CanvasSpace
		if err := rows.Scan(&cs.ID, &cs.CanvasSpaceID, &cs.RoomName, &cs.CreatedAt); err != nil {
			return nil, err
		}
		spaces = append(spaces, cs)
	}

	return spaces, nil
}

func (r *SQLiteCanvasSpaceRepository) Delete(id string) error {
	query := `DELETE FROM canvas_spaces WHERE canvas_space_id = ?`
	_, err := r.DB.Exec(query, id)
	return err
}
