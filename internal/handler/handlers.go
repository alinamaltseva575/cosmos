package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

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
