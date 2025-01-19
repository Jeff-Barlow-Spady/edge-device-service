	package main

	import (
	    "os"
	    "github.com/gofiber/fiber/v2"
	    "github.com/gofiber/fiber/v2/middleware/logger"
	    "github.com/gofiber/fiber/v2/middleware/recover"
	    "github.com/rs/zerolog"
	    "github.com/prometheus/client_golang/prometheus/promhttp"
	    "github.com/valyala/fasthttp/fasthttpadaptor"
	    "github.com/Jeff-Barlow-Spady/edge-device-service/internal/gpio"
	)

	func main() {
	    // Initialize logger
	    log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	    // Initialize Fiber app with custom config
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
	    })

	    app.Use(recover.New())
	    app.Use(logger.New())

	    // Initialize GPIO manager
	    gpioManager := gpio.NewGPIOManager()
	    wsManager := gpio.NewWebSocketManager(gpioManager)

	    // Set up routes, including metrics endpoint
	    setupRoutes(app, gpioManager, wsManager)

	    // Start server
	    port := os.Getenv("PORT")
	    if port == "" {
	        port = "8000"
	    }

	    log.Info().Msgf("Starting GPIO service on port %s", port)
	    log.Fatal().Err(app.Listen(":" + port)).Msg("Server stopped")
	}

	func setupRoutes(app *fiber.App, gpioManager *gpio.GPIOManager, wsManager *gpio.WebSocketManager) {
	    // Metrics endpoint
	    promHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	    app.Get("/metrics", func(c *fiber.Ctx) error {
	        promHandler(c.Context())
	        return nil
	    })

	    // Health check
	    app.Get("/health", func(c *fiber.Ctx) error {
	        return c.JSON(fiber.Map{
	            "status": "healthy",
	            "service": "gpio",
	        })
	    })

	    // GPIO endpoints
	    app.Post("/gpio/:pin/setup", handleGPIOSetup(gpioManager))
	    app.Post("/gpio/:pin/write", handleGPIOWrite(gpioManager))
	    app.Get("/gpio/:pin/read", handleGPIORead(gpioManager))

	    // WebSocket endpoint
	    app.Get("/ws/gpio", wsManager.HandleWebSocket)
	}
