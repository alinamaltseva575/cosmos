package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Planet struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	GalaxyID          *int64    `json:"galaxy_id,omitempty"`
	GalaxyName        string    `json:"galaxy_name,omitempty"`
	Type              string    `json:"type"`
	DiameterKm        float64   `json:"diameter_km"`
	MassKg            float64   `json:"mass_kg"`
	OrbitalPeriodDays float64   `json:"orbital_period_days"`
	HasLife           bool      `json:"has_life"`
	IsHabitable       bool      `json:"is_habitable"`
	DiscoveredYear    *int      `json:"discovered_year,omitempty"`
	Description       string    `json:"description"`
	CreatedBy         *int64    `json:"created_by,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

type Galaxy struct {
	ID                  int64     `json:"id"`
	Name                string    `json:"name"`
	Type                string    `json:"type"`
	DiameterLy          *float64  `json:"diameter_ly,omitempty"`
	MassSuns            *float64  `json:"mass_suns,omitempty"`
	DistanceFromEarthLy *float64  `json:"distance_from_earth_ly,omitempty"`
	DiscoveredYear      *int      `json:"discovered_year,omitempty"`
	Description         string    `json:"description"`
	CreatedBy           *int64    `json:"created_by,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type PageData struct {
	Title       string
	CurrentPage string
	PlanetCount int
	GalaxyCount int
	UserCount   int
	Planets     []Planet
	Planet      *Planet
	Galaxies    []Galaxy
	Galaxy      *Galaxy
	Users       []User
	User        *User
	IsAdmin     bool
	IsAuth      bool // НОВОЕ: авторизован ли пользователь
	Username    string
	Role        string
	UserID      int64 // НОВОЕ: ID текущего пользователя
	AppPort     string
	Environment string
	Error       string
	Success     string
}
