	package main

	import (
	    "context"
	    "github.com/gofiber/fiber/v2"
	    "github.com/rs/zerolog"
	    "os"
	    "os/signal"
	    "sync"
	    "syscall"
	    "time"

	    "internal/gpio"
	    "cmd/gpiosvc/handlers"
	    "cmd/gpiosvc/middleware"
	)

	func main() {
	    // Initialize logger
	    log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	    // Initialize GPIO service
	    gpioService := gpio.NewService()

	    // Create Fiber app with custom config
	    app := fiber.New(fiber.Config{
	        ReadTimeout:  10 * time.Second,
	        WriteTimeout: 10 * time.Second,
	        IdleTimeout:  120 * time.Second,
	        ErrorHandler: func(c *fiber.Ctx, err error) error {
	          

