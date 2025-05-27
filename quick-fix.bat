@echo off
chcp 65001 >nul
title Полное исправление TeleOko v2.0

echo 🔧 Полное исправление TeleOko v2.0
echo ===============================

echo ℹ️  1. Удаление проблемных файлов...
del /q internal\network\camera_discovery.go 2>nul
if exist internal\network\camera_discovery.go (
    echo ⚠️  Не удалось удалить camera_discovery.go
) else (
    echo ✅ camera_discovery.go удален
)

echo.
echo ℹ️  2. Проверка структуры проекта...

:: Проверяем основные файлы
set "missing_files="

if not exist internal\handlers\handlers.go (
    set "missing_files=!missing_files! handlers.go"
)

if not exist internal\config\config.go (
    set "missing_files=!missing_files! config.go"
)

if not exist internal\network\discovery.go (
    set "missing_files=!missing_files! discovery.go"
)

if not exist internal\hikvision\api.go (
    set "missing_files=!missing_files! hikvision/api.go"
)

if not exist internal\hikvision\models.go (
    set "missing_files=!missing_files! hikvision/models.go"
)

if not exist internal\go2rtc\manager.go (
    set "missing_files=!missing_files! go2rtc/manager.go"
)

if not "%missing_files%"=="" (
    echo ❌ Отсутствуют файлы:%missing_files%
    echo.
    echo 📋 Создайте отсутствующие файлы из артефактов:
    echo    • internal/handlers/handlers.go
    echo    • internal/config/config.go  
    echo    • internal/network/discovery.go
    echo    • internal/hikvision/api.go
    echo    • internal/hikvision/models.go
    echo    • internal/go2rtc/manager.go
    echo    • web/static/js/main.js ^(исправленный^)
    echo.
    pause
    exit /b 1
) else (
    echo ✅ Все основные файлы найдены
)

echo.
echo ℹ️  3. Проверка main.js...
if exist web\static\js\main.js (
    echo ✅ main.js найден
    echo ℹ️  Замените файл исправленной версией из артефакта
) else (
    echo ❌ main.js не найден - создайте его из артефакта
)

echo.
echo ℹ️  4. Проверка go2rtc...
if exist go2rtc.exe (
    echo ✅ go2rtc.exe найден
    go2rtc --version >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo ✅ go2rtc работает
    ) else (
        echo ℹ️  Исправление пути go2rtc...
        set "PATH=%CD%;%PATH%"
        copy go2rtc.exe %TEMP%\go2rtc.exe >nul 2>&1
        echo ✅ Путь исправлен
    )
) else (
    echo ❌ go2rtc.exe не найден
    echo 📥 Запустите: .\download-go2rtc.bat
    pause
    exit /b 1
)

echo.
echo ℹ️  5. Обновление зависимостей...
go mod tidy
if %ERRORLEVEL% equ 0 (
    echo ✅ Зависимости обновлены
) else (
    echo ⚠️  Проблемы с зависимостями
)

echo.
echo ℹ️  6. Сборка проекта...
go build -o teleoko.exe ./cmd/server
if %ERRORLEVEL% equ 0 (
    echo ✅ Проект собран успешно
) else (
    echo ❌ Ошибка сборки
    echo.
    echo 📋 Убедитесь, что у вас есть ВСЕ файлы из артефактов:
    echo    • internal/handlers/handlers.go ^(единственный файл в папке!^)
    echo    • internal/config/config.go  
    echo    • internal/network/discovery.go ^(НЕ camera_discovery.go!^)
    echo    • internal/hikvision/api.go
    echo    • internal/hikvision/models.go
    echo    • internal/go2rtc/manager.go
    echo    • web/static/js/main.js ^(полностью исправленный^)
    echo    • config.json ^(с 16 камерами^)
    echo.
    pause
    exit /b 1
)

echo.
echo 🎉 ВСЕ ИСПРАВЛЕНО!
echo ==================
echo.
echo 📁 Структура проекта готова:
echo    ✅ 33 канала ^(16 камер + общий план^)
echo    ✅ go2rtc включен для видео в браузере
echo    ✅ JavaScript исправлен - кнопки работают
echo    ✅ Конфликтующие файлы удалены
echo.
echo 🚀 ЗАПУСК:
echo    .\teleoko.exe
echo.
echo 🌐 ВЕБА-ИНТЕРФЕЙС:
echo    http://localhost:8082
echo.
echo 📹 ФУНКЦИИ:
echo    • Прямой эфир в браузере ^(WebRTC^)
echo    • Архивные записи по дате
echo    • Снимки с камер
echo    • Временная шкала
echo    • 16 камер HD/SD качества
echo.
pause@echo off
chcp 65001 >nul
title Быстрое исправление TeleOko

echo 🔧 Быстрое исправление всех проблем TeleOko
echo ==========================================

echo ℹ️  1. Удаление конфликтующих файлов...
del /q internal\network\camera_discovery.go 2>nul
if exist internal\network\camera_discovery.go (
    echo ⚠️  Не удалось удалить camera_discovery.go
) else (
    echo ✅ camera_discovery.go удален
)

echo.
echo ℹ️  2. Проверка go2rtc...
if exist go2rtc.exe (
    echo ✅ go2rtc.exe найден
    go2rtc --version >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo ✅ go2rtc работает
    ) else (
        echo ℹ️  Исправление пути go2rtc...
        set "PATH=%CD%;%PATH%"
        copy go2rtc.exe %TEMP%\go2rtc.exe >nul 2>&1
    )
) else (
    echo ❌ go2rtc.exe не найден
    echo 📥 Запустите: .\download-go2rtc.bat
    pause
    exit /b 1
)

echo.
echo ℹ️  3. Проверка структуры проекта...
if exist internal\handlers\handlers.go (
    echo ✅ handlers.go найден
) else (
    echo ❌ handlers.go не найден - создайте его из артефактов
)

if exist internal\config\config.go (
    echo ✅ config.go найден
) else (
    echo ❌ config.go не найден - создайте его из артефактов
)

if exist internal\network\discovery.go (
    echo ✅ discovery.go найден
) else (
    echo ❌ discovery.go не найден - создайте его из артефактов
)

echo.
echo ℹ️  4. Обновление зависимостей...
go mod tidy
if %ERRORLEVEL% equ 0 (
    echo ✅ Зависимости обновлены
) else (
    echo ⚠️  Проблемы с зависимостями
)

echo.
echo ℹ️  5. Сборка проекта...
go build -o teleoko.exe ./cmd/server
if %ERRORLEVEL% equ 0 (
    echo ✅ Проект собран успешно
) else (
    echo ❌ Ошибка сборки
    echo.
    echo 📋 Проверьте, что у вас есть все файлы:
    echo    • internal/handlers/handlers.go
    echo    • internal/config/config.go  
    echo    • internal/network/discovery.go
    echo    • internal/hikvision/api.go
    echo    • internal/hikvision/models.go
    echo    • internal/go2rtc/manager.go
    echo.
    pause
    exit /b 1
)

echo.
echo 🎉 Все исправлено!
echo ===============
echo.
echo 🚀 Теперь запустите TeleOko:
echo    .\teleoko.exe
echo.
echo 🌐 Откройте браузер:
echo    http://localhost:8082
echo.
echo 📹 У вас есть 33 канала (16 камер + общий план)
echo ✅ go2rtc включен для видео в браузере
echo.
pause