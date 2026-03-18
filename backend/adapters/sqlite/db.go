package sqlite

import (
	"log"

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

	// Run Migrations
	err = db.AutoMigrate(
		&userModel{},
		&userHealthModel{},
		&userLocationModel{},
		&notificationModel{},
		&requestModel{},
		&requestStateModel{},
		&jobModel{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established")

	return db, nil
}
