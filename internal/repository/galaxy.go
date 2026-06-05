package repository

import (
	"database/sql"
	"errors"

	"cosmos/internal/models"
)

type GalaxyRepository interface {
	GetAll() ([]models.Galaxy, error)
	GetByID(id int64) (*models.Galaxy, error)
	GetUserGalaxies(userID int64) ([]models.Galaxy, error)
	Create(galaxy *models.Galaxy) error
	Update(galaxy *models.Galaxy) error
	Delete(id int64) error
	Count() (int, error)
	GetPlanetCount(id int64) (int, error)
}

type galaxyRepository struct {
	db *sql.DB
}

func NewGalaxyRepository(db *sql.DB) GalaxyRepository {
	return &galaxyRepository{db: db}
}

func (r *galaxyRepository) GetAll() ([]models.Galaxy, error) {
	query := `
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description, created_by
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
		var createdBy sql.NullInt64

		err := rows.Scan(
			&g.ID, &g.Name, &g.Type, &diameterLy, &massSuns,
			&distanceFromEarthLy, &discoveredYear, &g.Description, &createdBy,
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
		if createdBy.Valid {
			g.CreatedBy = &createdBy.Int64
		}

		galaxies = append(galaxies, g)
	}

	return galaxies, nil
}

func (r *galaxyRepository) GetUserGalaxies(userID int64) ([]models.Galaxy, error) {
	query := `
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description, created_by
		FROM galaxies
		WHERE created_by = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var galaxies []models.Galaxy
	for rows.Next() {
		var g models.Galaxy
		var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
		var discoveredYear sql.NullInt64
		var createdBy sql.NullInt64

		err := rows.Scan(
			&g.ID, &g.Name, &g.Type, &diameterLy, &massSuns,
			&distanceFromEarthLy, &discoveredYear, &g.Description, &createdBy,
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
		if createdBy.Valid {
			g.CreatedBy = &createdBy.Int64
		}

		galaxies = append(galaxies, g)
	}

	return galaxies, nil
}

func (r *galaxyRepository) GetByID(id int64) (*models.Galaxy, error) {
	var galaxy models.Galaxy
	var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
	var discoveredYear sql.NullInt64
	var createdBy sql.NullInt64

	query := `
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description, created_by
		FROM galaxies
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&galaxy.ID, &galaxy.Name, &galaxy.Type, &diameterLy, &massSuns,
		&distanceFromEarthLy, &discoveredYear, &galaxy.Description, &createdBy,
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
	if createdBy.Valid {
		galaxy.CreatedBy = &createdBy.Int64
	}

	return &galaxy, nil
}

func (r *galaxyRepository) Create(galaxy *models.Galaxy) error {
	query := `
		INSERT INTO galaxies (name, type, description, diameter_ly, mass_suns,
		                     distance_from_earth_ly, discovered_year, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

	var createdBy any
	if galaxy.CreatedBy != nil && *galaxy.CreatedBy > 0 {
		createdBy = *galaxy.CreatedBy
	} else {
		createdBy = nil
	}

	return r.db.QueryRow(query,
		galaxy.Name, galaxy.Type, galaxy.Description,
		diameterLy, massSuns, distanceFromEarthLy, discoveredYear, createdBy,
	).Scan(&galaxy.ID, &galaxy.CreatedAt)
}

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

func (r *galaxyRepository) Delete(id int64) error {
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

func (r *galaxyRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM galaxies").Scan(&count)
	return count, err
}

func (r *galaxyRepository) GetPlanetCount(id int64) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM planets WHERE galaxy_id = $1", id).Scan(&count)
	return count, err
}
