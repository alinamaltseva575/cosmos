package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

// AdminPlanetsHandler - список планет в админке
func (h *Handler) AdminPlanetsHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Получаем планеты из БД
	query := `
        SELECT p.id, p.name, p.type, p.diameter_km, p.has_life,
               COALESCE(g.name, 'Не указана') as galaxy_name
        FROM planets p
        LEFT JOIN galaxies g ON p.galaxy_id = g.id
        ORDER BY p.id DESC
    `

	rows, err := h.DB.Query(query)
	if err != nil {
		log.Printf("❌ Ошибка SQL запроса планет (админка): %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var planets []models.Planet
	for rows.Next() {
		var p models.Planet
		err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.DiameterKm, &p.HasLife, &p.GalaxyName)
		if err != nil {
			log.Printf("❌ Ошибка сканирования планеты (админка): %v", err)
			continue
		}
		planets = append(planets, p)
	}

	// Получаем общее количество
	var planetCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM planets").Scan(&planetCount)

	// Получаем сообщение об успехе из URL параметра
	success := r.URL.Query().Get("success") // ВОТ ТАК ДОБАВИТЬ

	data := models.PageData{
		Title:       "Управление планетами",
		CurrentPage: "admin_planets",
		Planets:     planets,
		PlanetCount: planetCount,
		IsAdmin:     true,
		Success:     success, // ВОТ ТАК ДОБАВИТЬ
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_planets: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminNewPlanetHandler - форма создания новой планеты
func (h *Handler) AdminNewPlanetHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Получаем список галактик для выпадающего списка
	galaxies, err := h.getGalaxies()
	if err != nil {
		log.Printf("❌ Ошибка получения галактик: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Создаем структуру для данных формы
	type FormData struct {
		models.PageData
		Planet   models.Planet
		Galaxies []models.Galaxy
		Error    string
	}

	data := FormData{
		PageData: models.PageData{
			Title:       "Добавление планеты",
			CurrentPage: "admin_planet_form",
			IsAdmin:     true,
		},
		Planet:   models.Planet{},
		Galaxies: galaxies,
	}

	// Обработка POST запроса
	if r.Method == http.MethodPost {
		planet, err := h.parsePlanetForm(r)
		if err != nil {
			data.Error = err.Error()
			data.Planet = planet
		} else {
			// Сохраняем в БД
			err = h.savePlanet(&planet)
			if err != nil {
				data.Error = "Ошибка сохранения в базу данных"
				data.Planet = planet
			} else {
				http.Redirect(w, r, "/admin/planets", http.StatusFound)
				return
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_planet_form: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminDeletePlanetHandler - удаление планеты
func (h *Handler) AdminDeletePlanetHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL: /admin/planets/delete/{id}
	// pathParts: ["", "admin", "planets", "delete", "{id}"]
	pathParts := strings.Split(r.URL.Path, "/")

	if len(pathParts) != 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[4]) // pathParts[4] это ID
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Если GET запрос - показываем страницу подтверждения
	if r.Method == http.MethodGet {
		h.showDeletePlanetConfirmation(w, r, id)
		return
	}

	// Если POST запрос - выполняем удаление
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем имя планеты для логирования
	var planetName string
	h.DB.QueryRow("SELECT name FROM planets WHERE id = $1", id).Scan(&planetName)

	// Удаляем планету
	result, err := h.DB.Exec("DELETE FROM planets WHERE id = $1", id)
	if err != nil {
		log.Printf("❌ Ошибка удаления планеты %d: %v", id, err)

		if strings.Contains(err.Error(), "foreign key constraint") {
			http.Error(w, "Нельзя удалить планету, так как она связана с другими данными", http.StatusBadRequest)
		} else {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	log.Printf("✅ Планета удалена: %s (ID %d)", planetName, id)

	// Редирект с сообщением об успехе
	http.Redirect(w, r, "/admin/planets?success=Планета+"+planetName+"+удалена", http.StatusFound)
}

// НОВАЯ функция для страницы подтверждения
func (h *Handler) showDeletePlanetConfirmation(w http.ResponseWriter, r *http.Request, id int) {
	h.setEncoding(w)

	// Получаем планету из БД
	var planet models.Planet
	var galaxyName sql.NullString

	err := h.DB.QueryRow(`
        SELECT p.id, p.name, p.type, p.diameter_km, COALESCE(g.name, 'Не указана') as galaxy_name
        FROM planets p
        LEFT JOIN galaxies g ON p.galaxy_id = g.id
        WHERE p.id = $1
    `, id).Scan(&planet.ID, &planet.Name, &planet.Type, &planet.DiameterKm, &galaxyName)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("❌ Ошибка получения планеты для удаления: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	if galaxyName.Valid {
		planet.GalaxyName = galaxyName.String
	}

	// Структура для данных страницы подтверждения
	type DeleteData struct {
		models.PageData
		ObjectType string
		ObjectName string
		ObjectData interface{}
		DeleteURL  string
		ReturnURL  string
	}

	data := DeleteData{
		PageData: models.PageData{
			Title:       "Подтверждение удаления планеты",
			CurrentPage: "admin_confirm_delete",
			IsAdmin:     true,
		},
		ObjectType: "Планета",
		ObjectName: planet.Name,
		ObjectData: planet,
		DeleteURL:  "/admin/planets/" + strconv.Itoa(id) + "/delete",
		ReturnURL:  "/admin/planets",
	}

	err = h.Tmpl.ExecuteTemplate(w, "admin_confirm_delete.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_confirm_delete: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// ========== Вспомогательные методы ==========

func (h *Handler) getGalaxies() ([]models.Galaxy, error) {
	rows, err := h.DB.Query("SELECT id, name FROM galaxies ORDER BY name")
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

func (h *Handler) parsePlanetForm(r *http.Request) (models.Planet, error) {
	var planet models.Planet

	// Парсим форму
	planet.Name = r.FormValue("name")
	planet.Type = r.FormValue("type")
	planet.Description = r.FormValue("description")

	log.Printf("📝 Парсим форму: name=%s, type=%s", planet.Name, planet.Type)

	// Проверяем обязательные поля
	if planet.Name == "" {
		return planet, errors.New("название планеты обязательно")
	}
	if planet.Type == "" {
		return planet, errors.New("тип планеты обязателен")
	}
	if planet.Description == "" {
		return planet, errors.New("описание обязательно")
	}

	// Числовые поля
	if diameter := r.FormValue("diameter_km"); diameter != "" {
		if val, err := strconv.ParseFloat(diameter, 64); err == nil {
			planet.DiameterKm = val
		} else {
			log.Printf("⚠️ Ошибка парсинга diameter_km: %v", err)
		}
	}

	if mass := r.FormValue("mass_kg"); mass != "" {
		if val, err := strconv.ParseFloat(mass, 64); err == nil {
			planet.MassKg = val
		} else {
			log.Printf("⚠️ Ошибка парсинга mass_kg: %v", err)
		}
	}

	if period := r.FormValue("orbital_period_days"); period != "" {
		if val, err := strconv.ParseFloat(period, 64); err == nil {
			planet.OrbitalPeriodDays = val
		} else {
			log.Printf("⚠️ Ошибка парсинга orbital_period_days: %v", err)
		}
	}

	if year := r.FormValue("discovered_year"); year != "" {
		if val, err := strconv.Atoi(year); err == nil {
			planet.DiscoveredYear = &val
		} else {
			log.Printf("⚠️ Ошибка парсинга discovered_year: %v", err)
		}
	}

	// Galaxy ID
	if galaxyID := r.FormValue("galaxy_id"); galaxyID != "" {
		if val, err := strconv.Atoi(galaxyID); err == nil {
			planet.GalaxyID = &val
		} else {
			log.Printf("⚠️ Ошибка парсинга galaxy_id: %v", err)
		}
	}

	// Checkboxes
	planet.HasLife = r.FormValue("has_life") == "on" || r.FormValue("has_life") == "true"
	planet.IsHabitable = r.FormValue("is_habitable") == "on" || r.FormValue("is_habitable") == "true"

	log.Printf("📊 Результат парсинга: %+v", planet)

	return planet, nil
}

func (h *Handler) savePlanet(planet *models.Planet) error {
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

	err := h.DB.QueryRow(query,
		planet.Name, planet.Type, planet.Description,
		planet.DiameterKm, planet.MassKg, planet.OrbitalPeriodDays,
		discoveredYear, galaxyID,
		planet.HasLife, planet.IsHabitable,
	).Scan(&planet.ID, &planet.CreatedAt)

	return err
}

func (h *Handler) updatePlanet(id int, planet *models.Planet) error {
	query := `
        UPDATE planets
        SET name = $1, type = $2, description = $3, diameter_km = $4,
            mass_kg = $5, orbital_period_days = $6, discovered_year = $7,
            galaxy_id = $8, has_life = $9, is_habitable = $10,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $11
        RETURNING updated_at
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

	err := h.DB.QueryRow(query,
		planet.Name, planet.Type, planet.Description,
		planet.DiameterKm, planet.MassKg, planet.OrbitalPeriodDays,
		discoveredYear, galaxyID,
		planet.HasLife, planet.IsHabitable,
		id,
	).Scan(&planet.UpdatedAt)

	return err
}

// AdminConfirmDeletePlanetHandler - страница подтверждения удаления планеты
func (h *Handler) AdminConfirmDeletePlanetHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 5 || pathParts[3] != "delete" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[2])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Получаем планету из БД
	var planet models.Planet
	var galaxyName sql.NullString

	err = h.DB.QueryRow(`
        SELECT p.id, p.name, p.type, p.diameter_km, COALESCE(g.name, 'Не указана') as galaxy_name
        FROM planets p
        LEFT JOIN galaxies g ON p.galaxy_id = g.id
        WHERE p.id = $1
    `, id).Scan(&planet.ID, &planet.Name, &planet.Type, &planet.DiameterKm, &galaxyName)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("❌ Ошибка получения планеты для удаления: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	if galaxyName.Valid {
		planet.GalaxyName = galaxyName.String
	}

	// Структура для данных страницы подтверждения
	type DeleteData struct {
		models.PageData
		ObjectType string
		ObjectName string
		ObjectData interface{}
		DeleteURL  string
		ReturnURL  string
	}

	data := DeleteData{
		PageData: models.PageData{
			Title:       "Подтверждение удаления планеты",
			CurrentPage: "admin_confirm_delete",
			IsAdmin:     true,
		},
		ObjectType: "Планета",
		ObjectName: planet.Name,
		ObjectData: planet,
		DeleteURL:  "/admin/planets/" + strconv.Itoa(id) + "/delete",
		ReturnURL:  "/admin/planets",
	}

	err = h.Tmpl.ExecuteTemplate(w, "admin_confirm_delete.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_confirm_delete: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminEditPlanetHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 5 {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(pathParts[4])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Получаем список галактик для выпадающего списка
	galaxies, err := h.getGalaxies()
	if err != nil {
		log.Printf("❌ Ошибка получения галактик: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Структура для данных формы
	type FormData struct {
		models.PageData
		Planet   models.Planet
		Galaxies []models.Galaxy
		Error    string
		Success  string
	}

	// Получаем планету из БД
	var planet models.Planet
	var discoveredYear sql.NullInt64
	var galaxyID sql.NullInt64
	var massKg sql.NullFloat64
	var orbitalPeriodDays sql.NullFloat64

	err = h.DB.QueryRow(`
        SELECT id, name, type, description, diameter_km, mass_kg,
               orbital_period_days, discovered_year, galaxy_id,
               has_life, is_habitable, created_at
        FROM planets
        WHERE id = $1
    `, id).Scan(
		&planet.ID, &planet.Name, &planet.Type, &planet.Description,
		&planet.DiameterKm, &massKg, &orbitalPeriodDays,
		&discoveredYear, &galaxyID, &planet.HasLife, &planet.IsHabitable,
		&planet.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("❌ Ошибка получения планеты: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
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

	data := FormData{
		PageData: models.PageData{
			Title:       "Редактирование планеты",
			CurrentPage: "admin_planet_form",
			IsAdmin:     true,
		},
		Planet:   planet,
		Galaxies: galaxies,
	}

	// Обработка POST запроса (обновление)
	if r.Method == http.MethodPost {
		updatedPlanet, err := h.parsePlanetForm(r)
		if err != nil {
			data.Error = err.Error()
			data.Planet = updatedPlanet
			data.Planet.ID = planet.ID // Сохраняем оригинальный ID
		} else {
			// Обновляем в БД
			err = h.updatePlanet(id, &updatedPlanet)
			if err != nil {
				data.Error = "Ошибка обновления в базе данных: " + err.Error()
				data.Planet = updatedPlanet
				data.Planet.ID = planet.ID
			} else {
				data.Success = "Планета успешно обновлена!"
				// Обновляем данные для отображения
				data.Planet = updatedPlanet
				data.Planet.ID = planet.ID
				data.Planet.CreatedAt = planet.CreatedAt
				log.Printf("✅ Планета обновлена: ID %d", id)
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_planet_form (edit): %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}
