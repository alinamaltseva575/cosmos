package main

import (
	"html/template"
	"net/http"
)

// Структура для данных страницы
type PageData struct {
	Title       string
	CurrentPage string
}

func main() {
	// Загружаем шаблоны
	tmpl := template.Must(template.ParseGlob("templates/*.html"))

	// Настройка маршрутов
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:       "Главная",
			CurrentPage: "home",
		}
		tmpl.ExecuteTemplate(w, "base.html", data)
	})

	http.HandleFunc("/planets", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:       "Планеты",
			CurrentPage: "planets",
		}
		tmpl.ExecuteTemplate(w, "base.html", data)
	})

	http.HandleFunc("/galaxies", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:       "Галактики",
			CurrentPage: "galaxies",
		}
		tmpl.ExecuteTemplate(w, "base.html", data)
	})

	// Статические файлы (CSS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Запуск сервера
	println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
