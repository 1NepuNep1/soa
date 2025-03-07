package tests

import (
	"os"
	"testing"
	"userservice/database"
	"userservice/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}

	database.DB = db
	database.DB.AutoMigrate(&models.User{}, &models.UserProfile{}, &models.Session{})
}

func CleanupTestDB() {
	sqlDB, err := database.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func TestMain(m *testing.M) {
	SetupTestDB()
	exitCode := m.Run()
	CleanupTestDB()
	os.Exit(exitCode)
}
