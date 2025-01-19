package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jeff-Barlow-Spady/docker-setup/services/metrics/internal"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	collector := internal.NewMetricsCollector()

	// Get update interval from environment
	intervalStr := os.Getenv("METRICS_UPDATE_INTERVAL")
	updateInterval := 15 * time.Second
	if intervalStr != "" {
		if parsed, err := time.ParseDuration(intervalStr); err == nil {
			updateInterval = parsed
		}
	}

	// Update metrics periodically with context
	go func() {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping metrics collection")
				return
			case <-ticker.C:
				if err := collector.UpdateMetrics(); err != nil {
					log.Printf("Error updating metrics: %v", err)
				}
			}
		}
	}()

	app.Get("/metrics", func(c *fiber.Ctx) error {
		metrics := collector.GetMetrics()
		return c.JSON(metrics)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		health := collector.GetHealth()
		status := fiber.StatusOK
		if health.Status == "degraded" {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(health)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting metrics service on port %s with update interval %s", port, updateInterval)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
