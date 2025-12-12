package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/auth"
	"cosmos/internal/models"
)

// AdminUsersHandler - список пользователей в админке
func (h *Handler) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Получаем пользователей из БД
	rows, err := h.DB.Query(`
		SELECT id, username, email, role, created_at
		FROM users
		ORDER BY id DESC
	`)
	if err != nil {
		log.Printf("Ошибка SQL запроса пользователей: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			log.Printf("Ошибка сканирования пользователя: %v", err)
			continue
		}
		users = append(users, user)
	}

	data := models.PageData{
		Title:       "Управление пользователями",
		CurrentPage: "admin_users",
		Users:       users,
		IsAdmin:     true,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_users: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminUserDetailHandler - просмотр пользователя
func (h *Handler) AdminUserDetailHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 4 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Получаем пользователя из БД
	var user models.User
	err = h.DB.QueryRow(`
		SELECT id, username, email, role, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.NotFound(w, r)
		} else {
			log.Printf("Ошибка получения пользователя: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Получаем статистику пользователя
	var planetCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM planets WHERE created_by = $1", id).Scan(&planetCount)

	data := models.PageData{
		Title:       "Просмотр пользователя: " + user.Username,
		CurrentPage: "admin_user_form",
		User:        &user,
		PlanetCount: planetCount,
		IsAdmin:     true,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_user_detail: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminNewUserHandler - форма создания нового пользователя
func (h *Handler) AdminNewUserHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Структура для данных формы
	type FormData struct {
		models.PageData
		User  models.User
		Error string
	}

	data := FormData{
		PageData: models.PageData{
			Title:       "Создание пользователя",
			CurrentPage: "admin_user_form",
			IsAdmin:     true,
		},
		User: models.User{},
	}

	// Обработка POST запроса
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		role := r.FormValue("role")

		data.User.Username = username
		data.User.Email = email
		data.User.Role = role

		// Валидация
		if username == "" || email == "" || password == "" || role == "" {
			data.Error = "Все поля обязательны для заполнения"
		} else if len(password) < 6 {
			data.Error = "Пароль должен быть не менее 6 символов"
		} else if role != "admin" && role != "user" {
			data.Error = "Роль должна быть 'admin' или 'user'"
		} else {
			// Проверяем, нет ли уже такого пользователя
			var exists bool
			h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)",
				username, email).Scan(&exists)

			if exists {
				data.Error = "Пользователь с таким логином или email уже существует"
			} else {
				// Хэшируем пароль
				hashedPassword, err := auth.HashPassword(password)
				if err != nil {
					log.Printf("Ошибка хэширования пароля: %v", err)
					data.Error = "Ошибка сервера"
				} else {
					// Сохраняем пользователя
					var userID int
					err := h.DB.QueryRow(
						`INSERT INTO users (username, email, password_hash, role)
                         VALUES ($1, $2, $3, $4) RETURNING id`,
						username, email, hashedPassword, role,
					).Scan(&userID)

					if err != nil {
						log.Printf("Ошибка создания пользователя: %v", err)
						data.Error = "Ошибка сохранения в базу данных"
					} else {
						log.Printf("Создан пользователь: %s (ID: %d, роль: %s)", username, userID, role)
						http.Redirect(w, r, "/admin/users", http.StatusFound)
						return
					}
				}
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_user_form: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminEditUserHandler - форма редактирования пользователя
func (h *Handler) AdminEditUserHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL: /admin/users/edit/{id}
	pathParts := strings.Split(r.URL.Path, "/")

	if len(pathParts) != 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[4]) // pathParts[4] это ID
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Структура для данных формы
	type FormData struct {
		models.PageData
		User         models.User
		Error        string
		Success      string
		ShowPassword bool
	}

	// Получаем пользователя из БД
	var user models.User
	err = h.DB.QueryRow(`
        SELECT id, username, email, role, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("Ошибка получения пользователя: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	data := FormData{
		PageData: models.PageData{
			Title:       "Редактирование пользователя: " + user.Username,
			CurrentPage: "admin_user_form",
			IsAdmin:     true,
		},
		User: user,
	}

	// Обработка POST запроса (обновление)
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		role := r.FormValue("role")

		data.User.Username = username
		data.User.Email = email
		data.User.Role = role
		data.ShowPassword = password != ""

		// Валидация
		if username == "" || email == "" || role == "" {
			data.Error = "Логин, email и роль обязательны"
		} else if role != "admin" && role != "user" {
			data.Error = "Роль должна быть 'admin' или 'user'"
		} else if password != "" && len(password) < 6 {
			data.Error = "Пароль должен быть не менее 6 символов"
		} else {
			// Проверяем, не занят ли логин/email другим пользователем
			var exists bool
			h.DB.QueryRow(
				`SELECT EXISTS(SELECT 1 FROM users WHERE (username = $1 OR email = $2) AND id != $3)`,
				username, email, id,
			).Scan(&exists)

			if exists {
				data.Error = "Логин или email уже заняты другим пользователем"
			} else {
				// Обновляем пользователя
				var query string
				var args []interface{}

				if password != "" {
					// Обновляем с паролем
					hashedPassword, err := auth.HashPassword(password)
					if err != nil {
						log.Printf("Ошибка хэширования пароля: %v", err)
						data.Error = "Ошибка сервера"
					} else {
						query = `UPDATE users SET username = $1, email = $2, role = $3, password_hash = $4 WHERE id = $5`
						args = []interface{}{username, email, role, hashedPassword, id}
					}
				} else {
					// Обновляем без пароля
					query = `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4`
					args = []interface{}{username, email, role, id}
				}

				if query != "" {
					result, err := h.DB.Exec(query, args...)
					if err != nil {
						log.Printf("Ошибка обновления пользователя: %v", err)
						data.Error = "Ошибка сохранения в базу данных"
					} else {
						rowsAffected, _ := result.RowsAffected()
						if rowsAffected > 0 {
							data.Success = "Пользователь успешно обновлен"
							data.User.Username = username
							data.User.Email = email
							data.User.Role = role
							log.Printf("Обновлен пользователь ID %d", id)
						}
					}
				}
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_user_form (edit): %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminDeleteUserHandler - удаление пользователя
func (h *Handler) AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL: /admin/users/delete/{id}
	pathParts := strings.Split(r.URL.Path, "/")

	if len(pathParts) != 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[4]) // pathParts[4] это ID
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Если GET запрос - показываем страницу подтверждения
	if r.Method == http.MethodGet {
		h.showDeleteUserConfirmation(w, r, id)
		return
	}

	// Если POST запрос - выполняем удаление
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Нельзя удалить первого админа (ID=1)
	if id == 1 {
		http.Error(w, "Нельзя удалить главного администратора", http.StatusBadRequest)
		return
	}

	// Получаем имя пользователя для логирования
	var username string
	h.DB.QueryRow("SELECT username FROM users WHERE id = $1", id).Scan(&username)

	// Удаляем пользователя
	result, err := h.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Printf("Ошибка удаления пользователя %d: %v", id, err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	log.Printf("Пользователь удален: %s (ID %d)", username, id)
	http.Redirect(w, r, "/admin/users?success=Пользователь+"+username+"+удален", http.StatusFound)
}

// Добавим функцию подтверждения удаления для пользователей
func (h *Handler) showDeleteUserConfirmation(w http.ResponseWriter, r *http.Request, id int) {
	log.Printf("showDeleteUserConfirmation вызван для ID: %d", id)

	h.setEncoding(w)

	// Получаем пользователя из БД
	var user models.User
	err := h.DB.QueryRow(`
        SELECT id, username, email, role, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		log.Printf("Ошибка получения пользователя для удаления: %v", err)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Структура для данных страницы подтверждения
	type DeleteData struct {
		models.PageData
		ObjectType  string
		ObjectName  string
		ObjectData  interface{}
		DeleteURL   string
		ReturnURL   string
		HasPlanets  bool
		PlanetCount int
	}

	data := DeleteData{
		PageData: models.PageData{
			Title:       "Подтверждение удаления пользователя",
			CurrentPage: "admin_confirm_delete",
			IsAdmin:     true,
		},
		ObjectType:  "Пользователь",
		ObjectName:  user.Username,
		ObjectData:  user,
		DeleteURL:   "/admin/users/delete/" + strconv.Itoa(id),
		ReturnURL:   "/admin/users",
		HasPlanets:  false,
		PlanetCount: 0,
	}

	log.Printf("Данные для шаблона пользователя: ObjectType=%s, ObjectName=%s, Role=%s",
		data.ObjectType, data.ObjectName, user.Role)

	// Пробуем выполнить шаблон
	err = h.Tmpl.ExecuteTemplate(w, "admin_confirm_delete", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона admin_confirm_delete для пользователя: %v", err)

		// Покажем простую страницу ошибки
		fmt.Fprintf(w, `
            <html><body style="background:#0a0a2a;color:white;padding:50px;">
            <h1>Ошибка загрузки шаблона</h1>
            <p>%v</p>
            <p>ObjectType: %s</p>
            <p>ObjectName: %s</p>
            <p>Role: %s</p>
            <a href="/admin/users">Назад к пользователям</a>
            </body></html>
        `, err, data.ObjectType, data.ObjectName, user.Role)
	}
}
