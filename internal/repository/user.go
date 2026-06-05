package repository

import (
	"database/sql"
	"errors"

	"cosmos/internal/models"
)

type UserRepository interface {
	GetAll() ([]models.User, error)
	GetByID(id int64) (*models.User, error) // int → int64
	GetByUsername(username string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id int64) error // int → int64
	Count() (int, error)
	CountAdmins() (int, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, email, role, created_at
		FROM users
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) GetByID(id int64) (*models.User, error) { // int64
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, username, email, role, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at
		FROM users
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	return r.db.QueryRow(query,
		user.Username, user.Email, user.PasswordHash, user.Role,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepository) Update(user *models.User) error {
	var query string
	var args []interface{}

	if user.PasswordHash != "" {
		query = `UPDATE users SET username = $1, email = $2, role = $3, password_hash = $4 WHERE id = $5`
		args = []interface{}{user.Username, user.Email, user.Role, user.PasswordHash, user.ID}
	} else {
		query = `UPDATE users SET username = $1, email = $2, role = $3 WHERE id = $4`
		args = []interface{}{user.Username, user.Email, user.Role, user.ID}
	}

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("пользователь не найден")
	}

	return nil
}

func (r *userRepository) Delete(id int64) error { // int64
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("пользователь не найден")
	}

	return nil
}

func (r *userRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *userRepository) CountAdmins() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	return count, err
}
