package main

import (
	"log"
	"net/http"
    "embed"

	"main/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

    "github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {

	// Database setup
    db, err := gorm.Open(postgres.New(postgres.Config{DSN: service.GetDsnFromEnv()}), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

	// Include existing migration files in the binary
	goose.SetBaseFS(embedMigrations)

	// Goose requires setting dialect first
    if err := goose.SetDialect("postgres"); err != nil {
        log.Fatal(err)
    }
    // Get the underlying *sql.DB
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatal(err)
    }
	// Apply existing migrations to the database
    if err := goose.Up(sqlDB, "migrations"); err != nil {
        log.Fatal(err)
    }

	// Init router structure
 	router := service.SetupRouter(db)

	// Start http server
 	log.Println("Server starting on :80")
 	if err := http.ListenAndServe(":80", router); err != nil {
 	    log.Fatal(err)
 	}
}
