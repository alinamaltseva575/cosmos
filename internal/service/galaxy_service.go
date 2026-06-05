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

func (s *GalaxyService) GetAllGalaxies() ([]models.Galaxy, error) {
	return s.galaxyRepo.GetAll()
}

func (s *GalaxyService) GetGalaxyByID(id int64) (*models.Galaxy, error) { // int64
	if id <= 0 {
		return nil, errors.New("неверный ID галактики")
	}
	return s.galaxyRepo.GetByID(id)
}

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

func (s *GalaxyService) DeleteGalaxy(id int64) error { // int64
	if id <= 0 {
		return errors.New("неверный ID галактики")
	}

	planetCount, err := s.galaxyRepo.GetPlanetCount(id)
	if err != nil {
		return err
	}
	if planetCount > 0 {
		return errors.New("нельзя удалить галактику, в которой есть планеты")
	}

	return s.galaxyRepo.Delete(id)
}

func (s *GalaxyService) GetGalaxyCount() (int, error) {
	return s.galaxyRepo.Count()
}

func (s *GalaxyService) GetPlanetCountInGalaxy(id int64) (int, error) { // int64
	return s.galaxyRepo.GetPlanetCount(id)
}
