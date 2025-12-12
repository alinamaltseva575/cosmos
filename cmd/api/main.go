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
		log.Fatalf("Ошибка подключения к БД: %v", err)
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

	// Авторизация
	http.HandleFunc("/admin/login", h.AdminLoginHandler)
	http.HandleFunc("/admin/logout", h.AdminLogoutHandler)

	// Админ-панель
	http.HandleFunc("/admin", h.AdminDashboardHandler)

	// Планеты
	http.HandleFunc("/admin/planets", h.AdminPlanetsHandler)
	http.HandleFunc("/admin/planets/new", h.AdminNewPlanetHandler)
	http.HandleFunc("/admin/planets/delete/", h.AdminDeletePlanetHandler)
	http.HandleFunc("/admin/planets/edit/", h.AdminEditPlanetHandler)

	// Галактики
	http.HandleFunc("/admin/galaxies", h.AdminGalaxiesHandler)
	http.HandleFunc("/admin/galaxies/new", h.AdminNewGalaxyHandler)
	http.HandleFunc("/admin/galaxies/delete/", h.AdminDeleteGalaxyHandler)
	http.HandleFunc("/admin/galaxies/edit/", h.AdminEditGalaxyHandler)

	// Пользователи
	http.HandleFunc("/admin/users", h.AdminUsersHandler)
	http.HandleFunc("/admin/users/new", h.AdminNewUserHandler)
	http.HandleFunc("/admin/users/delete/", h.AdminDeleteUserHandler)
	http.HandleFunc("/admin/users/edit/", h.AdminEditUserHandler)
	http.HandleFunc("/admin/users/view/", h.AdminUserDetailHandler)

	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Запуск сервера
	log.Printf("Сервер запущен на http://localhost:%s", cfg.AppPort)
	log.Printf("База данных: %s", cfg.DBName)

	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
