package database

import (
	"NewListingBot/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var cfg config.Config

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
	}
}

func DBConnection() *gorm.DB {
	/*
		This is used to connect to the database, and if the database has issues, i panic on the console
	*/

	connectionString := fmt.Sprintf(" host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.PostgresHost, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDatabaseName, cfg.PostgresPort)

	connection, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to database", err)
	}

	return connection
}

func CloseDB() {

	db := DBConnection()
	closeDb, err := db.DB()
	if err != nil {
		log.Panicln("Error closing database connection ", err)
	}

	err = closeDb.Close()
	if err != nil {
		log.Panicln(err)
	}
}
