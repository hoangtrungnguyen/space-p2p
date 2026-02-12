package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"space-p2p/repository"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/livekit"
)

// MockRoomService
type MockRoomService struct {
	CreateRoomFunc       func(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error)
	ListRoomsFunc        func(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error)
	DeleteRoomFunc       func(ctx context.Context, req *livekit.DeleteRoomRequest) (*livekit.DeleteRoomResponse, error)
	ListParticipantsFunc func(ctx context.Context, req *livekit.ListParticipantsRequest) (*livekit.ListParticipantsResponse, error)
}

func (m *MockRoomService) CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
	if m.CreateRoomFunc != nil {
		return m.CreateRoomFunc(ctx, req)
	}
	return &livekit.Room{Name: req.Name}, nil
}

func (m *MockRoomService) ListRooms(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error) {
	if m.ListRoomsFunc != nil {
		return m.ListRoomsFunc(ctx, req)
	}
	return &livekit.ListRoomsResponse{}, nil
}

func (m *MockRoomService) DeleteRoom(ctx context.Context, req *livekit.DeleteRoomRequest) (*livekit.DeleteRoomResponse, error) {
	if m.DeleteRoomFunc != nil {
		return m.DeleteRoomFunc(ctx, req)
	}
	return &livekit.DeleteRoomResponse{}, nil
}

func (m *MockRoomService) ListParticipants(ctx context.Context, req *livekit.ListParticipantsRequest) (*livekit.ListParticipantsResponse, error) {
	if m.ListParticipantsFunc != nil {
		return m.ListParticipantsFunc(ctx, req)
	}
	return &livekit.ListParticipantsResponse{}, nil
}

// MockCanvasSpaceRepository
type MockCanvasSpaceRepository struct {
	CreateFunc             func(canvasSpaceID, roomName string) (*repository.CanvasSpace, error)
	GetByCanvasSpaceIDFunc func(id string) (*repository.CanvasSpace, error)
	GetAllFunc             func() ([]repository.CanvasSpace, error)
	DeleteFunc             func(id string) error
}

func (m *MockCanvasSpaceRepository) Create(canvasSpaceID, roomName string) (*repository.CanvasSpace, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(canvasSpaceID, roomName)
	}
	return &repository.CanvasSpace{CanvasSpaceID: canvasSpaceID, RoomName: roomName}, nil
}

func (m *MockCanvasSpaceRepository) GetByCanvasSpaceID(id string) (*repository.CanvasSpace, error) {
	if m.GetByCanvasSpaceIDFunc != nil {
		return m.GetByCanvasSpaceIDFunc(id)
	}
	return nil, nil
}

func (m *MockCanvasSpaceRepository) GetAll() ([]repository.CanvasSpace, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return []repository.CanvasSpace{}, nil
}

func (m *MockCanvasSpaceRepository) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/rooms", createRoom)
	r.GET("/rooms", listRooms)
	r.POST("/token", generateToken)
	r.DELETE("/rooms/:room", deleteRoom)
	r.GET("/rooms/:room/participants/count", getRoomParticipantCount)
	r.GET("/rooms/:room/ping", pingRoom)

	r.POST("/canvas-spaces/token", generateTokenForCanvasSpace)
	r.GET("/canvas-spaces", listCanvasSpaces)
	r.GET("/canvas-spaces/:id", getCanvasSpace)
	r.DELETE("/canvas-spaces/:id", deleteCanvasSpace)
	r.GET("/canvas-spaces/:id/participants/count", getCanvasSpaceParticipantCount)
	return r
}

func TestCreateRoom(t *testing.T) {
	mockRoomService := &MockRoomService{
		CreateRoomFunc: func(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
			if req.Name == "test-room" {
				return &livekit.Room{Name: "test-room"}, nil
			}
			return nil, errors.New("unexpected room name")
		},
	}
	roomClient = mockRoomService

	r := setupRouter()

	reqBody := CreateRoomRequest{RoomName: "test-room"}
	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/rooms", bytes.NewBuffer(jsonBytes))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var room livekit.Room
	if err := json.Unmarshal(w.Body.Bytes(), &room); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if room.Name != "test-room" {
		t.Errorf("Expected room name test-room, got %s", room.Name)
	}
}

func TestListRooms(t *testing.T) {
	mockRoomService := &MockRoomService{
		ListRoomsFunc: func(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error) {
			return &livekit.ListRoomsResponse{
				Rooms: []*livekit.Room{{Name: "room1"}, {Name: "room2"}},
			}, nil
		},
	}
	roomClient = mockRoomService

	r := setupRouter()

	req, _ := http.NewRequest("GET", "/rooms", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	// Verify count locally if parsed
}

func TestGenerateToken(t *testing.T) {
	// Need to set config for token generation
	config = Config{ApiKey: "testapikey", ApiSecret: "testapisecret"}

	r := setupRouter()

	reqBody := TokenRequest{RoomName: "room1", Identity: "user1"}
	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/token", bytes.NewBuffer(jsonBytes))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestGenerateTokenForCanvasSpace(t *testing.T) {
	// Test creating new mapping
	mockRepo := &MockCanvasSpaceRepository{
		GetByCanvasSpaceIDFunc: func(id string) (*repository.CanvasSpace, error) {
			return nil, nil // Not found, create new
		},
		CreateFunc: func(id, room string) (*repository.CanvasSpace, error) {
			return &repository.CanvasSpace{ID: 1, CanvasSpaceID: id, RoomName: room, CreatedAt: time.Now()}, nil
		},
	}
	mockRoomService := &MockRoomService{
		CreateRoomFunc: func(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
			return &livekit.Room{Name: req.Name}, nil
		},
	}

	canvasRepo = mockRepo
	roomClient = mockRoomService
	config = Config{ApiKey: "testapikey", ApiSecret: "testapisecret"}

	r := setupRouter()

	reqBody := CanvasTokenRequest{CanvasSpaceID: "canvas1", Identity: "user1"}
	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/canvas-spaces/token", bytes.NewBuffer(jsonBytes))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestGetCanvasSpaceParticipantCount(t *testing.T) {
	mockRepo := &MockCanvasSpaceRepository{
		GetByCanvasSpaceIDFunc: func(id string) (*repository.CanvasSpace, error) {
			return &repository.CanvasSpace{CanvasSpaceID: id, RoomName: "mapped-room"}, nil
		},
	}
	mockRoomService := &MockRoomService{
		ListParticipantsFunc: func(ctx context.Context, req *livekit.ListParticipantsRequest) (*livekit.ListParticipantsResponse, error) {
			if req.Room == "mapped-room" {
				return &livekit.ListParticipantsResponse{
					Participants: []*livekit.ParticipantInfo{{Identity: "p1"}},
				}, nil
			}
			return nil, errors.New("room not found")
		},
	}

	canvasRepo = mockRepo
	roomClient = mockRoomService

	r := setupRouter()

	req, _ := http.NewRequest("GET", "/canvas-spaces/c1/participants/count", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestListActiveRooms(t *testing.T) {
	mockRoomService := &MockRoomService{
		ListRoomsFunc: func(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error) {
			return &livekit.ListRoomsResponse{
				Rooms: []*livekit.Room{
					{Name: "empty-room", NumParticipants: 0},
					{Name: "active-room", NumParticipants: 2},
				},
			}, nil
		},
	}
	roomClient = mockRoomService

	r := setupRouter()

	// 1. Test filtering active=true
	req, _ := http.NewRequest("GET", "/rooms?active=true", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var rooms []livekit.Room
	if err := json.Unmarshal(w.Body.Bytes(), &rooms); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(rooms) != 1 {
		t.Fatalf("Expected 1 active room, got %d", len(rooms))
	}
	if rooms[0].Name != "active-room" {
		t.Errorf("Expected room 'active-room', got '%s'", rooms[0].Name)
	}

	// 2. Test default (all rooms)
	reqDefault, _ := http.NewRequest("GET", "/rooms", nil)
	wDefault := httptest.NewRecorder()
	r.ServeHTTP(wDefault, reqDefault)

	if wDefault.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", wDefault.Code)
	}

	var allRooms []livekit.Room
	if err := json.Unmarshal(wDefault.Body.Bytes(), &allRooms); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(allRooms) != 2 {
		t.Errorf("Expected 2 rooms (default), got %d", len(allRooms))
	}
}

func TestPingRoom(t *testing.T) {
	mockRoomService := &MockRoomService{
		ListRoomsFunc: func(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error) {
			if len(req.Names) > 0 && req.Names[0] == "existing-room" {
				return &livekit.ListRoomsResponse{
					Rooms: []*livekit.Room{{Name: "existing-room", NumParticipants: 5, Sid: "RM_123"}},
				}, nil
			}
			return &livekit.ListRoomsResponse{Rooms: []*livekit.Room{}}, nil
		},
	}
	roomClient = mockRoomService

	r := setupRouter()

	// 1. Test existing room
	req, _ := http.NewRequest("GET", "/rooms/existing-room/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for existing room, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if active, ok := resp["active"].(bool); !ok || !active {
		t.Errorf("Expected active=true, got %v", resp["active"])
	}
	if count, ok := resp["participants"].(float64); !ok || count != 5 {
		t.Errorf("Expected participants=5, got %v", resp["participants"])
	}

	// 2. Test non-existent room
	req2, _ := http.NewRequest("GET", "/rooms/unknown-room/ping", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for unknown room, got %d", w2.Code)
	}
}
