package main

import (
	"fmt"
	"os"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"sharing-vision-backend/internal/config"
)

func main() {
	cfg := config.Load()

	filePath := getenv("MIGRATION_FILE", "migrations/0001_create_posts.sql")
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("read migration file failed: %v", err))
	}

	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		panic(fmt.Sprintf("db connect failed: %v", err))
	}
	defer db.Close()

	if _, err := db.Exec(string(content)); err != nil {
		panic(fmt.Sprintf("migrate failed: %v", err))
	}

	fmt.Println("migration done")
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
