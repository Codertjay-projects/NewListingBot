package middleware

import (
	"NewListingBot/config"
	"NewListingBot/logger"
	"context"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"time"
)

var cfg config.Config

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		logger.Error(context.Background(), "Error loading config on middleware", zap.Error(err))
	}
}
func CustomHeaderMiddleware() func(*fiber.Ctx) error {
	/* This header is used to check if the user passed the header value and key or not*/

	return func(c *fiber.Ctx) error {
		if c.Get("NEW_LISTING_SK_HEADER") == cfg.NewListingSKHeader {
			return c.Next()
		}

		return c.Status(401).JSON(fiber.Map{"error": "Custom headers were not provided."})
	}
}

// LimitMiddleware Define the middleware function
func LimitMiddleware(c *fiber.Ctx) error {
	limitTime := 1 * time.Minute
	// currently, the user only has 50 requests per minute
	var limiter = rate.NewLimiter(50, int(limitTime))

	// Check if the IP address has exceeded the request limit
	if limiter.Allow() == false {
		// If so, return a 429 (Too Many Requests) status code
		return c.SendStatus(fiber.StatusTooManyRequests)
	}

	// If the request limit has not been exceeded, call the next middleware handler
	return c.Next()
}
