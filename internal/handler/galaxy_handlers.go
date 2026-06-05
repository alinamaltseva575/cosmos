package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

func (h *Handler) AdminGalaxiesHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	galaxies, err := h.GalaxyService.GetAllGalaxies()
	if err != nil {
		log.Printf("Ошибка получения галактик: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	galaxyCount, _ := h.GalaxyService.GetGalaxyCount()
	success := r.URL.Query().Get("success")

	data := models.PageData{
		Title:       "Управление галактиками",
		CurrentPage: "admin_galaxies",
		Galaxies:    galaxies,
		GalaxyCount: galaxyCount,
		IsAdmin:     true,
		Success:     success,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminNewGalaxyHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		Galaxy      models.Galaxy
		Error       string
	}

	data := FormData{
		Title:       "Добавление галактики",
		CurrentPage: "admin_galaxy_form",
		IsAdmin:     true,
		Galaxy:      models.Galaxy{},
	}

	if r.Method == http.MethodPost {
		galaxy := h.parseGalaxyForm(r)
		err := h.GalaxyService.CreateGalaxy(&galaxy)
		if err != nil {
			data.Error = err.Error()
			data.Galaxy = galaxy
		} else {
			http.Redirect(w, r, "/admin/galaxies?success=Галактика+добавлена", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminEditGalaxyHandler(w http.ResponseWriter, r *http.Request) {
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

	id, err := strconv.ParseInt(pathParts[4], 10, 64) // Atoi → ParseInt
	if err != nil {
		http.NotFound(w, r)
		return
	}

	galaxy, err := h.GalaxyService.GetGalaxyByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		Galaxy      models.Galaxy
		Error       string
	}

	data := FormData{
		Title:       "Редактирование галактики",
		CurrentPage: "admin_galaxy_form",
		IsAdmin:     true,
		Galaxy:      *galaxy,
	}

	if r.Method == http.MethodPost {
		updatedGalaxy := h.parseGalaxyForm(r)
		updatedGalaxy.ID = id
		err := h.GalaxyService.UpdateGalaxy(&updatedGalaxy)
		if err != nil {
			data.Error = err.Error()
			data.Galaxy = updatedGalaxy
		} else {
			http.Redirect(w, r, "/admin/galaxies?success=Галактика+обновлена", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminDeleteGalaxyHandler(w http.ResponseWriter, r *http.Request) {
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
		galaxy, err := h.GalaxyService.GetGalaxyByID(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		planetCount, _ := h.GalaxyService.GetPlanetCountInGalaxy(id)
		hasPlanets := planetCount > 0

		h.showDeleteConfirmation(w, "Галактика", galaxy.Name,
			"/admin/galaxies/delete/"+strconv.FormatInt(id, 10),
			"/admin/galaxies", galaxy, hasPlanets, planetCount)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	err = h.GalaxyService.DeleteGalaxy(id)
	if err != nil {
		log.Printf("Ошибка удаления галактики: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/galaxies?success=Галактика+удалена", http.StatusFound)
}

func (h *Handler) parseGalaxyForm(r *http.Request) models.Galaxy {
	var galaxy models.Galaxy
	galaxy.Name = r.FormValue("name")
	galaxy.Type = r.FormValue("type")
	galaxy.Description = r.FormValue("description")

	if diameter := r.FormValue("diameter_ly"); diameter != "" {
		if val, err := strconv.ParseFloat(diameter, 64); err == nil {
			galaxy.DiameterLy = &val
		}
	}
	if mass := r.FormValue("mass_suns"); mass != "" {
		if val, err := strconv.ParseFloat(mass, 64); err == nil {
			galaxy.MassSuns = &val
		}
	}
	if distance := r.FormValue("distance_from_earth_ly"); distance != "" {
		if val, err := strconv.ParseFloat(distance, 64); err == nil {
			galaxy.DistanceFromEarthLy = &val
		}
	}
	if year := r.FormValue("discovered_year"); year != "" {
		if val, err := strconv.Atoi(year); err == nil {
			galaxy.DiscoveredYear = &val
		}
	}

	return galaxy
}
