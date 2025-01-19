package main

import (
    "log"
    "os"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/valyala/fasthttp/fasthttpadaptor"
    "github.com/Jeff-Barlow-Spady/docker-setup/services/gpio/internal"
)

func main() {
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
    app.Use(logger.New(logger.Config{
        Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
    }))

    gpioManager := internal.NewGPIOManager()
    wsManager := internal.NewWebSocketManager(gpioManager)

    // Metrics endpoint with proper Prometheus handler
    promHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
    app.Get("/metrics", func(c *fiber.Ctx) error {
        promHandler(c.Context())
        return nil
    })

    // GPIO endpoints
    app.Post("/gpio/:pin/setup", func(c *fiber.Ctx) error {
        pin, err := c.ParamsInt("pin")
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid pin")
        }

        direction := c.Query("direction", "out")
        if direction != "in" && direction != "out" {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid direction. Must be 'in' or 'out'")
        }

        if err := gpioManager.SetupPin(pin, direction); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }

        return c.JSON(fiber.Map{
            "status": "success",
            "message": "Pin configured",
            "pin": pin,
            "direction": direction,
        })
    })

    app.Post("/gpio/:pin/write", func(c *fiber.Ctx) error {
        pin, err := c.ParamsInt("pin")
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid pin")
        }

        var req struct {
            Value bool `json:"value"`
        }

        if err := c.BodyParser(&req); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
        }

        if err := gpioManager.WritePin(pin, req.Value); err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }

        return c.JSON(fiber.Map{
            "status": "success",
            "pin": pin,
            "value": req.Value,
        })
    })

    app.Get("/gpio/:pin/read", func(c *fiber.Ctx) error {
        pin, err := c.ParamsInt("pin")
        if err != nil {
            return fiber.NewError(fiber.StatusBadRequest, "Invalid pin")
        }

        value, err := gpioManager.ReadPin(pin)
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, err.Error())
        }

        return c.JSON(fiber.Map{
            "status": "success",
            "pin": pin,
            "value": value,
        })
    })

    // WebSocket endpoint
    app.Get("/ws/gpio", wsManager.HandleWebSocket)

    // Health check endpoint
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "healthy",
            "service": "gpio",
        })
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }

    log.Printf("Starting GPIO service on port %s", port)
    log.Fatal(app.Listen(":" + port))
}
