package handler

import (
	"html/template"
	"log"
	"net/http"

	"cosmos/internal/service"
)

type Handler struct {
	Tmpl          *template.Template
	PlanetService *service.PlanetService
	GalaxyService *service.GalaxyService
	UserService   *service.UserService
}

func NewHandler(
	tmpl *template.Template,
	planetService *service.PlanetService,
	galaxyService *service.GalaxyService,
	userService *service.UserService,
) *Handler {
	return &Handler{
		Tmpl:          tmpl,
		PlanetService: planetService,
		GalaxyService: galaxyService,
		UserService:   userService,
	}
}

func (h *Handler) setEncoding(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

// showDeleteConfirmation - общая функция для подтверждения удаления
func (h *Handler) showDeleteConfirmation(w http.ResponseWriter, objectType, objectName, deleteURL, returnURL string, objectData interface{}, hasPlanets bool, planetCount int) {
	h.setEncoding(w)

	type DeleteData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		ObjectType  string
		ObjectName  string
		ObjectData  interface{}
		DeleteURL   string
		ReturnURL   string
		HasPlanets  bool
		PlanetCount int
	}

	data := DeleteData{
		Title:       "Подтверждение удаления " + objectType,
		CurrentPage: "admin_confirm_delete",
		IsAdmin:     true,
		ObjectType:  objectType,
		ObjectName:  objectName,
		ObjectData:  objectData,
		DeleteURL:   deleteURL,
		ReturnURL:   returnURL,
		HasPlanets:  hasPlanets,
		PlanetCount: planetCount,
	}

	err := h.Tmpl.ExecuteTemplate(w, "admin_confirm_delete", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона confirm_delete: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
