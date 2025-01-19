package internal

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"periph.io/x/conn/v3/gpio"
)

// InitializeCleanup sets up signal handlers for graceful shutdown
func (gm *GPIOManager) InitializeCleanup() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Cleaning up GPIO pins...")

		gm.mu.Lock()
		defer gm.mu.Unlock()

		// Set all output pins to low
		for _, state := range gm.pins {
			if state.direction == "out" {
				if err := state.pin.Out(gpio.Low); err != nil {
					log.Printf("Error setting pin to low during cleanup: %v", err)
				}
			}
		}

		log.Println("GPIO cleanup complete")
		os.Exit(0)
	}()
}
