package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"space-p2p/db"
	"space-p2p/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type Config struct {
	ApiKey    string
	ApiSecret string
	Host      string
	Port      string
}

var config Config
var roomClient *lksdk.RoomServiceClient
var canvasRepo repository.CanvasSpaceRepository

func main() {
	// Load .env file
	_ = godotenv.Load()

	config = Config{
		ApiKey:    os.Getenv("LIVEKIT_API_KEY"),
		ApiSecret: os.Getenv("LIVEKIT_API_SECRET"),
		Host:      os.Getenv("LIVEKIT_HOST"),
		Port:      os.Getenv("PORT"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	// Initialize Database
	db.InitDB()
	canvasRepo = repository.NewSQLiteCanvasSpaceRepository(db.DB)

	// Initialize RoomServiceClient
	roomClient = lksdk.NewRoomServiceClient(config.Host, config.ApiKey, config.ApiSecret)

	// Initialize Gin
	r := gin.Default()

	// Routes
	r.POST("/rooms", createRoom)
	r.GET("/rooms", listRooms)
	r.POST("/token", generateToken)
	r.DELETE("/rooms/:room", deleteRoom)

	// Canvas Space Routes
	r.POST("/canvas-spaces/token", generateTokenForCanvasSpace)
	r.GET("/canvas-spaces", listCanvasSpaces)
	r.GET("/canvas-spaces/:id", getCanvasSpace)
	r.DELETE("/canvas-spaces/:id", deleteCanvasSpace)

	log.Printf("Server starting on port %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Request/Response types
type CreateRoomRequest struct {
	RoomName     string `json:"room_name" binding:"required"`
	EmptyTimeout uint32 `json:"empty_timeout"` // Seconds
}

type TokenRequest struct {
	RoomName string `json:"room_name" binding:"required"`
	Identity string `json:"identity" binding:"required"`
}

type CanvasTokenRequest struct {
	CanvasSpaceID string `json:"canvas_space_id" binding:"required"`
	Identity      string `json:"identity" binding:"required"`
}

// Handlers
func createRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.EmptyTimeout == 0 {
		req.EmptyTimeout = 600
	}

	room, err := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name:         req.RoomName,
		EmptyTimeout: req.EmptyTimeout,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, room)
}

func listRooms(c *gin.Context) {
	res, err := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res.Rooms)
}

func generateToken(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := createToken(req.RoomName, req.Identity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func deleteRoom(c *gin.Context) {
	roomName := c.Param("room")
	_, err := roomClient.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
}

// Canvas Space Handlers

func generateTokenForCanvasSpace(c *gin.Context) {
	var req CanvasTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if mapping exists
	cs, err := canvasRepo.GetByCanvasSpaceID(req.CanvasSpaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check canvas space: " + err.Error()})
		return
	}

	var roomName string
	if cs != nil {
		roomName = cs.RoomName
	} else {
		// Create new mapping
		newRoomName := "room_" + uuid.New().String()

		// Create LiveKit room ensuring it exists
		_, err := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
			Name:         newRoomName,
			EmptyTimeout: 600, // 10 minutes default
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create LiveKit room: " + err.Error()})
			return
		}

		_, err = canvasRepo.Create(req.CanvasSpaceID, newRoomName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create canvas space mapping: " + err.Error()})
			return
		}
		roomName = newRoomName
	}

	token, err := createToken(roomName, req.Identity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"room_name": roomName,
	})
}

func listCanvasSpaces(c *gin.Context) {
	spaces, err := canvasRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, spaces)
}

func getCanvasSpace(c *gin.Context) {
	id := c.Param("id")
	cs, err := canvasRepo.GetByCanvasSpaceID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cs == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canvas space not found"})
		return
	}
	c.JSON(http.StatusOK, cs)
}

func deleteCanvasSpace(c *gin.Context) {
	id := c.Param("id")

	// Get it first to know the room name (optional: we might want to delete the room too?)
	// For now just deleting the mapping as per "CRUD for canvas-space" requirement.

	err := canvasRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Canvas space mapping deleted"})
}

func createToken(roomName, identity string) (string, error) {
	at := auth.NewAccessToken(config.ApiKey, config.ApiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(time.Hour)

	return at.ToJWT()
}
