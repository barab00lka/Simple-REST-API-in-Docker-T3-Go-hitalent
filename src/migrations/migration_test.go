package migrations

import (
	"embed"
	"log"
	"os"
	"testing"

	"main/internal/service"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// . "main/internal/models"
)

//go:embed test-migration/*.sql
var testMigrations embed.FS

var db *gorm.DB

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestMigration(m *testing.T) {
    db, err := gorm.Open(postgres.New(postgres.Config{DSN: service.GetDsnFromEnv()}), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Set up goose with embedded migrations
	goose.SetBaseFS(testMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	// Apply all test migrations (Up)
	if err := goose.Up(sqlDB, "test-migration"); err != nil {
		log.Fatalf("goose Up failed: %v", err)
	}
	log.Println("Test migrations applied successfully")


	// 6. Optionally clean up after tests
	if err := goose.DownTo(sqlDB, "test-migration", 1); err != nil {
		log.Printf("goose DownTo after tests failed (ignored): %v", err)
	}

}

// func TestExample(t *testing.T) {
// 	var count int64
// 	if err := db.Model(&Department{}).Count(&count).Error; err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Logf("Total departments: %d", count)

// 	if err := db.Model(&Employee{}).Count(&count).Error; err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Logf("Total employees: %d", count)
// }
