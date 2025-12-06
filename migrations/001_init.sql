-- Создание таблиц
CREATE TABLE IF NOT EXISTS galaxies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50),
    diameter_ly BIGINT,
    mass_suns NUMERIC(30, 2),
    distance_from_earth_ly NUMERIC(20, 2),
    discovered_year INT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
DROP TABLE IF EXISTS planets CASCADE;

-- Вставка тестовых данных для галактик
INSERT INTO galaxies (name, type, diameter_ly, mass_suns, distance_from_earth_ly, discovered_year, description) VALUES
('Млечный Путь', 'спиральная', 100000, 1500000000000, 0, -1, 'Наша родная галактика, содержащая Солнечную систему.'),
('Андромеда', 'спиральная', 220000, 1200000000000, 2537000, 964, 'Ближайшая к Млечному Пути крупная галактика.'),
('Треугольник', 'спиральная', 60000, 50000000000, 3000000, 1654, 'Третья по величине галактика в Местной группе.'),
('Сомбреро', 'спиральная', 50000, 800000000000, 29000000, 1781, 'Галактика в созвездии Девы, известная своим ярким ядром.'),
('Сигара', 'неправильная', 37000, 30000000000, 12000000, 1774, 'Галактика со вспышкой звездообразования в созвездии Большой Медведицы.');

CREATE TABLE IF NOT EXISTS planets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    galaxy_id INT REFERENCES galaxies(id) ON DELETE SET NULL,
    type VARCHAR(50),
    diameter_km NUMERIC(12, 2),
    mass_kg NUMERIC(30, 2),
    orbital_period_days NUMERIC(10, 2),
    has_life BOOLEAN DEFAULT false,
    is_habitable BOOLEAN DEFAULT false,
    discovered_year INT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Вставка тестовых данных для планет
INSERT INTO planets (name, galaxy_id, type, diameter_km, mass_kg, orbital_period_days, has_life, is_habitable, discovered_year, description) VALUES
('Земля', 1, 'землеподобная', 12742, 5.972e24, 365.25, true, true, -1, 'Третья планета от Солнца, единственная известная планета с жизнью.'),
('Марс', 1, 'землеподобная', 6779, 6.39e23, 687, false, true, -1, 'Красная планета, четвертая от Солнца. Имеет два спутника.'),
('Юпитер', 1, 'газовый гигант', 139820, 1.898e27, 4333, false, false, -1, 'Крупнейшая планета Солнечной системы.'),
('Сатурн', 1, 'газовый гигант', 116460, 5.683e26, 10759, false, false, -1, 'Планета с ярко выраженной системой колец.'),
('Венера', 1, 'землеподобная', 12104, 4.867e24, 225, false, false, -1, 'Вторая планета от Солнца, самая горячая планета системы.'),
('Кеплер-186f', 2, 'землеподобная', 14800, 5.5e24, 130, true, true, 2014, 'Первая землеподобная планета в обитаемой зоне другой звезды.'),
('TRAPPIST-1e', 3, 'землеподобная', 10500, 4.0e24, 6.1, true, true, 2017, 'Планета в системе TRAPPIST-1, потенциально пригодная для жизни.'),
('HD 209458 b', 4, 'газовый гигант', 218000, 2.2e27, 3.5, false, false, 1999, 'Первая планета, обнаруженная методом транзита.');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание администратора (пароль: admin123)
INSERT INTO users (username, email, password_hash, role) VALUES
('admin', 'admin@cosmos.ru', '$2a$10$N9qo8uLOickgx2ZMRZoMye1G3YZ5QzYbhFgJYVVpQp6.6dQ2Z7W6y', 'admin');

-- Создание индексов для ускорения поиска
CREATE INDEX idx_planets_galaxy_id ON planets(galaxy_id);
CREATE INDEX idx_planets_name ON planets(name);
CREATE INDEX idx_galaxies_name ON galaxies(name);
