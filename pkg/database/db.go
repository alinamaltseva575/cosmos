package database

import (
	"database/sql"
	"fmt"
	"log"

	"cosmos/config"

	_ "github.com/lib/pq"
)

// DB - структура для работы с БД (вместо глобальной переменной)
type DB struct {
	conn *sql.DB
}

// NewDB создает новое подключение к БД
func NewDB(cfg *config.Config) (*DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось проверить подключение: %v", err)
	}

	log.Println("✅ Подключение к PostgreSQL установлено")

	return &DB{conn: conn}, nil
}

// GetDB возвращает sql.DB (для совместимости со старым кодом)
func (d *DB) GetDB() *sql.DB {
	return d.conn
}

// Close закрывает подключение
func (d *DB) Close() error {
	if d.conn != nil {
		log.Println("Подключение к базе данных закрыто")
		return d.conn.Close()
	}
	return nil
}

// Ping проверяет соединение
func (d *DB) Ping() error {
	return d.conn.Ping()
}
