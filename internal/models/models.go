package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`    // Не отдаем хэш пароля в JSON
	Role         string    `json:"role"` // "admin" или "user"
	CreatedAt    time.Time `json:"created_at"`
}

type Planet struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	GalaxyID          *int      `json:"galaxy_id,omitempty"`
	GalaxyName        string    `json:"galaxy_name,omitempty"`
	Type              string    `json:"type"`
	DiameterKm        float64   `json:"diameter_km"`
	MassKg            float64   `json:"mass_kg"`
	OrbitalPeriodDays float64   `json:"orbital_period_days"`
	HasLife           bool      `json:"has_life"`
	IsHabitable       bool      `json:"is_habitable"`
	DiscoveredYear    *int      `json:"discovered_year,omitempty"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
}

type Galaxy struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Type                string    `json:"type"`
	DiameterLy          *int      `json:"diameter_ly,omitempty"`
	MassSuns            *float64  `json:"mass_suns,omitempty"`
	DistanceFromEarthLy *float64  `json:"distance_from_earth_ly,omitempty"`
	DiscoveredYear      *int      `json:"discovered_year,omitempty"`
	Description         string    `json:"description"`
	CreatedAt           time.Time `json:"created_at"`
}

// PageData - данные для передачи в HTML шаблоны
type PageData struct {
	Title       string
	CurrentPage string
	PlanetCount int
	GalaxyCount int
	Planets     []Planet
	Planet      *Planet
	Galaxies    []Galaxy
	Galaxy      *Galaxy
	// Для будущей авторизации
	IsAdmin  bool
	Username string
}
