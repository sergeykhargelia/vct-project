package database

import (
	"fmt"
	"os"

	"github.com/sergeykhargelia/vct-project/backend/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func InitDB() (*gorm.DB, error) {
	user := getEnvOrDefault("PGUSER", "postgres")
	password := getEnvOrDefault("PGPASSWORD", "postgres")
	host := getEnvOrDefault("PGHOST", "localhost")
	port := getEnvOrDefault("PGPORT", "5432")
	dbname := getEnvOrDefault("PGDATABASE", "regular_expenses_tracker")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.RegularExpense{},
		&model.Expense{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate tables: %w", err)
	}

	return db, nil
}
