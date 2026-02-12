package main

import (
	"testing"
)

func TestSimulator_Run_Usage(t *testing.T) {
	// Not enough arguments
	code := runSimulator([]string{"simulator"})
	if code != 1 {
		t.Errorf("Expected exit code 1, got %d", code)
	}
}

func TestSimulator_Run_ConnectionFail(t *testing.T) {
	// Should fail connection to invalid host/token
	code := runSimulator([]string{"simulator", "ws://invalid:7880", "invalidtoken"})
	if code != 1 {
		t.Errorf("Expected exit code 1 for connection failure, got %d", code)
	}
}
