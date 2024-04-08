package main

import (
	"go-admin/src/database"
	"go-admin/src/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	database.Connect()
	database.AutoMigrate()
	database.SetupRedis()
	database.SetupCacheChannel()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost",
		AllowCredentials: true,
	}))

	routes.Setup(app)

	app.Listen(":8080")
}
