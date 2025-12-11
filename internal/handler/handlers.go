package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"

	_ "github.com/lib/pq"
)

// Handler содержит зависимости
type Handler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

// NewHandler создает новый экземпляр Handler
func NewHandler(db *sql.DB) *Handler {
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
	}

	// Парсим шаблоны
	tmpl := template.New("").Funcs(funcMap)

	// Важно: указываем путь относительно корня проекта
	tmpl, err := tmpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("❌ Ошибка парсинга шаблонов: %v", err)
		// Создаем минимальный шаблон чтобы не падать
		tmpl = template.Must(template.New("base").Parse(`
			<!DOCTYPE html>
			<html>
			<head><title>{{.Title}}</title></head>
			<body>
				<h1>Ошибка загрузки шаблонов</h1>
				<p>Проверьте файлы шаблонов</p>
			</body>
			</html>`))
	}

	log.Printf("✅ Шаблоны загружены: %d штук", len(tmpl.Templates()))
	for _, t := range tmpl.Templates() {
		log.Printf("   - %s", t.Name())
	}

	return &Handler{
		DB:   db,
		Tmpl: tmpl,
	}
}

// setEncoding устанавливает правильную кодировку
func (h *Handler) setEncoding(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

// HomeHandler - главная страница
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	h.setEncoding(w)

	var planetCount, galaxyCount int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetCount)
	if err != nil {
		log.Printf("Ошибка получения количества планет: %v", err)
		planetCount = 0
	}

	err = h.DB.QueryRow("SELECT COUNT(*) FROM galaxies").Scan(&galaxyCount)
	if err != nil {
		log.Printf("Ошибка получения количества галактик: %v", err)
		galaxyCount = 0
	}

	data := models.PageData{
		Title:       "Главная",
		CurrentPage: "home",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона home: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// PlanetsHandler - список планет
func (h *Handler) PlanetsHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	query := `
		SELECT p.id, p.name, p.type, p.diameter_km, p.mass_kg,
		       p.orbital_period_days, p.has_life, p.is_habitable,
		       p.description, COALESCE(g.name, 'Не указана') as galaxy_name
		FROM planets p
		LEFT JOIN galaxies g ON p.galaxy_id = g.id
		ORDER BY p.name
	`

	rows, err := h.DB.Query(query)
	if err != nil {
		log.Printf("❌ Ошибка SQL запроса планет: %v", err)
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
		err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.DiameterKm, &p.MassKg,
			&p.OrbitalPeriodDays, &p.HasLife, &p.IsHabitable,
			&p.Description, &p.GalaxyName,
		)
		if err != nil {
			log.Printf("❌ Ошибка сканирования планеты: %v", err)
			continue
		}
		planets = append(planets, p)
	}

	log.Printf("✅ Найдено планет: %d", len(planets))
	for _, p := range planets {
		log.Printf("   - %s (тип: %s)", p.Name, p.Type)
	}

	data := models.PageData{
		Title:       "Планеты",
		CurrentPage: "planets",
		Planets:     planets,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона planets: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// PlanetDetailHandler - детальная страница планеты
func (h *Handler) PlanetDetailHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

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
		SELECT p.id, p.name, p.type, p.diameter_km, p.mass_kg,
		       p.orbital_period_days, p.has_life, p.is_habitable,
		       p.discovered_year, p.description,
		       COALESCE(g.name, 'Не указана') as galaxy_name
		FROM planets p
		LEFT JOIN galaxies g ON p.galaxy_id = g.id
		WHERE p.id = $1
	`, id).Scan(
		&planet.ID, &planet.Name, &planet.Type, &planet.DiameterKm,
		&planet.MassKg, &planet.OrbitalPeriodDays, &planet.HasLife,
		&planet.IsHabitable, &planet.DiscoveredYear, &planet.Description,
		&planet.GalaxyName,
	)

	if err != nil {
		log.Printf("❌ Планета не найдена ID %d: %v", id, err)
		data := models.PageData{
			Title:       "Не найдено",
			CurrentPage: "planets",
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	log.Printf("✅ Найдена планета: %s", planet.Name)

	data := models.PageData{
		Title:       planet.Name,
		CurrentPage: "planets",
		Planet:      &planet,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона planet detail: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// GalaxiesHandler - список галактик
func (h *Handler) GalaxiesHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	rows, err := h.DB.Query(`
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description
		FROM galaxies
		ORDER BY name
	`)

	if err != nil {
		log.Printf("❌ Ошибка SQL запроса галактик: %v", err)
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
		err := rows.Scan(
			&g.ID, &g.Name, &g.Type, &g.DiameterLy, &g.MassSuns,
			&g.DistanceFromEarthLy, &g.DiscoveredYear, &g.Description,
		)
		if err != nil {
			log.Printf("❌ Ошибка сканирования галактики: %v", err)
			continue
		}
		galaxies = append(galaxies, g)
	}

	log.Printf("✅ Найдено галактик: %d", len(galaxies))

	data := models.PageData{
		Title:       "Галактики",
		CurrentPage: "galaxies",
		Galaxies:    galaxies,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона galaxies: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// GalaxyDetailHandler - детальная страница галактики
func (h *Handler) GalaxyDetailHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

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
	err = h.DB.QueryRow(`
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description
		FROM galaxies
		WHERE id = $1
	`, id).Scan(
		&galaxy.ID, &galaxy.Name, &galaxy.Type, &galaxy.DiameterLy,
		&galaxy.MassSuns, &galaxy.DistanceFromEarthLy,
		&galaxy.DiscoveredYear, &galaxy.Description,
	)

	if err != nil {
		log.Printf("❌ Галактика не найдена ID %d: %v", id, err)
		data := models.PageData{
			Title:       "Не найдено",
			CurrentPage: "galaxies",
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	data := models.PageData{
		Title:       galaxy.Name,
		CurrentPage: "galaxies",
		Galaxy:      &galaxy,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона galaxy detail: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}
