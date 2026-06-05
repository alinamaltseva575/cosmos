package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

func (h *Handler) AdminPlanetsHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	planets, err := h.PlanetService.GetAllPlanets()
	if err != nil {
		log.Printf("Ошибка получения планет: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	planetCount, _ := h.PlanetService.GetPlanetCount()
	success := r.URL.Query().Get("success")

	data := models.PageData{
		Title:       "Управление планетами",
		CurrentPage: "admin_planets",
		Planets:     planets,
		PlanetCount: planetCount,
		IsAdmin:     true,
		Success:     success,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_planets: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminNewPlanetHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	galaxies, err := h.PlanetService.GetGalaxiesForSelect()
	if err != nil {
		log.Printf("Ошибка получения галактик: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		Planet      models.Planet
		Galaxies    []models.Galaxy
		Error       string
	}

	data := FormData{
		Title:       "Добавление планеты",
		CurrentPage: "admin_planet_form",
		IsAdmin:     true,
		Planet:      models.Planet{},
		Galaxies:    galaxies,
	}

	if r.Method == http.MethodPost {
		planet := h.parsePlanetForm(r)
		err := h.PlanetService.CreatePlanet(&planet)
		if err != nil {
			data.Error = err.Error()
			data.Planet = planet
		} else {
			http.Redirect(w, r, "/admin/planets?success=Планета+добавлена", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminEditPlanetHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(pathParts[4], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	planet, err := h.PlanetService.GetPlanetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	galaxies, err := h.PlanetService.GetGalaxiesForSelect()
	if err != nil {
		log.Printf("Ошибка получения галактик: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		Planet      models.Planet
		Galaxies    []models.Galaxy
		Error       string
		Success     string
	}

	data := FormData{
		Title:       "Редактирование планеты",
		CurrentPage: "admin_planet_form",
		IsAdmin:     true,
		Planet:      *planet,
		Galaxies:    galaxies,
	}

	if r.Method == http.MethodPost {
		updatedPlanet := h.parsePlanetForm(r)
		updatedPlanet.ID = id
		err := h.PlanetService.UpdatePlanet(&updatedPlanet)
		if err != nil {
			data.Error = err.Error()
			data.Planet = updatedPlanet
		} else {
			data.Success = "Планета успешно обновлена!"
			data.Planet = updatedPlanet
			data.Planet.ID = id
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminDeletePlanetHandler(w http.ResponseWriter, r *http.Request) {
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(pathParts[4], 10, 64) // Atoi → ParseInt
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {
		planet, err := h.PlanetService.GetPlanetByID(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		h.showDeleteConfirmation(w, "Планета", planet.Name,
			"/admin/planets/delete/"+strconv.FormatInt(id, 10),
			"/admin/planets", planet, false, 0)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	err = h.PlanetService.DeletePlanet(id)
	if err != nil {
		log.Printf("Ошибка удаления планеты: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/planets?success=Планета+удалена", http.StatusFound)
}

func (h *Handler) parsePlanetForm(r *http.Request) models.Planet {
	var planet models.Planet
	planet.Name = r.FormValue("name")
	planet.Type = r.FormValue("type")
	planet.Description = r.FormValue("description")

	if diameter := r.FormValue("diameter_km"); diameter != "" {
		if val, err := strconv.ParseFloat(diameter, 64); err == nil {
			planet.DiameterKm = val
		}
	}
	if mass := r.FormValue("mass_kg"); mass != "" {
		if val, err := strconv.ParseFloat(mass, 64); err == nil {
			planet.MassKg = val
		}
	}
	if period := r.FormValue("orbital_period_days"); period != "" {
		if val, err := strconv.ParseFloat(period, 64); err == nil {
			planet.OrbitalPeriodDays = val
		}
	}
	if year := r.FormValue("discovered_year"); year != "" {
		if val, err := strconv.Atoi(year); err == nil {
			planet.DiscoveredYear = &val
		}
	}
	if galaxyID := r.FormValue("galaxy_id"); galaxyID != "" {
		if val, err := strconv.ParseInt(galaxyID, 10, 64); err == nil { // Atoi → ParseInt
			planet.GalaxyID = &val
		}
	}

	planet.HasLife = r.FormValue("has_life") == "on" || r.FormValue("has_life") == "true"
	planet.IsHabitable = r.FormValue("is_habitable") == "on" || r.FormValue("is_habitable") == "true"

	return planet
}
