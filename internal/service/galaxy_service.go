package service

import (
	"errors"

	"cosmos/internal/models"
	"cosmos/internal/repository"
)

type GalaxyService struct {
	galaxyRepo repository.GalaxyRepository
}

func NewGalaxyService(galaxyRepo repository.GalaxyRepository) *GalaxyService {
	return &GalaxyService{
		galaxyRepo: galaxyRepo,
	}
}

// GetAllGalaxies - получить все галактики
func (s *GalaxyService) GetAllGalaxies() ([]models.Galaxy, error) {
	return s.galaxyRepo.GetAll()
}

// GetGalaxyByID - получить галактику по ID
func (s *GalaxyService) GetGalaxyByID(id int) (*models.Galaxy, error) {
	if id <= 0 {
		return nil, errors.New("неверный ID галактики")
	}
	return s.galaxyRepo.GetByID(id)
}

// CreateGalaxy - создание галактики с валидацией
func (s *GalaxyService) CreateGalaxy(galaxy *models.Galaxy) error {
	if galaxy.Name == "" {
		return errors.New("название галактики обязательно")
	}
	if galaxy.Type == "" {
		return errors.New("тип галактики обязателен")
	}
	if galaxy.Description == "" {
		return errors.New("описание обязательно")
	}

	return s.galaxyRepo.Create(galaxy)
}

// UpdateGalaxy - обновление галактики
func (s *GalaxyService) UpdateGalaxy(galaxy *models.Galaxy) error {
	if galaxy.ID <= 0 {
		return errors.New("неверный ID галактики")
	}
	if galaxy.Name == "" {
		return errors.New("название галактики обязательно")
	}
	if galaxy.Type == "" {
		return errors.New("тип галактики обязателен")
	}
	if galaxy.Description == "" {
		return errors.New("описание обязательно")
	}

	return s.galaxyRepo.Update(galaxy)
}

// DeleteGalaxy - удаление галактики (с проверкой, что нет планет)
func (s *GalaxyService) DeleteGalaxy(id int) error {
	if id <= 0 {
		return errors.New("неверный ID галактики")
	}

	// Проверяем, есть ли планеты в галактике
	planetCount, err := s.galaxyRepo.GetPlanetCount(id)
	if err != nil {
		return err
	}
	if planetCount > 0 {
		return errors.New("нельзя удалить галактику, в которой есть планеты")
	}

	return s.galaxyRepo.Delete(id)
}

// GetGalaxyCount - количество галактик
func (s *GalaxyService) GetGalaxyCount() (int, error) {
	return s.galaxyRepo.Count()
}

// GetPlanetCountInGalaxy - количество планет в галактике
func (s *GalaxyService) GetPlanetCountInGalaxy(id int) (int, error) {
	return s.galaxyRepo.GetPlanetCount(id)
}
