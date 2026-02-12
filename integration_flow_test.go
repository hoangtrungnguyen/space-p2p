package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"space-p2p/db"
	"space-p2p/repository"

	"github.com/livekit/protocol/livekit"
)

func TestCanvasSpaceIntegrationFlow(t *testing.T) {
	// 1. Setup Pure SQLite DB (Temporary File)
	tempDir, err := os.MkdirTemp("", "integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "integration.db")
	db.InitDB(dbPath)
	defer func() {
		if db.DB != nil {
			db.DB.Close()
		}
	}()

	// 2. Setup Repository with Real DB
	realRepo := repository.NewSQLiteCanvasSpaceRepository(db.DB)
	canvasRepo = realRepo

	// 3. Setup Mock Room Service
	// We want to verify that CreateRoom is called exactly ONCE for the first user.
	createRoomCalled := 0
	mockRoomService := &MockRoomService{
		CreateRoomFunc: func(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
			createRoomCalled++
			return &livekit.Room{
				Name: req.Name,
				Sid:  "room_sid_" + req.Name,
			}, nil
		},
	}
	roomClient = mockRoomService

	// 4. Setup Router
	// Mock config for token generation
	config = Config{
		ApiKey:    "devkey",
		ApiSecret: "secret",
		Host:      "http://localhost:7880",
	}
	r := setupRouter()

	// ----------------------------------------------------------------
	// Step 1: User A connects to Canvas Space "canvas-101"
	// Expected: New room created, token returned.
	// ----------------------------------------------------------------
	canvasID := "canvas-101"
	userA := "user-A"

	reqBodyA := CanvasTokenRequest{
		CanvasSpaceID: canvasID,
		Identity:      userA,
	}
	bodyA, _ := json.Marshal(reqBodyA)
	reqA, _ := http.NewRequest("POST", "/canvas-spaces/token", bytes.NewBuffer(bodyA))
	wA := httptest.NewRecorder()

	r.ServeHTTP(wA, reqA)

	if wA.Code != http.StatusOK {
		t.Fatalf("User A request failed with status %d: %s", wA.Code, wA.Body.String())
	}

	var respA map[string]string
	if err := json.Unmarshal(wA.Body.Bytes(), &respA); err != nil {
		t.Fatalf("Failed to parse User A response: %v", err)
	}

	roomNameA, ok := respA["room_name"]
	if !ok || roomNameA == "" {
		t.Fatal("User A response missing room_name")
	}
	tokenA, ok := respA["token"]
	if !ok || tokenA == "" {
		t.Fatal("User A response missing token")
	}

	// Verify DB state
	storedSpace, err := realRepo.GetByCanvasSpaceID(canvasID)
	if err != nil {
		t.Fatalf("Failed to query DB for canvas space: %v", err)
	}
	if storedSpace == nil {
		t.Fatal("Canvas space not found in DB after User A request")
	}
	if storedSpace.RoomName != roomNameA {
		t.Errorf("DB room name %s mismatch with response room name %s", storedSpace.RoomName, roomNameA)
	}

	// ----------------------------------------------------------------
	// Step 2: User B connects to SAME Canvas Space "canvas-101"
	// Expected: EXISTING room returned (CreateRoom NOT called again), new token.
	// ----------------------------------------------------------------
	userB := "user-B"
	reqBodyB := CanvasTokenRequest{
		CanvasSpaceID: canvasID,
		Identity:      userB,
	}
	bodyB, _ := json.Marshal(reqBodyB)
	reqB, _ := http.NewRequest("POST", "/canvas-spaces/token", bytes.NewBuffer(bodyB))
	wB := httptest.NewRecorder()

	r.ServeHTTP(wB, reqB)

	if wB.Code != http.StatusOK {
		t.Fatalf("User B request failed with status %d: %s", wB.Code, wB.Body.String())
	}

	var respB map[string]string
	if err := json.Unmarshal(wB.Body.Bytes(), &respB); err != nil {
		t.Fatalf("Failed to parse User B response: %v", err)
	}

	roomNameB, ok := respB["room_name"]
	if !ok {
		t.Fatal("User B response missing room_name")
	}

	// ----------------------------------------------------------------
	// Step 3: Assertions
	// ----------------------------------------------------------------

	// 1. Room names must match
	if roomNameA != roomNameB {
		t.Errorf("User A got room %s, but User B got room %s. They should be the same.", roomNameA, roomNameB)
	}

	// 2. CreateRoom in MockService should be called EXACTLY ONCE
	if createRoomCalled != 1 {
		t.Errorf("Expected CreateRoom to be called 1 time, but was called %d times", createRoomCalled)
	}

	t.Logf("Integration Test Passed: User A and User B joined same room: %s", roomNameA)
}
