package main

import (
	"NewListingBot/config"
	"NewListingBot/exchange"
	"NewListingBot/logger"
	"context"
	"go.uber.org/zap"
	"log"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		logger.Error(context.Background(), "error loading config on scheduler", zap.Error(err))
	}

	// Mock MEXCExchange instance
	mexc := exchange.NewMXCExchange(cfg) // Replace with your actual implementation

	buy, err := mexc.Buy("KASUSDT", 100)
	if err != nil {
		return
	}
	log.Println(buy)
}

// Helper functions for creating pointers
func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
