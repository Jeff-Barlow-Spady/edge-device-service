package internal

import (
	"testing"
)

func TestNewGPIOManager(t *testing.T) {
	manager := NewGPIOManager()
	if manager == nil {
		t.Error("Expected non-nil GPIO manager")
	}
}

func TestPinOperations(t *testing.T) {
	manager := NewGPIOManager()

	// Test pin setup
	err := manager.SetupPin(18, "out")
	if err != nil {
		t.Logf("SetupPin error (expected on non-Pi hardware): %v", err)
	}

	// Test pin write
	err = manager.WritePin(18, true)
	if err != nil {
		t.Logf("WritePin error (expected on non-Pi hardware): %v", err)
	}

	// Test pin read
	_, err = manager.ReadPin(18)
	if err != nil {
		t.Logf("ReadPin error (expected on non-Pi hardware): %v", err)
	}
}

func TestWebSocketManager(t *testing.T) {
	gpioManager := NewGPIOManager()
	wsManager := NewWebSocketManager(gpioManager)
	if wsManager == nil {
		t.Error("Expected non-nil WebSocket manager")
	}
}
