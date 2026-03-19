package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

// Connect устанавливает соединение с базой данных PostgreSQL
func Connect() error {
	var dsn string

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	} else {
		// Сборка DSN из отдельных переменных (fallback)
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname,
		)
	}

	var err error
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		log.Printf("Попытка подключения к БД (%d/%d)...", i+1, maxRetries)
		DB, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			// Проверка соединения
			if err = DB.Ping(); err == nil {
				log.Println(" Подключение к PostgreSQL установлено")
				return nil
			}
		}

		log.Printf("Ошибка подключения: %v. Ждем 3 сек...", err)
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("не удалось подключиться к БД после %d попыток: %w", maxRetries, err)
}

// Close закрывает соединение с базой данных
func Close() {
	if DB != nil {
		DB.Close()
	}
}
