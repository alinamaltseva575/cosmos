package handler

import (
	"errors"
	"log"
	"net/http" // ДОБАВЬТЕ ЭТОТ ИМПОРТ

	"cosmos/internal/auth"
)

// requireAdminAuth - middleware для проверки авторизации админа
func (h *Handler) requireAdminAuth(w http.ResponseWriter, r *http.Request) (*auth.Claims, error) {
	token := auth.GetTokenFromRequest(r)
	if token == "" {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
		return nil, errors.New("не авторизован")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
		return nil, err
	}

	if claims.Role != "admin" {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return nil, errors.New("не админ")
	}

	log.Printf("✅ Авторизован: %s (роль: %s)", claims.Username, claims.Role)
	return claims, nil
}
