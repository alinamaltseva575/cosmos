package auth

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default_secret_key_change_in_production")
		log.Println("⚠️  Используется дефолтный JWT_SECRET. Установите JWT_SECRET в .env файле!")
	}
}

// Claims - структура для JWT токена
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	UserID   int    `json:"user_id"`
	jwt.RegisteredClaims
}

// HashPassword - хэширование пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword - проверка пароля
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken - создание JWT токена
func GenerateToken(username, role string, userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		Role:     role,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken - проверка JWT токена
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("неверный токен")
	}

	return claims, nil
}

// GetTokenFromRequest - получение токена из запроса
func GetTokenFromRequest(r *http.Request) string {
	// Пробуем получить из cookie
	if cookie, err := r.Cookie("auth_token"); err == nil {
		return cookie.Value
	}

	// Пробуем получить из заголовка Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}

	return ""
}
