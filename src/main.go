package main

import (
	"log"
	"net/http"

	"main/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=postgres_user password=t3-go dbname=postgres_db port=5432 sslmode=disable TimeZone=Europe/Moscow",
	}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	h := &service.Handler{DB: db}
	http.Handle("/departments", h)
	http.Handle("/departments/", h)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
