package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: simulator <host> <token>")
		os.Exit(1)
	}

	host := os.Args[1]
	token := os.Args[2]

	fmt.Printf("Attempting to connect to %s with provided token...\n", host)
	room, err := lksdk.ConnectToRoomWithToken(host, token, lksdk.NewRoomCallback())
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer room.Disconnect()

	fmt.Printf("Connected successfully to room: %s\n", room.Name())

	// Create a channel to handle OS signals or timeout
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sig:
		fmt.Println("Received signal, disconnecting...")
	case <-time.After(30 * time.Second):
		fmt.Println("Simulation timeout reached, disconnecting...")
	}
}
