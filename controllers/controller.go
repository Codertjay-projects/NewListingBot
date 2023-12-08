package controllers

import (
	"NewListingBot/adapters"
	"NewListingBot/config"
	"NewListingBot/database"
	"NewListingBot/exchange"
	"NewListingBot/models"
	"NewListingBot/serializers"
	"context"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
	"time"
)

type Response struct {
	Message any  `json:"message,omitempty"`
	Success bool `json:"success,omitempty"`
	Error   any  `json:"error,omitempty"`
	Errors  any  `json:"errors,omitempty"`
	Detail  any  `json:"detail,omitempty"` // used by the backend developer
}

func OrderListController(c *fiber.Ctx) error {
	var orders []models.Order

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Open the database connection
	db := database.DBConnection()
	defer database.CloseDB()

	err := db.WithContext(ctx).Model(&models.Order{}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "timestamp"}, Desc: true}).Find(&orders).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching orders")
	}

	return c.Status(200).JSON(orders)
}

func OrderCreateController(c *fiber.Ctx) error {
	var order models.Order
	var requestBody serializers.OrderCreateRequestSerializer

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Open the database connection
	db := database.DBConnection()
	defer database.CloseDB()

	validateAdapter := adapters.NewValidate()

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{Message: "Invalid request body ", Success: false, Detail: err.Error()})
	}

	vErr := validateAdapter.ValidateData(&requestBody)
	if vErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Response{Errors: vErr, Success: false, Detail: vErr})
	}

	scheduleTime := requestBody.ScheduleTime
	scheduleSellTime := scheduleTime.Add(time.Minute * 1) // add 15 minutes for ScheduleSellTime

	order = models.Order{
		Symbol:           requestBody.Symbol,
		ScheduleTime:     scheduleTime,
		ScheduleSellTime: &scheduleSellTime, // start the sell
		Price:            requestBody.Price,
	}

	err := db.WithContext(ctx).Model(&models.Order{}).Create(&order).Error
	if err != nil {
		return c.Status(400).JSON(Response{Errors: err.Error(), Success: false, Detail: err.Error()})
	}

	// make the schedule for all buy operations
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*5))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*5))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*5))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*4))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*4))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*4))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*4))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*3))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*3))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*3))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*2))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*2))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*2))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*1))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*1))
	order.ScheduleBuyScheduler(ctx, db, order.ID, order.ScheduleTime.Add(-time.Second*1))

	// make the schedule for all sell operations
	order.ScheduleSellScheduler(ctx, db)

	return c.Status(200).JSON(order)
}

func GetMarketDataController(c *fiber.Ctx) error {

	cfg, err := config.Load()
	if err != nil {
		return c.Status(400).JSON(Response{Message: err.Error(), Success: false})
	}
	mexc := exchange.NewMXCExchange(cfg)

	token := c.Query("token")

	marketData, err := mexc.GetMarketData()
	if err != nil {
		return c.Status(400).JSON(Response{Message: err.Error(), Success: false})
	}

	if token != "" {
		for index, symbol := range marketData.Symbols {

			if symbol.BaseAsset == token {
				return c.Status(200).JSON(marketData.Symbols[index])
			}
		}
	}

	return c.Status(200).JSON(marketData)
}
