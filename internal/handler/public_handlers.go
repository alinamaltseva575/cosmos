package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	h.setEncoding(w)

	planetCount, _ := h.PlanetService.GetPlanetCount()
	galaxyCount, _ := h.GalaxyService.GetGalaxyCount()

	data := models.PageData{
		Title:       "Главная",
		CurrentPage: "home",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
	}

	err := h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона home: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

func (h *Handler) PlanetsHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	planets, err := h.PlanetService.GetAllPlanets()
	if err != nil {
		log.Printf("Ошибка получения планет: %v", err)
		planets = []models.Planet{}
	}

	data := models.PageData{
		Title:       "Планеты",
		CurrentPage: "planets",
		Planets:     planets,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона planets: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) PlanetDetailHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64) // Atoi → ParseInt
	if err != nil {
		http.NotFound(w, r)
		return
	}

	planet, err := h.PlanetService.GetPlanetByID(id)
	if err != nil {
		log.Printf("Планета с ID %d не найдена: %v", id, err)
		data := models.PageData{
			Title:       "Планета не найдена",
			CurrentPage: "planets",
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	data := models.PageData{
		Title:       planet.Name,
		CurrentPage: "planets",
		Planet:      planet,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона planet detail: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) GalaxiesHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	galaxies, err := h.GalaxyService.GetAllGalaxies()
	if err != nil {
		log.Printf("Ошибка получения галактик: %v", err)
		galaxies = []models.Galaxy{}
	}

	data := models.PageData{
		Title:       "Галактики",
		CurrentPage: "galaxies",
		Galaxies:    galaxies,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона galaxies: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) GalaxyDetailHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64) // Atoi → ParseInt
	if err != nil {
		http.NotFound(w, r)
		return
	}

	galaxy, err := h.GalaxyService.GetGalaxyByID(id)
	if err != nil {
		log.Printf("Галактика с ID %d не найдена: %v", id, err)
		data := models.PageData{
			Title:       "Галактика не найдена",
			CurrentPage: "galaxies",
		}
		h.Tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	data := models.PageData{
		Title:       galaxy.Name,
		CurrentPage: "galaxies",
		Galaxy:      galaxy,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона galaxy detail: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
