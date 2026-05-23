package migrations

import (
	"embed"
	"log"
	"os"
	"testing"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	. "main/internal/models"
)

//go:embed test_migrations/*.sql
var testMigrations embed.FS

var db *gorm.DB

func TestMain(m *testing.M) {
dsn := "host=test-db user=test password=test dbname=test-db port=5432 sslmode=disable"
	// Database setup
    db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
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

	// 3. Optional: clean up previous state (Down all migrations)
	if err := goose.DownTo(sqlDB, "test_migrations", 0); err != nil {
		log.Printf("goose DownTo (cleanup) failed (ignored): %v", err)
	}

	// 4. Apply all test migrations (Up)
	if err := goose.Up(sqlDB, "test_migrations"); err != nil {
		log.Fatalf("goose Up failed: %v", err)
	}
	log.Println("Test migrations applied successfully")

	// 5. Run tests
	code := m.Run()

	// 6. Optionally clean up after tests
	if err := goose.DownTo(sqlDB, "test_migrations", 0); err != nil {
		log.Printf("goose DownTo after tests failed (ignored): %v", err)
	}

	os.Exit(code)
}

func TestExample(t *testing.T) {
	var count int64
	if err := db.Model(&Department{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("Total departments: %d", count)

	if err := db.Model(&Employee{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("Total employees: %d", count)
}
