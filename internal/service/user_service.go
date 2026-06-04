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

// GetAllUsers - получить всех пользователей
func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAll()
}

// GetUserByID - получить пользователя по ID
func (s *UserService) GetUserByID(id int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("неверный ID пользователя")
	}
	return s.userRepo.GetByID(id)
}

// GetUserByUsername - получить пользователя по имени (для авторизации)
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	if username == "" {
		return nil, errors.New("логин не может быть пустым")
	}
	return s.userRepo.GetByUsername(username)
}

// CreateUser - создание пользователя
func (s *UserService) CreateUser(username, email, password, role string) (*models.User, error) {
	// Валидация
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

	// Хэшируем пароль
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

// UpdateUser - обновление пользователя
func (s *UserService) UpdateUser(id int, username, email, role, newPassword string) error {
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

	// Получаем текущего пользователя
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	user.Username = username
	user.Email = email
	user.Role = role

	// Если передан новый пароль - хэшируем
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
		user.PasswordHash = "" // пустая строка = не менять пароль
	}

	return s.userRepo.Update(user)
}

// DeleteUser - удаление пользователя (с проверкой последнего админа)
func (s *UserService) DeleteUser(id int) error {
	if id <= 0 {
		return errors.New("неверный ID пользователя")
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Если это админ - проверяем, не последний ли он
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

// GetUserCount - количество пользователей
func (s *UserService) GetUserCount() (int, error) {
	return s.userRepo.Count()
}

// GetAdminCount - количество администраторов
func (s *UserService) GetAdminCount() (int, error) {
	return s.userRepo.CountAdmins()
}

// Authenticate - аутентификация пользователя
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
