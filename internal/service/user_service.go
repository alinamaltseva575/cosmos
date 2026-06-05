package service

import (
	"errors"

	"cosmos/internal/auth"
	"cosmos/internal/models"
	"cosmos/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) GetUserByID(id int64) (*models.User, error) { // int64
	if id <= 0 {
		return nil, errors.New("неверный ID пользователя")
	}
	return s.userRepo.GetByID(id)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("логин не может быть пустым")
	}
	return s.userRepo.GetByUsername(username)
}

func (s *UserService) CreateUser(username, email, password, role string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("логин обязателен")
	}
	if email == "" {
		return nil, errors.New("email обязателен")
	}
	if password == "" {
		return nil, errors.New("пароль обязателен")
	}
	if len(password) < 6 {
		return nil, errors.New("пароль должен быть не менее 6 символов")
	}
	if role != "admin" && role != "user" {
		return nil, errors.New("роль должна быть 'admin' или 'user'")
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, errors.New("ошибка хэширования пароля")
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdateUser(id int64, username, email, role, newPassword string) error { // int64
	if id <= 0 {
		return errors.New("неверный ID пользователя")
	}
	if username == "" {
		return errors.New("логин обязателен")
	}
	if email == "" {
		return errors.New("email обязателен")
	}
	if role != "admin" && role != "user" {
		return errors.New("роль должна быть 'admin' или 'user'")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	user.Username = username
	user.Email = email
	user.Role = role

	if newPassword != "" {
		if len(newPassword) < 6 {
			return errors.New("пароль должен быть не менее 6 символов")
		}
		hashedPassword, err := auth.HashPassword(newPassword)
		if err != nil {
			return errors.New("ошибка хэширования пароля")
		}
		user.PasswordHash = hashedPassword
	} else {
		user.PasswordHash = ""
	}

	return s.userRepo.Update(user)
}

func (s *UserService) DeleteUser(id int64) error { // int64
	if id <= 0 {
		return errors.New("неверный ID пользователя")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	if user.Role == "admin" {
		adminCount, err := s.userRepo.CountAdmins()
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return errors.New("нельзя удалить последнего администратора")
		}
	}

	return s.userRepo.Delete(id)
}

func (s *UserService) GetUserCount() (int, error) {
	return s.userRepo.Count()
}

func (s *UserService) GetAdminCount() (int, error) {
	return s.userRepo.CountAdmins()
}

func (s *UserService) Authenticate(username, password string) (*models.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("неверный логин или пароль")
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("неверный логин или пароль")
	}

	return user, nil
}
