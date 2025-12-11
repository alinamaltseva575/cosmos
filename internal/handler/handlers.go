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
	}

	// Парсим шаблоны из templates/ и templates/admin/
	tmpl := template.New("").Funcs(funcMap)

	// Парсим все HTML файлы в папке templates
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

	// Проверяем, какие шаблоны загрузились
	templateNames := []string{}
	for _, t := range tmpl.Templates() {
		if t.Name() != "" && t.Name() != "base" {
			templateNames = append(templateNames, t.Name())
		}
	}

	log.Printf("✅ Шаблоны загружены: %v", templateNames)

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

	// Используем базовый шаблон base.html
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
		// Создаем данные с пустым списком планет
		data := models.PageData{
			Title:       "Планеты",
			CurrentPage: "planets",
			Planets:     []models.Planet{},
		}
		// Указываем явно какой шаблон использовать для контента
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

	if err = rows.Err(); err != nil {
		log.Printf("❌ Ошибка итерации планет: %v", err)
	}

	log.Printf("✅ Найдено планет: %d", len(planets))

	data := models.PageData{
		Title:       "Планеты",
		CurrentPage: "planets",
		Planets:     planets,
	}

	// Сначала парсим шаблон планет, потом base
	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона planets: %v", err)
		// Показываем простую ошибку пользователю
		http.Error(w, "Ошибка отображения страницы. Проверьте консоль сервера.", http.StatusInternalServerError)
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
	var discoveredYear sql.NullInt64

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
		&planet.IsHabitable, &discoveredYear, &planet.Description,
		&planet.GalaxyName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ Планета с ID %d не найдена", id)
			// Показываем страницу 404
			data := models.PageData{
				Title:       "Планета не найдена",
				CurrentPage: "planets",
			}
			h.Tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}
		log.Printf("❌ Ошибка запроса планеты ID %d: %v", id, err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Обрабатываем nullable поля
	if discoveredYear.Valid {
		year := int(discoveredYear.Int64)
		planet.DiscoveredYear = &year
	}

	log.Printf("✅ Найдена планета: %s (ID: %d)", planet.Name, planet.ID)

	data := models.PageData{
		Title:       planet.Name,
		CurrentPage: "planets",
		Planet:      &planet,
	}

	// Используем шаблон planet.html внутри base.html
	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона planet detail: %v", err)
		log.Printf("Доступные шаблоны:")
		for _, t := range h.Tmpl.Templates() {
			if t.Name() != "" {
				log.Printf("  - %s", t.Name())
			}
		}
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
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
		var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
		var discoveredYear sql.NullInt64

		err := rows.Scan(
			&g.ID, &g.Name, &g.Type, &diameterLy, &massSuns,
			&distanceFromEarthLy, &discoveredYear, &g.Description,
		)
		if err != nil {
			log.Printf("❌ Ошибка сканирования галактики: %v", err)
			continue
		}

		// Обрабатываем nullable поля - исправлено здесь
		if diameterLy.Valid {
			val := diameterLy.Float64
			g.DiameterLy = &val
		}
		if massSuns.Valid {
			val := massSuns.Float64
			g.MassSuns = &val
		}
		if distanceFromEarthLy.Valid {
			val := distanceFromEarthLy.Float64
			g.DistanceFromEarthLy = &val
		}
		if discoveredYear.Valid {
			year := int(discoveredYear.Int64)
			g.DiscoveredYear = &year
		}

		galaxies = append(galaxies, g)
	}

	if err = rows.Err(); err != nil {
		log.Printf("❌ Ошибка итерации галактик: %v", err)
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
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
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
	var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
	var discoveredYear sql.NullInt64

	err = h.DB.QueryRow(`
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description
		FROM galaxies
		WHERE id = $1
	`, id).Scan(
		&galaxy.ID, &galaxy.Name, &galaxy.Type, &diameterLy, &massSuns,
		&distanceFromEarthLy, &discoveredYear, &galaxy.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ Галактика с ID %d не найдена", id)
			data := models.PageData{
				Title:       "Галактика не найдена",
				CurrentPage: "galaxies",
			}
			h.Tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}
		log.Printf("❌ Ошибка запроса галактики ID %d: %v", id, err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Обрабатываем nullable поля - исправлено здесь
	if diameterLy.Valid {
		val := diameterLy.Float64
		galaxy.DiameterLy = &val
	}
	if massSuns.Valid {
		val := massSuns.Float64
		galaxy.MassSuns = &val
	}
	if distanceFromEarthLy.Valid {
		val := distanceFromEarthLy.Float64
		galaxy.DistanceFromEarthLy = &val
	}
	if discoveredYear.Valid {
		year := int(discoveredYear.Int64)
		galaxy.DiscoveredYear = &year
	}

	log.Printf("✅ Найдена галактика: %s (ID: %d)", galaxy.Name, galaxy.ID)

	data := models.PageData{
		Title:       galaxy.Name,
		CurrentPage: "galaxies",
		Galaxy:      &galaxy,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона galaxy detail: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
