package main

import (
	lmLogger "NewListingBot/logger"
	"NewListingBot/middleware"
	"NewListingBot/migrate"
	"NewListingBot/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/skip"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

func main() {
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024, // Set the body limit to 20MB
		Views:     engine,
	})
	// Use the logger middleware
	app.Use(logger.New())

	// first i have to load the .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	lmLogger.InitLogger()
	// Make migrations
	migrate.MigrateDatabase()

	app.Use(skip.New(middleware.CustomHeaderMiddleware(), func(c *fiber.Ctx) bool {
		// Skip the middleware for webhook routes
		return strings.HasPrefix(c.Path(), "/api/v1/webhook")
	}))

	// Register user routes
	routes.HttpRoutes(app)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
