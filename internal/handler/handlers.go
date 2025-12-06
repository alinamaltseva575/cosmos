package handler

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

// Handler содержит зависимости (БД, шаблоны и т.д.)
type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

// NewHandler создает новый экземпляр Handler
// В internal/handler/handlers.go
func NewHandler(db *sql.DB) *Handler {
	// Шаблоны создаются ВНУТРИ функции
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	// Проверяем загрузку шаблонов
	for _, t := range tmpl.Templates() {
		log.Printf("✅ Загружен шаблон: %s", t.Name())
	}

	return &Handler{
		DB:   db,
		Tmpl: tmpl, // Tmpl создается здесь
	}
}

// HomeHandler - главная страница
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var planetCount, galaxyCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetCount)
	h.DB.QueryRow("SELECT COUNT(*) FROM galaxies").Scan(&galaxyCount)

	data := models.PageData{
		Title:       "Главная",
		CurrentPage: "home",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
	}

	// Используйте ваш общий Tmpl, который должен загружать все файлы
	h.Tmpl.ExecuteTemplate(w, "base.html", data)
}

// PlanetsHandler - список планет
func (h *Handler) PlanetsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, name, type, diameter_km, has_life, is_habitable, description
		FROM planets
		ORDER BY name
	`)

	if err != nil {
		log.Printf("Ошибка получения планет: %v", err)
		data := models.PageData{
			Title:       "Планеты",
			CurrentPage: "planets",
			Planets:     []models.Planet{},
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}
	defer rows.Close()

	var planets []models.Planet
	for rows.Next() {
		var p models.Planet
		err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.DiameterKm, &p.HasLife, &p.IsHabitable, &p.Description)
		if err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			continue
		}
		planets = append(planets, p)
	}

	data := models.PageData{
		Title:       "Планеты",
		CurrentPage: "planets",
		Planets:     planets,
	}
	h.Tmpl.ExecuteTemplate(w, "base.html", data)
}

// PlanetDetailHandler - детальная страница планеты
func (h *Handler) PlanetDetailHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var planet models.Planet
	err = h.DB.QueryRow(`
		SELECT id, name, type, diameter_km, has_life, is_habitable, description
		FROM planets
		WHERE id = $1
	`, id).Scan(&planet.ID, &planet.Name, &planet.Type, &planet.DiameterKm, &planet.HasLife, &planet.IsHabitable, &planet.Description)

	if err != nil {
		log.Printf("Планета не найдена: %v", err)
		data := models.PageData{
			Title:       "Не найдено",
			CurrentPage: "planets",
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	data := models.PageData{
		Title:       planet.Name,
		CurrentPage: "planets",
		Planet:      &planet,
	}
	h.Tmpl.ExecuteTemplate(w, "base.html", data)
}

// GalaxiesHandler - список галактик
func (h *Handler) GalaxiesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, name, type, diameter_ly, description
		FROM galaxies
		ORDER BY name
	`)

	if err != nil {
		log.Printf("Ошибка получения галактик: %v", err)
		data := models.PageData{
			Title:       "Галактики",
			CurrentPage: "galaxies",
			Galaxies:    []models.Galaxy{},
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}
	defer rows.Close()

	var galaxies []models.Galaxy
	for rows.Next() {
		var g models.Galaxy
		var diameterLy *int
		err := rows.Scan(&g.ID, &g.Name, &g.Type, &diameterLy, &g.Description)
		if err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			continue
		}
		g.DiameterLy = diameterLy
		galaxies = append(galaxies, g)
	}

	data := models.PageData{
		Title:       "Галактики",
		CurrentPage: "galaxies",
		Galaxies:    galaxies,
	}
	h.Tmpl.ExecuteTemplate(w, "base.html", data)
}

// GalaxyDetailHandler - детальная страница галактики
func (h *Handler) GalaxyDetailHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var galaxy models.Galaxy
	var diameterLy *int
	err = h.DB.QueryRow(`
		SELECT id, name, type, diameter_ly, description
		FROM galaxies
		WHERE id = $1
	`, id).Scan(&galaxy.ID, &galaxy.Name, &galaxy.Type, &diameterLy, &galaxy.Description)

	if err != nil {
		log.Printf("Галактика не найдена: %v", err)
		data := models.PageData{
			Title:       "Не найдено",
			CurrentPage: "galaxies",
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	galaxy.DiameterLy = diameterLy

	data := models.PageData{
		Title:       galaxy.Name,
		CurrentPage: "galaxies",
		Galaxy:      &galaxy,
	}
	h.Tmpl.ExecuteTemplate(w, "base.html", data)
}
