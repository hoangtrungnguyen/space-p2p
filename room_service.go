package main

import (
	"context"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type RoomService interface {
	CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error)
	ListRooms(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error)
	DeleteRoom(ctx context.Context, req *livekit.DeleteRoomRequest) (*livekit.DeleteRoomResponse, error)
	ListParticipants(ctx context.Context, req *livekit.ListParticipantsRequest) (*livekit.ListParticipantsResponse, error)
}

type LiveKitRoomService struct {
	client *lksdk.RoomServiceClient
}

func NewLiveKitRoomService(client *lksdk.RoomServiceClient) *LiveKitRoomService {
	return &LiveKitRoomService{client: client}
}

func (s *LiveKitRoomService) CreateRoom(ctx context.Context, req *livekit.CreateRoomRequest) (*livekit.Room, error) {
	return s.client.CreateRoom(ctx, req)
}

func (s *LiveKitRoomService) ListRooms(ctx context.Context, req *livekit.ListRoomsRequest) (*livekit.ListRoomsResponse, error) {
	return s.client.ListRooms(ctx, req)
}

func (s *LiveKitRoomService) DeleteRoom(ctx context.Context, req *livekit.DeleteRoomRequest) (*livekit.DeleteRoomResponse, error) {
	return s.client.DeleteRoom(ctx, req)
}

func (s *LiveKitRoomService) ListParticipants(ctx context.Context, req *livekit.ListParticipantsRequest) (*livekit.ListParticipantsResponse, error) {
	return s.client.ListParticipants(ctx, req)
}
