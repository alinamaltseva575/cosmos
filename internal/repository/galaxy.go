package repository

import (
	"database/sql"
	"errors"

	"cosmos/internal/models"
)

// GalaxyRepository - интерфейс для работы с галактиками
type GalaxyRepository interface {
	GetAll() ([]models.Galaxy, error)
	GetByID(id int) (*models.Galaxy, error)
	Create(galaxy *models.Galaxy) error
	Update(galaxy *models.Galaxy) error
	Delete(id int) error
	Count() (int, error)
	GetPlanetCount(id int) (int, error) // проверка есть ли планеты
}

type galaxyRepository struct {
	db *sql.DB
}

// NewGalaxyRepository - конструктор
func NewGalaxyRepository(db *sql.DB) GalaxyRepository {
	return &galaxyRepository{db: db}
}

// GetAll - получить все галактики
func (r *galaxyRepository) GetAll() ([]models.Galaxy, error) {
	query := `
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description
		FROM galaxies
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galaxies []models.Galaxy
	for rows.Next() {
		var g models.Galaxy
		var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
		var discoveredYear sql.NullInt64

		err := rows.Scan(
			&g.ID, &g.Name, &g.Type, &diameterLy, &massSuns,
			&distanceFromEarthLy, &discoveredYear, &g.Description,
		)
		if err != nil {
			return nil, err
		}

		if diameterLy.Valid {
			val := diameterLy.Float64
			g.DiameterLy = &val
		}
		if massSuns.Valid {
			val := massSuns.Float64
			g.MassSuns = &val
		}
		if distanceFromEarthLy.Valid {
			val := distanceFromEarthLy.Float64
			g.DistanceFromEarthLy = &val
		}
		if discoveredYear.Valid {
			year := int(discoveredYear.Int64)
			g.DiscoveredYear = &year
		}

		galaxies = append(galaxies, g)
	}

	return galaxies, nil
}

// GetByID - получить галактику по ID
func (r *galaxyRepository) GetByID(id int) (*models.Galaxy, error) {
	var galaxy models.Galaxy
	var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
	var discoveredYear sql.NullInt64

	query := `
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description
		FROM galaxies
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&galaxy.ID, &galaxy.Name, &galaxy.Type, &diameterLy, &massSuns,
		&distanceFromEarthLy, &discoveredYear, &galaxy.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("галактика не найдена")
		}
		return nil, err
	}

	if diameterLy.Valid {
		val := diameterLy.Float64
		galaxy.DiameterLy = &val
	}
	if massSuns.Valid {
		val := massSuns.Float64
		galaxy.MassSuns = &val
	}
	if distanceFromEarthLy.Valid {
		val := distanceFromEarthLy.Float64
		galaxy.DistanceFromEarthLy = &val
	}
	if discoveredYear.Valid {
		year := int(discoveredYear.Int64)
		galaxy.DiscoveredYear = &year
	}

	return &galaxy, nil
}

// Create - создание галактики
func (r *galaxyRepository) Create(galaxy *models.Galaxy) error {
	query := `
		INSERT INTO galaxies (name, type, description, diameter_ly, mass_suns,
		                     distance_from_earth_ly, discovered_year)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	var diameterLy, massSuns, distanceFromEarthLy, discoveredYear any

	if galaxy.DiameterLy != nil {
		diameterLy = *galaxy.DiameterLy
	} else {
		diameterLy = nil
	}

	if galaxy.MassSuns != nil {
		massSuns = *galaxy.MassSuns
	} else {
		massSuns = nil
	}

	if galaxy.DistanceFromEarthLy != nil {
		distanceFromEarthLy = *galaxy.DistanceFromEarthLy
	} else {
		distanceFromEarthLy = nil
	}

	if galaxy.DiscoveredYear != nil {
		discoveredYear = *galaxy.DiscoveredYear
	} else {
		discoveredYear = nil
	}

	return r.db.QueryRow(query,
		galaxy.Name, galaxy.Type, galaxy.Description,
		diameterLy, massSuns, distanceFromEarthLy, discoveredYear,
	).Scan(&galaxy.ID, &galaxy.CreatedAt)
}

// Update - обновление галактики
func (r *galaxyRepository) Update(galaxy *models.Galaxy) error {
	query := `
		UPDATE galaxies
		SET name = $1, type = $2, description = $3, diameter_ly = $4,
		    mass_suns = $5, distance_from_earth_ly = $6, discovered_year = $7,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
	`

	var diameterLy, massSuns, distanceFromEarthLy, discoveredYear any

	if galaxy.DiameterLy != nil {
		diameterLy = *galaxy.DiameterLy
	} else {
		diameterLy = nil
	}

	if galaxy.MassSuns != nil {
		massSuns = *galaxy.MassSuns
	} else {
		massSuns = nil
	}

	if galaxy.DistanceFromEarthLy != nil {
		distanceFromEarthLy = *galaxy.DistanceFromEarthLy
	} else {
		distanceFromEarthLy = nil
	}

	if galaxy.DiscoveredYear != nil {
		discoveredYear = *galaxy.DiscoveredYear
	} else {
		discoveredYear = nil
	}

	result, err := r.db.Exec(query,
		galaxy.Name, galaxy.Type, galaxy.Description,
		diameterLy, massSuns, distanceFromEarthLy, discoveredYear,
		galaxy.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("галактика не найдена")
	}

	return nil
}

// Delete - удаление галактики
func (r *galaxyRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM galaxies WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("галактика не найдена")
	}

	return nil
}

// Count - количество галактик
func (r *galaxyRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM galaxies").Scan(&count)
	return count, err
}

// GetPlanetCount - количество планет в галактике
func (r *galaxyRepository) GetPlanetCount(id int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM planets WHERE galaxy_id = $1", id).Scan(&count)
	return count, err
}
