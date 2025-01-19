package main

import (
    "log"
    "os"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/Jeff-Barlow-Spady/docker-setup/services/auths/internal"
)

func main() {
    app := fiber.New()
    app.Use(logger.New())

    authService := internal.NewAuthService()

    app.Post("/auths/register", func(c *fiber.Ctx) error {
        var req struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }

        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{
                "error": "Invalid request",
            })
        }

        if success := authService.CreateUser(req.Username, req.Password); !success {
            return c.Status(409).JSON(fiber.Map{
                "error": "User already exists",
            })
        }

        return c.Status(201).JSON(fiber.Map{
            "message": "User created successfully",
        })
    })

    app.Post("/auths/login", func(c *fiber.Ctx) error {
        var req struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }

        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{
                "error": "Invalid request",
            })
        }

        if !authService.VerifyUser(req.Username, req.Password) {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid credentials",
            })
        }

        token := authService.CreateToken(req.Username)
        return c.JSON(fiber.Map{
            "token": token,
        })
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Fatal(app.Listen(":" + port))
}
