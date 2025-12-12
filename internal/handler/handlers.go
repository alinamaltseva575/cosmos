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
	// Создаем карту функций для шаблонов
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
		// ДОБАВЛЯЕМ НОВЫЕ ФУНКЦИИ ДЛЯ РАБОТЫ С УКАЗАТЕЛЯМИ
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

	// Парсим шаблоны
	tmpl := template.New("").Funcs(funcMap)

	// Парсим ВСЕ HTML файлы
	tmpl, err := tmpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("❌ Ошибка парсинга шаблонов: %v", err)
	}

	// Проверяем, какие шаблоны загрузились
	log.Printf("📋 Загруженные шаблоны:")
	for _, t := range tmpl.Templates() {
		if t.Name() != "" {
			log.Printf("  - %s", t.Name())
		}
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
