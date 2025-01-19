package internal

import (
	"fmt"
	"log"
	"sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// GPIOCallback is a function type for GPIO state change callbacks
type GPIOCallback func(pin int, value bool)

// gpioOperations represents the possible operations on a GPIO pin
type gpioOperations struct {
	direction string
	value     bool
}

// gpioState represents the current state of a GPIO pin
type gpioState struct {
	pin       gpio.PinIO
	direction string
	value     bool
}

// GPIOManager manages GPIO pins and their states
type GPIOManager struct {
	pins      map[int]*gpioState
	callbacks []GPIOCallback
	mu        sync.RWMutex
}

// NewGPIOManager creates a new GPIO manager
func NewGPIOManager() *GPIOManager {
	// Initialize host
	if _, err := host.Init(); err != nil {
		log.Printf("Failed to initialize host: %v", err)
	}

	return &GPIOManager{
		pins:      make(map[int]*gpioState),
		callbacks: make([]GPIOCallback, 0),
	}
}

// SetupPin configures a GPIO pin with the specified direction
func (gm *GPIOManager) SetupPin(pinNumber int, direction string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Get the GPIO pin
	pin := gpioreg.ByName(fmt.Sprintf("GPIO%d", pinNumber))
	if pin == nil {
		return fmt.Errorf("failed to find pin %d", pinNumber)
	}

	var err error
	switch direction {
	case "in":
		err = pin.In(gpio.PullUp, gpio.NoEdge)
	case "out":
		err = pin.Out(gpio.Low)
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}

	if err != nil {
		return fmt.Errorf("failed to set pin direction: %v", err)
	}

	gm.pins[pinNumber] = &gpioState{
		pin:       pin,
		direction: direction,
		value:     false,
	}

	return nil
}

// WritePin sets the value of a GPIO pin
func (gm *GPIOManager) WritePin(pinNumber int, value bool) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	state, exists := gm.pins[pinNumber]
	if !exists {
		return fmt.Errorf("pin %d not configured", pinNumber)
	}

	if state.direction != "out" {
		return fmt.Errorf("pin %d not configured for output", pinNumber)
	}

	level := gpio.Low
	if value {
		level = gpio.High
	}

	if err := state.pin.Out(level); err != nil {
		return fmt.Errorf("failed to set pin value: %v", err)
	}

	state.value = value
	gm.notifyCallbacks(pinNumber, value)
	return nil
}

// ReadPin reads the current value of a GPIO pin
func (gm *GPIOManager) ReadPin(pinNumber int) (bool, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	state, exists := gm.pins[pinNumber]
	if !exists {
		return false, fmt.Errorf("pin %d not configured", pinNumber)
	}

	if state.direction != "in" {
		return false, fmt.Errorf("pin %d not configured for input", pinNumber)
	}

	return state.pin.Read() == gpio.High, nil
}

// RegisterCallback registers a callback function for GPIO state changes
func (gm *GPIOManager) RegisterCallback(callback GPIOCallback) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	gm.callbacks = append(gm.callbacks, callback)
}

// notifyCallbacks notifies all registered callbacks of a GPIO state change
func (gm *GPIOManager) notifyCallbacks(pin int, value bool) {
	for _, callback := range gm.callbacks {
		go callback(pin, value)
	}
}

// boolToFloat64 converts a boolean to a float64 (1.0 for true, 0.0 for false)
func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
