	package gpio

	import (
	  "encoding/json"
	  "errors"
	  "time"
	)

	// Direction represents GPIO pin direction (input/output)
	type Direction int

	const (
	  Input Direction = iota
	  Output
	)

	// State represents HIGH/LOW pin state
	type State bool

	const (
	  Low  State = false
	  High State = true
	)

	// Pin represents a GPIO pin configuration
	type Pin struct {
	  Number    int       `json:"number" validate:"required,min=0,max=40"`
	  Direction Direction `json:"direction" validate:"required"`
	  State     State     `json:"state,omitempty"`
	  PullUp    bool      `json:"pull_up,omitempty"`
	}

	// Event represents a GPIO pin state change event
	type Event struct {
	  Pin       int       `json:"pin"`
	  State     State     `json:"state"`
	  Timestamp time.Time `json:"timestamp"`
	}

	// ValidationError represents pin validation errors
	type ValidationError struct {
	  Field string
	  Msg   string
	}

	func (e *ValidationError) Error() string {
	  return e.Field + ": " + e.Msg
	}

