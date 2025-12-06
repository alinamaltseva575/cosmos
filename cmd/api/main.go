package main

import (
	"log"
	"net/http"

	"cosmos/config"
	"cosmos/internal/handler"
	"cosmos/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к БД
	err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer database.Close()

	// Получаем соединение через функцию
	db := database.GetDB()

	// Создаем обработчик
	h := handler.NewHandler(db)

	// Настраиваем маршруты
	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/planets", h.PlanetsHandler)
	http.HandleFunc("/planets/", h.PlanetDetailHandler)
	http.HandleFunc("/galaxies", h.GalaxiesHandler)
	http.HandleFunc("/galaxies/", h.GalaxyDetailHandler)

	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Запуск сервера
	log.Printf("Сервер запущен на http://localhost:%s", cfg.AppPort)
	log.Printf("База данных: %s", cfg.DBName)

	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}
