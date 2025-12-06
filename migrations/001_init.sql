-- Расширение для UUID (опционально)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица галактик
CREATE TABLE galaxies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50), -- spiral, elliptical, irregular
    diameter_ly BIGINT, -- диаметр в световых годах
    mass_suns NUMERIC(30, 2), -- масса в массах Солнца
    distance_from_earth_ly NUMERIC(20, 2), -- расстояние от Земли
    discovered_year INT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица планет
CREATE TABLE planets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    galaxy_id INT REFERENCES galaxies(id) ON DELETE SET NULL,
    type VARCHAR(50), -- terrestrial, gas_giant, ice_giant, dwarf
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

-- Индексы для ускорения поиска
CREATE INDEX idx_planets_galaxy_id ON planets(galaxy_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_galaxies_updated_at BEFORE UPDATE ON galaxies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_planets_updated_at BEFORE UPDATE ON planets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Создание администратора (пароль будет захеширован в коде)
-- INSERT INTO users (username, email, password_hash, role)
-- VALUES ('admin', 'admin@cosmos.com', 'хеш_пароля', 'admin');
