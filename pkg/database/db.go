package database

import (
	"database/sql"
	"fmt"
	"log"

	"cosmos/config"

	_ "github.com/lib/pq"
)

var db *sql.DB // приватная переменная

func Connect(cfg *config.Config) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("не удалось проверить подключение: %v", err)
	}

	log.Println("Подключение к PostgreSQL установлено")
	return nil
}

// GetDB возвращает соединение с БД
func GetDB() *sql.DB {
	if db == nil {
		log.Fatal("БД не подключена! Сначала вызовите Connect()")
	}
	return db
}

func Close() {
	if db != nil {
		db.Close()
		log.Println("Подключение к базе данных закрыто")
	}
}
