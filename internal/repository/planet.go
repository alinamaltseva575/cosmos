package repository

import (
	"database/sql"
	"errors"

	"cosmos/internal/models"
)

// PlanetRepository - интерфейс для работы с планетами
type PlanetRepository interface {
	GetAll() ([]models.Planet, error)
	GetByID(id int) (*models.Planet, error)
	Create(planet *models.Planet) error
	Update(planet *models.Planet) error
	Delete(id int) error
	Count() (int, error)
	GetGalaxies() ([]models.Galaxy, error) // для выпадающего списка
}

type planetRepository struct {
	db *sql.DB
}

// NewPlanetRepository - конструктор
func NewPlanetRepository(db *sql.DB) PlanetRepository {
	return &planetRepository{db: db}
}

// GetAll - получить все планеты
func (r *planetRepository) GetAll() ([]models.Planet, error) {
	query := `
		SELECT p.id, p.name, p.type, p.diameter_km, p.mass_kg,
		       p.orbital_period_days, p.has_life, p.is_habitable,
		       p.description, COALESCE(g.name, 'Не указана') as galaxy_name
		FROM planets p
		LEFT JOIN galaxies g ON p.galaxy_id = g.id
		ORDER BY p.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var planets []models.Planet
	for rows.Next() {
		var p models.Planet
		err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.DiameterKm, &p.MassKg,
			&p.OrbitalPeriodDays, &p.HasLife, &p.IsHabitable,
			&p.Description, &p.GalaxyName,
		)
		if err != nil {
			return nil, err
		}
		planets = append(planets, p)
	}

	return planets, nil
}

// GetByID - получить планету по ID
func (r *planetRepository) GetByID(id int) (*models.Planet, error) {
	var planet models.Planet
	var discoveredYear sql.NullInt64
	var galaxyID sql.NullInt64
	var massKg sql.NullFloat64
	var orbitalPeriodDays sql.NullFloat64

	query := `
		SELECT id, name, type, description, diameter_km, mass_kg,
		       orbital_period_days, discovered_year, galaxy_id,
		       has_life, is_habitable, created_at
		FROM planets
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&planet.ID, &planet.Name, &planet.Type, &planet.Description,
		&planet.DiameterKm, &massKg, &orbitalPeriodDays,
		&discoveredYear, &galaxyID, &planet.HasLife, &planet.IsHabitable,
		&planet.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("планета не найдена")
		}
		return nil, err
	}

	// Обрабатываем nullable поля
	if discoveredYear.Valid {
		year := int(discoveredYear.Int64)
		planet.DiscoveredYear = &year
	}
	if galaxyID.Valid {
		id := int(galaxyID.Int64)
		planet.GalaxyID = &id
	}
	if massKg.Valid {
		planet.MassKg = massKg.Float64
	}
	if orbitalPeriodDays.Valid {
		planet.OrbitalPeriodDays = orbitalPeriodDays.Float64
	}

	return &planet, nil
}

// Create - создание планеты
func (r *planetRepository) Create(planet *models.Planet) error {
	query := `
		INSERT INTO planets (name, type, description, diameter_km, mass_kg,
		                    orbital_period_days, discovered_year, galaxy_id,
		                    has_life, is_habitable)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	var galaxyID any
	if planet.GalaxyID != nil && *planet.GalaxyID > 0 {
		galaxyID = *planet.GalaxyID
	} else {
		galaxyID = nil
	}

	var discoveredYear any
	if planet.DiscoveredYear != nil && *planet.DiscoveredYear != 0 {
		discoveredYear = *planet.DiscoveredYear
	} else {
		discoveredYear = nil
	}

	return r.db.QueryRow(query,
		planet.Name, planet.Type, planet.Description,
		planet.DiameterKm, planet.MassKg, planet.OrbitalPeriodDays,
		discoveredYear, galaxyID,
		planet.HasLife, planet.IsHabitable,
	).Scan(&planet.ID, &planet.CreatedAt)
}

// Update - обновление планеты
func (r *planetRepository) Update(planet *models.Planet) error {
	query := `
		UPDATE planets
		SET name = $1, type = $2, description = $3, diameter_km = $4,
		    mass_kg = $5, orbital_period_days = $6, discovered_year = $7,
		    galaxy_id = $8, has_life = $9, is_habitable = $10,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
	`

	var galaxyID any
	if planet.GalaxyID != nil && *planet.GalaxyID > 0 {
		galaxyID = *planet.GalaxyID
	} else {
		galaxyID = nil
	}

	var discoveredYear any
	if planet.DiscoveredYear != nil && *planet.DiscoveredYear != 0 {
		discoveredYear = *planet.DiscoveredYear
	} else {
		discoveredYear = nil
	}

	result, err := r.db.Exec(query,
		planet.Name, planet.Type, planet.Description,
		planet.DiameterKm, planet.MassKg, planet.OrbitalPeriodDays,
		discoveredYear, galaxyID,
		planet.HasLife, planet.IsHabitable,
		planet.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("планета не найдена")
	}

	return nil
}

// Delete - удаление планеты
func (r *planetRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM planets WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("планета не найдена")
	}

	return nil
}

// Count - количество планет
func (r *planetRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM planets").Scan(&count)
	return count, err
}

// GetGalaxies - получить все галактики (для выпадающего списка)
func (r *planetRepository) GetGalaxies() ([]models.Galaxy, error) {
	rows, err := r.db.Query("SELECT id, name FROM galaxies ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galaxies []models.Galaxy
	for rows.Next() {
		var g models.Galaxy
		err := rows.Scan(&g.ID, &g.Name)
		if err != nil {
			return nil, err
		}
		galaxies = append(galaxies, g)
	}

	return galaxies, nil
}
