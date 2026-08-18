package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := getenv("DB_DSN", "root:root@tcp(127.0.0.1:3306)/article?charset=utf8mb4&parseTime=True&loc=Local")
	if dsn == "" {
		panic("DB_DSN is required")
	}

	filePath := getenv("MIGRATION_FILE", "migrations/0001_create_posts.sql")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("read migration file error: %v", err))
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("mysql connect error: %v", err))
	}
	defer db.Close()

	if _, err := db.Exec(string(content)); err != nil {
		panic(fmt.Sprintf("migration failed: %v", err))
	}

	fmt.Println("Migration finished")
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
