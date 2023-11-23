package migrate

import (
	"NewListingBot/database"
	"NewListingBot/models"
	"log"
)

// MigrateDatabase performs database migrations for all models
func MigrateDatabase() {
	// Migrate the schema
	db := database.DBConnection()
	// close the database after the connection
	defer database.CloseDB()

	// Migrate the models
	err := db.AutoMigrate(
		&models.Order{},
	)
	if err != nil {
		log.Println(err)
	}

}
