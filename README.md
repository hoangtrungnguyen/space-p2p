<<<<<<< HEAD
# space-p2p
=======
# LiveKit P2P Session Manager (Go)

This project provides a Go-based backend to manage LiveKit rooms and generate access tokens (JWTs) for peer-to-peer sessions using the LiveKit SFU.

## Prerequisites

- [Go](https://go.dev/) (Installed)
- [LiveKit Server](https://docs.livekit.io/realtime/self-host/install/) (Running locally or on LiveKit Cloud)

## Setup

1. **Configure Environment**:
   Update the `.env` file with your LiveKit credentials:
   ```env
   LIVEKIT_API_KEY=your_key
   LIVEKIT_API_SECRET=your_secret
   LIVEKIT_HOST=https://your-livekit-instance.livekit.cloud
   PORT=8080
   ```

2. **Run the Server**:
   ```bash
   go run main.go
   ```

## API Endpoints

### 1. Create a Room
**POST** `/rooms`
```json
{
  "room_name": "my-awesome-room",
  "empty_timeout": 600
}
```

### 2. List Active Rooms
**GET** `/rooms`

### 3. Generate Join Token
**POST** `/token`
```json
{
  "room_name": "my-awesome-room",
  "identity": "user-unique-id"
}
```
*Response*:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 4. Delete a Room
**DELETE** `/rooms/:room_name`

## Canvas Space API

### 1. Generate Token (with Auto-Room Creation)
**POST** `/canvas-spaces/token`
Maps a `canvas_space_id` to a LiveKit room. If no room is mapped, creates a new one.
```json
{
  "canvas_space_id": "space-123",
  "identity": "user-123"
}
```
*Response*:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "room_name": "room_uuid_..."
}
```

### 2. List Canvas Spaces
**GET** `/canvas-spaces`

### 3. Get Canvas Space Details
**GET** `/canvas-spaces/:id`

### 4. Delete Canvas Space Mapping
**DELETE** `/canvas-spaces/:id`

## Architecture

This server acts as a **Signaling & Management** layer.
- Clients request tokens from this server.
- Using the returned JWT, clients connect directly to the LiveKit SFU (WebRTC).
- Media and data are then synchronized between peers via the LiveKit protocol.
- **SQLite** is used to persist mappings between `canvas_space_id` and LiveKit `room_name`.

>>>>>>> 8842f0e (feat: implementation Canvas Space CRUD with SQLite)
