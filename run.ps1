# run.ps1 - скрипт для запуска проекта на Windows

Write-Host "=== Cosmos API Launcher ===" -ForegroundColor Green

# Проверка .env файла
if (-not (Test-Path ".env")) {
    Write-Host "Создаю .env файл из примера..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env" -ErrorAction SilentlyContinue
    if (-not (Test-Path ".env")) {
        Write-Host "Создайте .env файл вручную!" -ForegroundColor Red
        Write-Host "Скопируйте .env.example и настройте параметры" -ForegroundColor Yellow
        exit 1
    }
}

# Проверка зависимостей
Write-Host "Проверяю зависимости..." -ForegroundColor Cyan
go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "Go не установлен!" -ForegroundColor Red
    exit 1
}

# Загрузка переменных окружения
Write-Host "Загружаю настройки..." -ForegroundColor Cyan
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match '^([^=]+)=(.*)$') {
            [Environment]::SetEnvironmentVariable($matches[1], $matches[2], "Process")
        }
    }
}

# Проверка PostgreSQL
Write-Host "Проверяю подключение к PostgreSQL..." -ForegroundColor Cyan
$env:PGPASSWORD = $env:DB_PASSWORD
$pgTest = & psql -h $env:DB_HOST -p $env:DB_PORT -U $env:DB_USER -d $env:DB_NAME -c "SELECT 1;" 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка подключения к PostgreSQL!" -ForegroundColor Red
    Write-Host "Проверьте настройки в .env файле" -ForegroundColor Yellow
    Write-Host "Host: $($env:DB_HOST), Port: $($env:DB_PORT), User: $($env:DB_USER), DB: $($env:DB_NAME)" -ForegroundColor Yellow
    exit 1
}

# Установка зависимостей
Write-Host "Устанавливаю зависимости Go..." -ForegroundColor Cyan
go mod tidy

# Запуск приложения
Write-Host "Запускаю Cosmos API на порту $($env:APP_PORT)..." -ForegroundColor Green
Write-Host "API будет доступно по адресу: http://localhost:$($env:APP_PORT)" -ForegroundColor Green
Write-Host "Нажмите Ctrl+C для остановки" -ForegroundColor Yellow
go run cmd/api/main.go
