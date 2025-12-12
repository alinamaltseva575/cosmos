package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cosmos/internal/models"
)

// AdminGalaxiesHandler - список галактик в админке
func (h *Handler) AdminGalaxiesHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Получаем галактики из БД
	rows, err := h.DB.Query(`
        SELECT id, name, type, diameter_ly, discovered_year
        FROM galaxies
        ORDER BY id DESC
    `)
	if err != nil {
		log.Printf("❌ Ошибка SQL запроса галактик (админка): %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var galaxies []models.Galaxy
	for rows.Next() {
		var g models.Galaxy
		var diameterLy sql.NullFloat64
		var discoveredYear sql.NullInt64

		err := rows.Scan(&g.ID, &g.Name, &g.Type, &diameterLy, &discoveredYear)
		if err != nil {
			log.Printf("❌ Ошибка сканирования галактики (админка): %v", err)
			continue
		}

		if diameterLy.Valid {
			g.DiameterLy = &diameterLy.Float64
		}
		if discoveredYear.Valid {
			year := int(discoveredYear.Int64)
			g.DiscoveredYear = &year
		}

		galaxies = append(galaxies, g)
	}

	// Получаем общее количество
	var galaxyCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM galaxies").Scan(&galaxyCount)

	// Получаем сообщение об успехе из URL параметра
	success := r.URL.Query().Get("success") // ВОТ ТАК ДОБАВИТЬ

	data := models.PageData{
		Title:       "Управление галактиками",
		CurrentPage: "admin_galaxies",
		Galaxies:    galaxies,
		GalaxyCount: galaxyCount,
		IsAdmin:     true,
		Success:     success, // ВОТ ТАК ДОБАВИТЬ
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_galaxies: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminNewGalaxyHandler - форма создания новой галактики
func (h *Handler) AdminNewGalaxyHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Создаем структуру для данных формы
	type FormData struct {
		models.PageData
		Galaxy models.Galaxy
		Error  string
	}

	data := FormData{
		PageData: models.PageData{
			Title:       "Добавление галактики",
			CurrentPage: "admin_galaxy_form",
			IsAdmin:     true,
		},
		Galaxy: models.Galaxy{},
	}

	// Обработка POST запроса
	if r.Method == http.MethodPost {
		galaxy, err := h.parseGalaxyForm(r)
		if err != nil {
			data.Error = err.Error()
			data.Galaxy = galaxy
		} else {
			// Сохраняем в БД
			err = h.saveGalaxy(&galaxy)
			if err != nil {
				data.Error = "Ошибка сохранения в базу данных"
				data.Galaxy = galaxy
			} else {
				http.Redirect(w, r, "/admin/galaxies", http.StatusFound)
				return
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_galaxy_form: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminEditGalaxyHandler - форма редактирования галактики
func (h *Handler) AdminEditGalaxyHandler(w http.ResponseWriter, r *http.Request) {
	h.setEncoding(w)

	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL: /admin/galaxies/edit/{id}
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

	// Структура для данных формы
	type FormData struct {
		models.PageData
		Galaxy models.Galaxy
		Error  string
	}

	// Получаем галактику из БД
	var galaxy models.Galaxy
	var diameterLy, massSuns, distanceFromEarthLy sql.NullFloat64
	var discoveredYear sql.NullInt64

	err = h.DB.QueryRow(`
		SELECT id, name, type, diameter_ly, mass_suns,
		       distance_from_earth_ly, discovered_year, description
		FROM galaxies
		WHERE id = $1
	`, id).Scan(
		&galaxy.ID, &galaxy.Name, &galaxy.Type, &diameterLy, &massSuns,
		&distanceFromEarthLy, &discoveredYear, &galaxy.Description,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("❌ Ошибка получения галактики: %v", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	// Обрабатываем nullable поля
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

	data := FormData{
		PageData: models.PageData{
			Title:       "Редактирование галактики",
			CurrentPage: "admin_galaxy_form",
			IsAdmin:     true,
		},
		Galaxy: galaxy,
	}

	// Обработка POST запроса (обновление)
	if r.Method == http.MethodPost {
		updatedGalaxy, err := h.parseGalaxyForm(r)
		if err != nil {
			data.Error = err.Error()
			data.Galaxy = updatedGalaxy
			data.Galaxy.ID = galaxy.ID // Сохраняем оригинальный ID
		} else {
			// Обновляем в БД
			err = h.updateGalaxy(id, &updatedGalaxy)
			if err != nil {
				data.Error = "Ошибка обновления в базе данных"
				data.Galaxy = updatedGalaxy
				data.Galaxy.ID = galaxy.ID
			} else {
				http.Redirect(w, r, "/admin/galaxies", http.StatusFound)
				return
			}
		}
	}

	err = h.Tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_galaxy_form (edit): %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// AdminDeleteGalaxyHandler - удаление галактики
func (h *Handler) AdminDeleteGalaxyHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию
	_, err := h.requireAdminAuth(w, r)
	if err != nil {
		return
	}

	// Извлекаем ID из URL: /admin/galaxies/delete/{id}
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
		h.showDeleteGalaxyConfirmation(w, r, id)
		return
	}

	// Если POST запрос - выполняем удаление
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем имя галактики для логирования
	var galaxyName string
	h.DB.QueryRow("SELECT name FROM galaxies WHERE id = $1", id).Scan(&galaxyName)

	// Проверяем, есть ли зависимые планеты
	var planetCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM planets WHERE galaxy_id = $1", id).Scan(&planetCount)

	if planetCount > 0 {
		http.Error(w, "Нельзя удалить галактику, у которой есть планеты. Сначала удалите или переместите планеты.", http.StatusBadRequest)
		return
	}

	// Удаляем галактику
	result, err := h.DB.Exec("DELETE FROM galaxies WHERE id = $1", id)
	if err != nil {
		log.Printf("❌ Ошибка удаления галактики %d: %v", id, err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	log.Printf("✅ Галактика удалена: %s (ID %d)", galaxyName, id)

	http.Redirect(w, r, "/admin/galaxies?success=Галактика+"+galaxyName+"+удалена", http.StatusFound)
}

// Добавить вспомогательную функцию
func (h *Handler) showDeleteGalaxyConfirmation(w http.ResponseWriter, r *http.Request, id int) {
	log.Printf("🔍 showDeleteGalaxyConfirmation вызван для ID: %d", id)

	h.setEncoding(w)

	// Получаем галактику из БД
	var galaxy models.Galaxy
	var diameterLy sql.NullFloat64

	err := h.DB.QueryRow(`
        SELECT id, name, type, diameter_ly, description
        FROM galaxies
        WHERE id = $1
    `, id).Scan(&galaxy.ID, &galaxy.Name, &galaxy.Type, &diameterLy, &galaxy.Description)

	if err != nil {
		log.Printf("❌ Ошибка получения галактики для удаления: %v", err)
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		}
		return
	}

	if diameterLy.Valid {
		val := diameterLy.Float64
		galaxy.DiameterLy = &val
	}

	// Проверяем, есть ли планеты в этой галактике
	var planetCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM planets WHERE galaxy_id = $1", id).Scan(&planetCount)

	// Структура для данных страницы подтверждения
	type DeleteData struct {
		models.PageData
		ObjectType  string
		ObjectName  string
		ObjectData  interface{}
		DeleteURL   string
		ReturnURL   string
		HasPlanets  bool
		PlanetCount int
	}

	data := DeleteData{
		PageData: models.PageData{
			Title:       "Подтверждение удаления галактики",
			CurrentPage: "admin_confirm_delete",
			IsAdmin:     true,
		},
		ObjectType:  "Галактика",
		ObjectName:  galaxy.Name,
		ObjectData:  galaxy,
		DeleteURL:   "/admin/galaxies/delete/" + strconv.Itoa(id),
		ReturnURL:   "/admin/galaxies",
		HasPlanets:  planetCount > 0,
		PlanetCount: planetCount,
	}

	log.Printf("📊 Данные для шаблона галактики: ObjectType=%s, ObjectName=%s, HasPlanets=%v",
		data.ObjectType, data.ObjectName, data.HasPlanets)

	// Пробуем выполнить шаблон
	err = h.Tmpl.ExecuteTemplate(w, "admin_confirm_delete", data)
	if err != nil {
		log.Printf("❌ Ошибка выполнения шаблона admin_confirm_delete для галактики: %v", err)

		// Покажем простую страницу ошибки
		fmt.Fprintf(w, `
            <html><body style="background:#0a0a2a;color:white;padding:50px;">
            <h1>Ошибка загрузки шаблона</h1>
            <p>%v</p>
            <p>ObjectType: %s</p>
            <p>ObjectName: %s</p>
            <p>HasPlanets: %v</p>
            <a href="/admin/galaxies">Назад к галактикам</a>
            </body></html>
        `, err, data.ObjectType, data.ObjectName, data.HasPlanets)
	}
}

// ========== Вспомогательные методы для галактик ==========

func (h *Handler) parseGalaxyForm(r *http.Request) (models.Galaxy, error) {
	var galaxy models.Galaxy

	// Парсим форму
	galaxy.Name = r.FormValue("name")
	galaxy.Type = r.FormValue("type")
	galaxy.Description = r.FormValue("description")

	// Проверяем обязательные поля
	if galaxy.Name == "" {
		return galaxy, errors.New("название галактики обязательно")
	}
	if galaxy.Type == "" {
		return galaxy, errors.New("тип галактики обязателен")
	}
	if galaxy.Description == "" {
		return galaxy, errors.New("описание обязательно")
	}

	// Числовые поля
	if diameter := r.FormValue("diameter_ly"); diameter != "" {
		if val, err := strconv.ParseFloat(diameter, 64); err == nil {
			galaxy.DiameterLy = &val
		}
	}

	if mass := r.FormValue("mass_suns"); mass != "" {
		if val, err := strconv.ParseFloat(mass, 64); err == nil {
			galaxy.MassSuns = &val
		}
	}

	if distance := r.FormValue("distance_from_earth_ly"); distance != "" {
		if val, err := strconv.ParseFloat(distance, 64); err == nil {
			galaxy.DistanceFromEarthLy = &val
		}
	}

	if year := r.FormValue("discovered_year"); year != "" {
		if val, err := strconv.Atoi(year); err == nil {
			galaxy.DiscoveredYear = &val
		}
	}

	return galaxy, nil
}

func (h *Handler) saveGalaxy(galaxy *models.Galaxy) error {
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

	err := h.DB.QueryRow(query,
		galaxy.Name, galaxy.Type, galaxy.Description,
		diameterLy, massSuns, distanceFromEarthLy, discoveredYear,
	).Scan(&galaxy.ID, &galaxy.CreatedAt)

	return err
}

func (h *Handler) updateGalaxy(id int, galaxy *models.Galaxy) error {
	query := `
		UPDATE galaxies
		SET name = $1, type = $2, description = $3, diameter_ly = $4,
		    mass_suns = $5, distance_from_earth_ly = $6, discovered_year = $7,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING updated_at
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

	err := h.DB.QueryRow(query,
		galaxy.Name, galaxy.Type, galaxy.Description,
		diameterLy, massSuns, distanceFromEarthLy, discoveredYear,
		id,
	).Scan(&galaxy.CreatedAt) // Используем CreatedAt для updated_at

	return err
}
