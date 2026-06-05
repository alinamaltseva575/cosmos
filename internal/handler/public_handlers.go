package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/auth"
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

	isAuth := false
	username := ""
	role := ""
	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil {
			isAuth = true
			username = claims.Username
			role = claims.Role
		}
	}

	data := models.PageData{
		Title:       "Главная",
		CurrentPage: "home",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
		IsAuth:      isAuth,
		Username:    username,
		Role:        role,
	}

	err := h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона home: %v", err)
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
	}
}

// RegisterHandler - страница регистрации
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	type RegisterData struct {
		Title       string
		CurrentPage string
		IsAuth      bool
		Username    string
		Role        string
		Error       string
		RegUsername string
		RegEmail    string
	}

	data := RegisterData{
		Title:       "Регистрация",
		CurrentPage: "register",
		IsAuth:      false,
		Username:    "",
		Role:        "",
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		data.RegUsername = username
		data.RegEmail = email

		if password != confirmPassword {
			data.Error = "Пароли не совпадают"
		} else {
			_, err := h.UserService.CreateUser(username, email, password, "user")
			if err != nil {
				data.Error = err.Error()
			} else {
				http.Redirect(w, r, "/login?registered=true", http.StatusFound)
				return
			}
		}
	}

	err := h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона register: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// LoginHandler - страница входа
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Если уже авторизован - редирект в зависимости от роли
	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil {
			if claims.Role == "admin" {
				http.Redirect(w, r, "/admin", http.StatusFound)
			} else {
				http.Redirect(w, r, "/dashboard", http.StatusFound)
			}
			return
		}
	}

	type LoginData struct {
		Title       string
		CurrentPage string
		IsAuth      bool
		Username    string
		Role        string
		Error       string
		LoginName   string
		Registered  bool
	}

	data := LoginData{
		Title:       "Вход",
		CurrentPage: "login",
		IsAuth:      false,
		Username:    "",
		Role:        "",
		Registered:  r.URL.Query().Get("registered") == "true",
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		data.LoginName = username

		user, err := h.UserService.Authenticate(username, password)
		if err != nil {
			data.Error = "Неверный логин или пароль"
		} else {
			token, err := auth.GenerateToken(user.Username, user.Role, user.ID)
			if err != nil {
				log.Printf("Ошибка создания токена: %v", err)
				data.Error = "Ошибка сервера"
			} else {
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					MaxAge:   24 * 60 * 60,
				})

				if user.Role == "admin" {
					http.Redirect(w, r, "/admin", http.StatusFound)
				} else {
					http.Redirect(w, r, "/dashboard", http.StatusFound)
				}
				return
			}
		}
	}

	err := h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона login: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// LogoutHandler - выход
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

// PlanetsHandler - список планет (публичный)
func (h *Handler) PlanetsHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	planets, err := h.PlanetService.GetAllPlanets()
	if err != nil {
		log.Printf("Ошибка получения планет: %v", err)
		planets = []models.Planet{}
	}

	isAuth := false
	username := ""
	role := ""
	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil {
			isAuth = true
			username = claims.Username
			role = claims.Role
		}
	}

	data := models.PageData{
		Title:       "Планеты",
		CurrentPage: "planets",
		Planets:     planets,
		IsAuth:      isAuth,
		Username:    username,
		Role:        role,
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

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
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

	isAuth := false
	username := ""
	role := ""
	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil {
			isAuth = true
			username = claims.Username
			role = claims.Role
		}
	}

	data := models.PageData{
		Title:       planet.Name,
		CurrentPage: "planets",
		Planet:      planet,
		IsAuth:      isAuth,
		Username:    username,
		Role:        role,
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

	isAuth := false
	username := ""
	role := ""
	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil {
			isAuth = true
			username = claims.Username
			role = claims.Role
		}
	}

	data := models.PageData{
		Title:       "Галактики",
		CurrentPage: "galaxies",
		Galaxies:    galaxies,
		IsAuth:      isAuth,
		Username:    username,
		Role:        role,
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

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
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

	isAuth := false
	username := ""
	role := ""
	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil {
			isAuth = true
			username = claims.Username
			role = claims.Role
		}
	}

	data := models.PageData{
		Title:       galaxy.Name,
		CurrentPage: "galaxies",
		Galaxy:      galaxy,
		IsAuth:      isAuth,
		Username:    username,
		Role:        role,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона galaxy detail: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
