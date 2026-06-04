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

	// ========== ФУНКЦИИ ДЛЯ ШАБЛОНОВ ==========
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
		"hasValue": func(p interface{}) bool {
			if p == nil {
				return false
			}
			switch v := p.(type) {
			case *int:
				return v != nil && *v != 0
			case *float64:
				return v != nil && *v != 0
			case *string:
				return v != nil && *v != ""
			}
			return false
		},
	}

	// Загружаем шаблоны с функциями
	tmpl := template.New("").Funcs(funcMap)
	tmpl, err = tmpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Ошибка парсинга шаблонов: %v", err)
	}

	// Создаем хендлер
	h := handler.NewHandler(tmpl, planetService, galaxyService, userService)

	// Настраиваем маршруты
	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/planets", h.PlanetsHandler)
	http.HandleFunc("/planets/", h.PlanetDetailHandler)
	http.HandleFunc("/galaxies", h.GalaxiesHandler)
	http.HandleFunc("/galaxies/", h.GalaxyDetailHandler)

	http.HandleFunc("/admin/login", h.AdminLoginHandler)
	http.HandleFunc("/admin/logout", h.AdminLogoutHandler)
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

	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Сервер запущен на http://localhost:%s", cfg.AppPort)
	log.Printf("База данных: %s", cfg.DBName)

	if err := http.ListenAndServe(":"+cfg.AppPort, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
