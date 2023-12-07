package database

import (
	"NewListingBot/config"
	"NewListingBot/logger"
	"context"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var cfg config.Config

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
	}
}

func DBConnection() *gorm.DB {
	// This is used to connect to the SQLite database
	connection, err := gorm.Open(sqlite.Open("newlisting.db"), &gorm.Config{})
	if err != nil {
		logger.Error(context.Background(), "Error connecting to SQLite database", zap.Error(err))
		panic("Failed to connect to database")
	}

	return connection
}

func CloseDB() {
	db := DBConnection()
	closeDb, err := db.DB()
	if err != nil {
		logger.Error(context.Background(), "Error closing SQLite Database", zap.Error(err))
	}

	err = closeDb.Close()
	if err != nil {
		logger.Error(context.Background(), "Error closing SQLite Database", zap.Error(err))
	}
}
