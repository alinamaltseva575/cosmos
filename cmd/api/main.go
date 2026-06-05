package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"cosmos/config"
	"cosmos/internal/handler"
	"cosmos/internal/repository"
	"cosmos/internal/service"
	"cosmos/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	cfg := config.Load()

	dbConn, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer dbConn.Close()

	db := dbConn.GetDB()

	// Создаем репозитории
	planetRepo := repository.NewPlanetRepository(db)
	galaxyRepo := repository.NewGalaxyRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Создаем сервисы
	planetService := service.NewPlanetService(planetRepo, galaxyRepo)
	galaxyService := service.NewGalaxyService(galaxyRepo)
	userService := service.NewUserService(userRepo)

	// Функции для шаблонов
	funcMap := template.FuncMap{
		"formatNumber": func(num float64) string {
			if num == 0 {
				return "0"
			}
			if num >= 1e12 {
				return fmt.Sprintf("%.1f трлн", num/1e12)
			}
			if num >= 1e9 {
				return fmt.Sprintf("%.1f млрд", num/1e9)
			}
			if num >= 1e6 {
				return fmt.Sprintf("%.1f млн", num/1e6)
			}
			if num >= 1e3 {
				return fmt.Sprintf("%.0f тыс", num/1e3)
			}
			return fmt.Sprintf("%.0f", num)
		},
		"formatMass": func(mass float64) string {
			if mass == 0 {
				return "0 кг"
			}
			if mass >= 1e24 {
				return fmt.Sprintf("%.2f ×10²⁴ кг", mass/1e24)
			}
			return fmt.Sprintf("%.0f кг", mass)
		},
		"derefInt": func(p interface{}) int {
			if p == nil {
				return 0
			}
			switch v := p.(type) {
			case *int:
				if v != nil {
					return *v
				}
			case int:
				return v
			}
			return 0
		},
		"derefFloat": func(p interface{}) float64 {
			if p == nil {
				return 0
			}
			switch v := p.(type) {
			case *float64:
				if v != nil {
					return *v
				}
			case float64:
				return v
			}
			return 0
		},
	}

	// Загружаем шаблоны
	tmpl := template.New("").Funcs(funcMap)
	tmpl, err = tmpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Ошибка парсинга шаблонов: %v", err)
	}

	// Создаем хендлер
	h := handler.NewHandler(tmpl, planetService, galaxyService, userService)

	// ========== ПУБЛИЧНЫЕ МАРШРУТЫ ==========
	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/planets", h.PlanetsHandler)
	http.HandleFunc("/planets/", h.PlanetDetailHandler)
	http.HandleFunc("/galaxies", h.GalaxiesHandler)
	http.HandleFunc("/galaxies/", h.GalaxyDetailHandler)

	// ========== АВТОРИЗАЦИЯ ==========
	http.HandleFunc("/register", h.RegisterHandler)
	http.HandleFunc("/login", h.LoginHandler)
	http.HandleFunc("/logout", h.LogoutHandler)

	// ========== АДМИН-ПАНЕЛЬ ==========
	http.HandleFunc("/admin", h.AdminDashboardHandler)
	http.HandleFunc("/admin/planets", h.AdminPlanetsHandler)
	http.HandleFunc("/admin/planets/new", h.AdminNewPlanetHandler)
	http.HandleFunc("/admin/planets/delete/", h.AdminDeletePlanetHandler)
	http.HandleFunc("/admin/planets/edit/", h.AdminEditPlanetHandler)
	http.HandleFunc("/admin/galaxies", h.AdminGalaxiesHandler)
	http.HandleFunc("/admin/galaxies/new", h.AdminNewGalaxyHandler)
	http.HandleFunc("/admin/galaxies/delete/", h.AdminDeleteGalaxyHandler)
	http.HandleFunc("/admin/galaxies/edit/", h.AdminEditGalaxyHandler)
	http.HandleFunc("/admin/users", h.AdminUsersHandler)
	http.HandleFunc("/admin/users/new", h.AdminNewUserHandler)
	http.HandleFunc("/admin/users/delete/", h.AdminDeleteUserHandler)
	http.HandleFunc("/admin/users/edit/", h.AdminEditUserHandler)
	http.HandleFunc("/admin/users/view/", h.AdminUserDetailHandler)

	// ========== ЛИЧНЫЙ КАБИНЕТ ПОЛЬЗОВАТЕЛЯ ==========
	http.HandleFunc("/dashboard", h.DashboardHandler)
	http.HandleFunc("/my/planets", h.MyPlanetsHandler)
	http.HandleFunc("/my/planets/new", h.MyPlanetNewHandler)
	http.HandleFunc("/my/planets/delete/", h.MyPlanetDeleteHandler)
	http.HandleFunc("/my/planets/edit/", h.MyPlanetEditHandler)
	http.HandleFunc("/my/galaxies", h.MyGalaxiesHandler)
	http.HandleFunc("/my/galaxies/new", h.MyGalaxyNewHandler)
	http.HandleFunc("/my/galaxies/delete/", h.MyGalaxyDeleteHandler)
	http.HandleFunc("/my/galaxies/edit/", h.MyGalaxyEditHandler)
	http.HandleFunc("/profile/delete", h.ProfileDeleteHandler)
	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Сервер запущен на http://localhost:%s", cfg.AppPort)
	log.Printf("База данных: %s", cfg.DBName)

	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
