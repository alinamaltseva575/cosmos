package handler

import (
	"database/sql"
	"log"
	"net/http"

	"cosmos/internal/auth"
	"cosmos/internal/models"
)

// AdminLoginHandler - страница входа в админку
func (h *Handler) AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	log.Printf("🔐 Запрос на вход: %s", r.Method)

	// Если уже авторизован - редирект в админку
	if token := auth.GetTokenFromRequest(r); token != "" {
		log.Printf("🔐 Найден токен в запросе")
		if claims, err := auth.ValidateToken(token); err == nil && claims.Role == "admin" {
			log.Printf("✅ Уже авторизован как %s", claims.Username)
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
	}

	// Создаем структуру для данных формы
	type LoginPageData struct {
		models.PageData
		Username string
		Error    string
	}

	data := LoginPageData{
		PageData: models.PageData{
			Title:       "Вход в админ-панель",
			CurrentPage: "admin_login",
		},
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		log.Printf("🔐 Попытка входа: %s", username)

		data.Username = username

		// Ищем пользователя в БД
		var user models.User
		err := h.DB.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = $1",
			username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)

		if err != nil {
			if err == sql.ErrNoRows {
				data.Error = "Неверный логин или пароль"
				log.Printf("Пользователь не найден: %s", username)
			} else {
				log.Printf("Ошибка запроса пользователя: %v", err)
				data.Error = "Ошибка сервера"
			}
		} else if !auth.CheckPassword(password, user.PasswordHash) {
			data.Error = "Неверный логин или пароль"
			log.Printf("Неверный пароль для: %s", username)
		} else if user.Role != "admin" {
			data.Error = "У вас нет прав администратора"
			log.Printf("Не админ: %s (роль: %s)", username, user.Role)
		} else {
			log.Printf("✅ Успешная проверка логина/пароля для: %s", username)

			// Создаем JWT токен
			token, err := auth.GenerateToken(user.Username, user.Role, user.ID)
			if err != nil {
				log.Printf("Ошибка создания токена: %v", err)
				http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
				return
			}

			log.Printf("✅ Токен создан для: %s", user.Username)

			// Сохраняем токен в cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   24 * 60 * 60, // 24 часа
			})

			log.Printf("✅ Cookie установлен, редирект на /admin")

			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
	}

	// Используем шаблон admin_login
	err := h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_login: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminDashboardHandler - главная страница админки
func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	claims, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Получаем статистику
	var planetCount, galaxyCount, adminCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetCount)
	h.DB.QueryRow("SELECT COUNT(*) FROM galaxies").Scan(&galaxyCount)
	h.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)

	data := models.PageData{
		Title:       "Админ-панель",
		CurrentPage: "admin",
		PlanetCount: planetCount,
		GalaxyCount: galaxyCount,
		UserCount:   adminCount, // Теперь динамическое количество админов
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

// AdminLogoutHandler - выход из админки
func (h *Handler) AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Удаляем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Удалить cookie
	})

	http.Redirect(w, r, "/admin/login", http.StatusFound)
}
