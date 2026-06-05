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

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	UserID   int64  `json:"user_id"` // int → int64
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(username, role string, userID int64) (string, error) { // int → int64
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

func GetTokenFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie("auth_token"); err == nil {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}

	return ""
}
