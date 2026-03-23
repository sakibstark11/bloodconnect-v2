package postgres

import (
	"log"

	"bloodconnect/adapters/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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
	)
	if err != nil {
		return nil, err
	}

	log.Println("Postgres database connection established")
	return db, nil
}
