package sqlite

import (
	"log"

	"bloodconnect/adapters/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupDatabase(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.UserHealth{},
		&models.UserLocation{},
		&models.Notification{},
		&models.Request{},
		&models.RequestState{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("SQLite database connection established")
	return db, nil
}
