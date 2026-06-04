package service

import (
	"errors"

	"cosmos/internal/models"
	"cosmos/internal/repository"
)

type PlanetService struct {
	planetRepo repository.PlanetRepository
	galaxyRepo repository.GalaxyRepository
}

func NewPlanetService(planetRepo repository.PlanetRepository, galaxyRepo repository.GalaxyRepository) *PlanetService {
	return &PlanetService{
		planetRepo: planetRepo,
		galaxyRepo: galaxyRepo,
	}
}

// GetAllPlanets - получить все планеты
func (s *PlanetService) GetAllPlanets() ([]models.Planet, error) {
	return s.planetRepo.GetAll()
}

// GetPlanetByID - получить планету по ID
func (s *PlanetService) GetPlanetByID(id int) (*models.Planet, error) {
	if id <= 0 {
		return nil, errors.New("неверный ID планеты")
	}
	return s.planetRepo.GetByID(id)
}

// CreatePlanet - создание планеты с валидацией
func (s *PlanetService) CreatePlanet(planet *models.Planet) error {
	// Валидация
	if planet.Name == "" {
		return errors.New("название планеты обязательно")
	}
	if planet.Type == "" {
		return errors.New("тип планеты обязателен")
	}
	if planet.Description == "" {
		return errors.New("описание обязательно")
	}
	if planet.DiameterKm <= 0 {
		return errors.New("диаметр должен быть больше 0")
	}

	return s.planetRepo.Create(planet)
}

// UpdatePlanet - обновление планеты
func (s *PlanetService) UpdatePlanet(planet *models.Planet) error {
	if planet.ID <= 0 {
		return errors.New("неверный ID планеты")
	}
	if planet.Name == "" {
		return errors.New("название планеты обязательно")
	}
	if planet.Type == "" {
		return errors.New("тип планеты обязателен")
	}
	if planet.Description == "" {
		return errors.New("описание обязательно")
	}

	return s.planetRepo.Update(planet)
}

// DeletePlanet - удаление планеты
func (s *PlanetService) DeletePlanet(id int) error {
	if id <= 0 {
		return errors.New("неверный ID планеты")
	}
	return s.planetRepo.Delete(id)
}

// GetPlanetCount - количество планет
func (s *PlanetService) GetPlanetCount() (int, error) {
	return s.planetRepo.Count()
}

// GetGalaxiesForSelect - получить галактики для выпадающего списка
func (s *PlanetService) GetGalaxiesForSelect() ([]models.Galaxy, error) {
	return s.planetRepo.GetGalaxies()
}
