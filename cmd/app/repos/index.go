package repos

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Driver struct {
	db *gorm.DB
}

func (driver *Driver) migrate() error {
	// Migrar os esquemas
	return driver.db.AutoMigrate(
		&Virtual{},
		&Device{},
	)

}

func NewDB() (*Driver, error) {
	dsn := os.Getenv("DB_URL")

	dbDriver, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	//	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	db := Driver{
		db: dbDriver,
	}
	//db.migrate()
	if os.Getenv("DB_MIGRITE") == "" {
		db.migrate()
	}
	//
	return &db, nil

}
