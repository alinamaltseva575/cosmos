package handler

import (
	"errors"
	"html/template"
	"log"
	"net/http"

	"cosmos/internal/auth"
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

// requireAuth - проверка авторизации (для пользователей)
func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (*auth.Claims, error) {
	token := auth.GetTokenFromRequest(r)
	if token == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, errors.New("не авторизован")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, err
	}

	return claims, nil
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
