package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

func (h *Handler) AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	users, err := h.UserService.GetAllUsers()
	if err != nil {
		log.Printf("Ошибка получения пользователей: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title:       "Управление пользователями",
		CurrentPage: "admin_users",
		Users:       users,
		IsAdmin:     true,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		UserID:      claims.UserID,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminNewUserHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		IsAuth      bool
		Username    string
		Role        string
		User        models.User
		Error       string
		Success     string
	}

	data := FormData{
		Title:       "Создание пользователя",
		CurrentPage: "admin_user_form",
		IsAdmin:     true,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		User:        models.User{},
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		role := r.FormValue("role")

		_, err := h.UserService.CreateUser(username, email, password, role)
		if err != nil {
			data.Error = err.Error()
			data.User.Username = username
			data.User.Email = email
			data.User.Role = role
		} else {
			http.Redirect(w, r, "/admin/users?success=Пользователь+создан", http.StatusFound)
			return
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := h.requireAdminAuth(w, r)
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

	// Нельзя удалить админа (id=1) НИКОГДА
	if id == 1 {
		http.Error(w, "Нельзя удалить главного администратора", http.StatusForbidden)
		return
	}

	userToDelete, err := h.UserService.GetUserByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Нельзя удалять админов
	if userToDelete.Role == "admin" {
		http.Error(w, "Нельзя удалять администраторов", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet {
		h.showDeleteConfirmation(w, "Пользователь", userToDelete.Username,
			"/admin/users/delete/"+strconv.FormatInt(id, 10),
			"/admin/users", userToDelete, false, 0)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	err = h.UserService.DeleteUser(id, claims.UserID, claims.Role)
	if err != nil {
		log.Printf("Ошибка удаления пользователя: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin/users?success=Пользователь+удален", http.StatusFound)
}

func (h *Handler) AdminEditUserHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAdminAuth(w, r)
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

	// Получаем пользователя
	user, err := h.UserService.GetUserByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Если это админ (id=1) и редактируем НЕ себя - запрещаем
	if user.Role == "admin" && id != claims.UserID {
		http.Error(w, "Нельзя редактировать другого администратора", http.StatusForbidden)
		return
	}

	// Если это обычный пользователь, но редактируем админа - запрещаем (на всякий случай)
	if user.Role == "admin" && claims.Role != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	type FormData struct {
		Title       string
		CurrentPage string
		IsAdmin     bool
		IsAuth      bool
		Username    string
		Role        string
		User        models.User
		Error       string
		Success     string
	}

	data := FormData{
		Title:       "Редактирование пользователя: " + user.Username,
		CurrentPage: "admin_user_form",
		IsAdmin:     true,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
		User:        *user,
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		role := r.FormValue("role")
		password := r.FormValue("password")

		// Обычный пользователь не может повысить себе роль до админа
		if claims.Role != "admin" && role == "admin" {
			data.Error = "Нельзя повысить роль до администратора"
		} else {
			err := h.UserService.UpdateUser(id, username, email, role, password)
			if err != nil {
				data.Error = err.Error()
				data.User.Username = username
				data.User.Email = email
				data.User.Role = role
			} else {
				data.Success = "Пользователь успешно обновлен!"
				data.User.Username = username
				data.User.Email = email
				data.User.Role = role
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminUserDetailHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	claims, err := h.requireAdminAuth(w, r)
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

	user, err := h.UserService.GetUserByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := models.PageData{
		Title:       "Просмотр пользователя: " + user.Username,
		CurrentPage: "admin_user_detail",
		User:        user,
		IsAdmin:     true,
		IsAuth:      true,
		Username:    claims.Username,
		Role:        claims.Role,
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
