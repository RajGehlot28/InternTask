package config

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// Simple .env loader
func loadEnv() {
	files := []string{".env", "backend/.env", "../.env"}
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if os.Getenv(key) == "" {
					os.Setenv(key, val)
				}
			}
		}
		file.Close()
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// Connect to PostgreSQL database
func InitDB() error {
	loadEnv()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		pgHost := getEnv("POSTGRES_HOST", "localhost")
		pgUser := getEnv("POSTGRES_USER", "postgres")
		pgPass := getEnv("POSTGRES_PASSWORD", "postgres")
		pgDB := getEnv("POSTGRES_DB", "ticketdb")
		pgPort := getEnv("POSTGRES_PORT", "5432")
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pgUser, pgPass, pgHost, pgPort, pgDB)
	}

	fmt.Println("Connecting to PostgreSQL database...")

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	fmt.Println("Successfully connected to PostgreSQL database!")
	return createTables()
}

// Create database tables
func createTables() error {
	usersSQL := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	ticketsSQL := `CREATE TABLE IF NOT EXISTS tickets (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		status VARCHAR(50) NOT NULL DEFAULT 'open',
		user_id INT NOT NULL REFERENCES users(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := DB.Exec(usersSQL); err != nil {
		return err
	}
	_, err := DB.Exec(ticketsSQL)
	return err
}
