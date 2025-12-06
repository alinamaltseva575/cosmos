package domain

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`    // не отдаём в JSON
	Role     string `json:"role"` // "admin" или "user"
}
