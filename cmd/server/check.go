package main

import (
	"NewListingBot/config"
	"NewListingBot/exchange"
	"NewListingBot/logger"
	"NewListingBot/models"
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	ctx := context.Background()

	// Sample order
	order := models.Order{
		Symbol:   strPtr("BTCUSDT"),
		Price:    float64Ptr(100.0),  // Sample buying price
		Quantity: float64Ptr(0.0030), // Sample quantity
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Error(context.Background(), "error loading config on scheduler", zap.Error(err))
	}

	// Mock MEXCExchange instance
	mexc := exchange.NewMXCExchange(cfg) // Replace with your actual implementation

	// Check if profit is available with a target percentage of 5%
	targetPercentage := 5.0
	available, err := models.IsProfitAvailable(ctx, *mexc, order, targetPercentage)
	if err != nil {
		logger.Error(ctx, "Error checking profit availability", zap.Error(err))
		return
	}

	if available {
		fmt.Printf("Profit of %f%% or more is available!\n", targetPercentage)
	} else {
		fmt.Printf("No profit of %f%% or more available.\n", targetPercentage)
	}

	isAfter := time.Now().Before(time.Now().Add(-time.Second * 5))
	log.Println(isAfter)
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
