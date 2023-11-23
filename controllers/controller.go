package controllers

import (
	"NewListingBot/database"
	"NewListingBot/models"
	"context"
	"github.com/gofiber/fiber/v2"
	"time"
)

func ListOrdersController(c *fiber.Ctx) error {
	var orders []models.Order

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Open the database connection
	db := database.DBConnection()
	defer database.CloseDB()

	err := db.WithContext(ctx).Model(&models.Order{}).Find(&orders).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching orders")
	}

	return c.Render("views/list_orders", fiber.Map{
		"Orders": orders,
	})
}
