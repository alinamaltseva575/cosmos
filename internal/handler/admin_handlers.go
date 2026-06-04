package handler

import (
	"log"
	"net/http"

	"cosmos/internal/auth"
	"cosmos/internal/models"
)

func (h *Handler) AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	if token := auth.GetTokenFromRequest(r); token != "" {
		if claims, err := auth.ValidateToken(token); err == nil && claims.Role == "admin" {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
	}

	type LoginPageData struct {
		Title       string
		CurrentPage string
		Username    string
		Error       string
	}

	data := LoginPageData{
		Title:       "Вход в админ-панель",
		CurrentPage: "admin_login",
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		data.Username = username

		user, err := h.UserService.Authenticate(username, password)
		if err != nil {
			data.Error = "Неверный логин или пароль"
			log.Printf("Ошибка аутентификации: %v", err)
		} else if user.Role != "admin" {
			data.Error = "У вас нет прав администратора"
			log.Printf("Не админ: %s", username)
		} else {
			token, err := auth.GenerateToken(user.Username, user.Role, user.ID)
			if err != nil {
				log.Printf("Ошибка создания токена: %v", err)
				http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   24 * 60 * 60,
			})

			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
	}

	err := h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_login: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	planetCount, _ := h.PlanetService.GetPlanetCount()
	galaxyCount, _ := h.GalaxyService.GetGalaxyCount()
	adminCount, _ := h.UserService.GetAdminCount()

	data := models.PageData{
		Title:       "Админ-панель",
		CurrentPage: "admin",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
		UserCount:   adminCount,
		IsAdmin:     true,
		Username:    claims.Username,
		Role:        claims.Role,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_dashboard: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}
