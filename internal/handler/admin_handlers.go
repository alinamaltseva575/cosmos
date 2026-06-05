package handler

import (
	"log"
	"net/http"

	"cosmos/internal/models"
)

// AdminDashboardHandler - главная страница админки
func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	planetCount, _ := h.PlanetService.GetPlanetCount()
	galaxyCount, _ := h.GalaxyService.GetGalaxyCount()
	userCount, _ := h.UserService.GetUserCount()

	data := models.PageData{
		Title:       "Админ-панель",
		CurrentPage: "admin",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
		UserCount:   userCount,
		IsAdmin:     true,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		UserID:      claims.UserID,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_dashboard: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
