package sqlite

import (
	"log"

	"bloodconnect/adapters/sqlite/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupDatabase initializes the SQLite database and runs auto-migrations.
func SetupDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.UserHealth{},
		&models.UserLocation{},
		&models.Notification{},
		&models.Request{},
		&models.RequestState{},
		&models.Job{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established")
	return db, nil
}
