package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

// DashboardHandler - личный кабинет пользователя
func (h *Handler) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
	if err != nil {
		return
	}

	planets, err := h.PlanetService.GetUserPlanets(claims.UserID)
	if err != nil {
		log.Printf("Ошибка получения планет пользователя: %v", err)
		planets = []models.Planet{}
	}

	galaxies, err := h.GalaxyService.GetUserGalaxies(claims.UserID)
	if err != nil {
		log.Printf("Ошибка получения галактик пользователя: %v", err)
		galaxies = []models.Galaxy{}
	}

	data := models.PageData{
		Title:       "Личный кабинет",
		CurrentPage: "dashboard",
		Planets:     planets,
		Galaxies:    galaxies,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		UserID:      claims.UserID,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона dashboard: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyPlanetsHandler - список планет пользователя
func (h *Handler) MyPlanetsHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
	if err != nil {
		return
	}

	planets, err := h.PlanetService.GetUserPlanets(claims.UserID)
	if err != nil {
		log.Printf("Ошибка получения планет: %v", err)
		planets = []models.Planet{}
	}

	success := r.URL.Query().Get("success")

	data := models.PageData{
		Title:       "Мои планеты",
		CurrentPage: "my_planets",
		Planets:     planets,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		UserID:      claims.UserID,
		Success:     success,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона my_planets: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyPlanetNewHandler - создание планеты пользователем
func (h *Handler) MyPlanetNewHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
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
		IsAuth      bool
		IsAdmin     bool
		Username    string
		Role        string
		Planet      models.Planet
		Galaxies    []models.Galaxy
		Error       string
		Success     string
	}

	data := FormData{
		Title:       "Добавление планеты",
		CurrentPage: "my_planet_form",
		IsAuth:      true,
		IsAdmin:     false,
		Username:    claims.Username,
		Role:        claims.Role,
		Planet:      models.Planet{},
		Galaxies:    galaxies,
	}

	if r.Method == http.MethodPost {
		planet := h.parsePlanetForm(r)
		planet.CreatedBy = &claims.UserID
		err := h.PlanetService.CreatePlanet(&planet)
		if err != nil {
			data.Error = err.Error()
			data.Planet = planet
		} else {
			http.Redirect(w, r, "/my/planets?success=Планета+добавлена", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyPlanetEditHandler - редактирование планеты пользователем
func (h *Handler) MyPlanetEditHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
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

	// Проверка, что планета принадлежит пользователю
	if planet.CreatedBy == nil || *planet.CreatedBy != claims.UserID {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
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
		IsAuth      bool
		IsAdmin     bool
		Username    string
		Role        string
		Planet      models.Planet
		Galaxies    []models.Galaxy
		Error       string
		Success     string
	}

	data := FormData{
		Title:       "Редактирование планеты",
		CurrentPage: "my_planet_form",
		IsAuth:      true,
		IsAdmin:     false,
		Username:    claims.Username,
		Role:        claims.Role,
		Planet:      *planet,
		Galaxies:    galaxies,
	}

	if r.Method == http.MethodPost {
		updatedPlanet := h.parsePlanetForm(r)
		updatedPlanet.ID = id
		updatedPlanet.CreatedBy = &claims.UserID
		err := h.PlanetService.UpdatePlanet(&updatedPlanet)
		if err != nil {
			data.Error = err.Error()
			data.Planet = updatedPlanet
		} else {
			data.Success = "Планета успешно обновлена!"
			data.Planet = updatedPlanet
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyPlanetDeleteHandler - удаление планеты пользователем
func (h *Handler) MyPlanetDeleteHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.requireAuth(w, r)
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

	// Проверка прав
	if planet.CreatedBy == nil || *planet.CreatedBy != claims.UserID {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		h.showDeleteConfirmation(w, "Планета", planet.Name,
			"/my/planets/delete/"+strconv.FormatInt(id, 10),
			"/my/planets", planet, false, 0)
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

	http.Redirect(w, r, "/my/planets?success=Планета+удалена", http.StatusFound)
}

// MyGalaxiesHandler - список галактик пользователя
func (h *Handler) MyGalaxiesHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
	if err != nil {
		return
	}

	galaxies, err := h.GalaxyService.GetUserGalaxies(claims.UserID)
	if err != nil {
		log.Printf("Ошибка получения галактик: %v", err)
		galaxies = []models.Galaxy{}
	}

	success := r.URL.Query().Get("success")

	data := models.PageData{
		Title:       "Мои галактики",
		CurrentPage: "my_galaxies",
		Galaxies:    galaxies,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		UserID:      claims.UserID,
		Success:     success,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона my_galaxies: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyGalaxyNewHandler - создание галактики пользователем
func (h *Handler) MyGalaxyNewHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
	if err != nil {
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAuth      bool
		IsAdmin     bool
		Username    string
		Role        string
		Galaxy      models.Galaxy
		Error       string
		Success     string
	}

	data := FormData{
		Title:       "Добавление галактики",
		CurrentPage: "my_galaxy_form",
		IsAuth:      true,
		IsAdmin:     false,
		Username:    claims.Username,
		Role:        claims.Role,
		Galaxy:      models.Galaxy{},
	}

	if r.Method == http.MethodPost {
		galaxy := h.parseGalaxyForm(r)
		galaxy.CreatedBy = &claims.UserID
		err := h.GalaxyService.CreateGalaxy(&galaxy)
		if err != nil {
			data.Error = err.Error()
			data.Galaxy = galaxy
		} else {
			http.Redirect(w, r, "/my/galaxies?success=Галактика+добавлена", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyGalaxyEditHandler - редактирование галактики пользователем
func (h *Handler) MyGalaxyEditHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
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

	galaxy, err := h.GalaxyService.GetGalaxyByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Проверка прав
	if galaxy.CreatedBy == nil || *galaxy.CreatedBy != claims.UserID {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAuth      bool
		IsAdmin     bool
		Username    string
		Role        string
		Galaxy      models.Galaxy
		Error       string
	}

	data := FormData{
		Title:       "Редактирование галактики",
		CurrentPage: "my_galaxy_form",
		IsAuth:      true,
		IsAdmin:     false,
		Username:    claims.Username,
		Role:        claims.Role,
		Galaxy:      *galaxy,
	}

	if r.Method == http.MethodPost {
		updatedGalaxy := h.parseGalaxyForm(r)
		updatedGalaxy.ID = id
		updatedGalaxy.CreatedBy = &claims.UserID
		err := h.GalaxyService.UpdateGalaxy(&updatedGalaxy)
		if err != nil {
			data.Error = err.Error()
			data.Galaxy = updatedGalaxy
		} else {
			http.Redirect(w, r, "/my/galaxies?success=Галактика+обновлена", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// MyGalaxyDeleteHandler - удаление галактики пользователем
func (h *Handler) MyGalaxyDeleteHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.requireAuth(w, r)
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

	galaxy, err := h.GalaxyService.GetGalaxyByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Проверка прав
	if galaxy.CreatedBy == nil || *galaxy.CreatedBy != claims.UserID {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	planetCount, _ := h.GalaxyService.GetPlanetCountInGalaxy(id)
	hasPlanets := planetCount > 0

	if r.Method == http.MethodGet {
		h.showDeleteConfirmation(w, "Галактика", galaxy.Name,
			"/my/galaxies/delete/"+strconv.FormatInt(id, 10),
			"/my/galaxies", galaxy, hasPlanets, planetCount)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if hasPlanets {
		http.Error(w, "Нельзя удалить галактику, в которой есть планеты", http.StatusBadRequest)
		return
	}

	err = h.GalaxyService.DeleteGalaxy(id)
	if err != nil {
		log.Printf("Ошибка удаления галактики: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/my/galaxies?success=Галактика+удалена", http.StatusFound)
}

// ProfileDeleteHandler - удаление своего аккаунта пользователем
func (h *Handler) ProfileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAuth(w, r)
	if err != nil {
		return
	}

	// Только НЕ главный админ может удалить себя через профиль
	// Главный админ удаляет себя через админ-панель
	if claims.Role == "admin" && claims.UserID == 1 {
		http.Error(w, "Главный администратор может удалить свой аккаунт только через админ-панель", http.StatusForbidden)
		return
	}

	// Разрешаем удаление для обычных пользователей и обычных админов
	if r.Method == http.MethodGet {
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Удаление аккаунта</title></head>
			<body style="background:#f5f7fa;color:#2c3e50;font-family:Arial;padding:50px;">
				<div style="max-width:500px;margin:0 auto;background:white;padding:30px;border-radius:15px;box-shadow:0 4px 15px rgba(0,0,0,0.1);">
					<h1 style="color:#e53e3e;">🗑️ Удаление аккаунта</h1>
					<div style="background:#fff5f5;border:1px solid #fc8181;padding:15px;border-radius:8px;margin:20px 0;">
						<p>Вы уверены, что хотите удалить свой аккаунт?</p>
						<p>Это действие <strong>нельзя отменить</strong>.</p>
					</div>
					<p><strong>Логин:</strong> ` + claims.Username + `</p>
					<p><strong>Роль:</strong> ` + claims.Role + `</p>
					<form method="POST" action="/profile/delete" style="margin-top:20px;">
						<button type="submit" style="background:#e53e3e;color:white;padding:10px 20px;border:none;cursor:pointer;border-radius:8px;">🗑️ Да, удалить</button>
						<a href="/dashboard" style="background:#a0aec0;color:white;padding:10px 20px;text-decoration:none;border-radius:8px;margin-left:10px;">Отмена</a>
					</form>
				</div>
			</body>
			</html>
		`))
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	err = h.UserService.DeleteUser(claims.UserID, claims.UserID, claims.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/?deleted=true", http.StatusFound)
}
