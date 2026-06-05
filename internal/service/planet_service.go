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

func (s *PlanetService) GetAllPlanets() ([]models.Planet, error) {
	return s.planetRepo.GetAll()
}

func (s *PlanetService) GetPlanetByID(id int64) (*models.Planet, error) { // int64
	if id <= 0 {
		return nil, errors.New("неверный ID планеты")
	}
	return s.planetRepo.GetByID(id)
}

func (s *PlanetService) CreatePlanet(planet *models.Planet) error {
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

func (s *PlanetService) DeletePlanet(id int64) error { // int64
	if id <= 0 {
		return errors.New("неверный ID планеты")
	}
	return s.planetRepo.Delete(id)
}

func (s *PlanetService) GetPlanetCount() (int, error) {
	return s.planetRepo.Count()
}

func (s *PlanetService) GetGalaxiesForSelect() ([]models.Galaxy, error) {
	return s.planetRepo.GetGalaxies()
}
