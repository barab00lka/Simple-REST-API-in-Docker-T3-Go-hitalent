package service

import (
	"fmt"
	"os"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetDsnFromEnv() string {
	host := "db"
    port := "5432" 
    user := os.Getenv("POSTGRES_USER")
    password := os.Getenv("POSTGRES_PASSWORD")
    dbname := os.Getenv("POSTGRES_DB")
    sslmode := os.Getenv("DB_SSLMODE")
    if sslmode == "" {
        sslmode = "disable"
    }

    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        host, user, password, dbname, port, sslmode,
    )
    return dsn
}

func getQueryInt(r *http.Request, key string, defaultVal, min, max int) int {
	val := defaultVal
	if s := r.URL.Query().Get(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			val = v
		}
	}
	if val < min {
		val = min
	}
	if max > 0 && val > max {
		val = max
	}
	return val
}

func getQueryBool(r *http.Request, key string, defaultVal bool) bool {
	if s := r.URL.Query().Get(key); s != "" {
		return s == "true"
	}
	return defaultVal
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{"error": message})
}
