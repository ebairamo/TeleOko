#!/bin/bash

# TeleOko v2.0 - Скрипт запуска
# ================================

set -e  # Остановка при ошибках

echo "🚀 Запуск TeleOko v2.0 - Система видеонаблюдения"
echo "=================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Проверка Go
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go не установлен! Установите Go 1.22+ с https://golang.org/"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_status "Go версия: $GO_VERSION"
}

# Проверка сетевых портов
check_ports() {
    local ports=(8082 1984)
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_warning "Порт $port уже используется"
            
            if [ "$port" = "8082" ]; then
                print_error "Порт 8082 (веб-интерфейс) занят! Остановите другие службы или измените порт в config.json"
                exit 1
            fi
        else
            print_status "Порт $port свободен"
        fi
    done
}

# Создание конфигурации по умолчанию
create_default_config() {
    if [ ! -f "config.json" ]; then
        print_info "Создание config.json с настройками по умолчанию..."
        
        cat > config.json << 'EOF'
{
    "server": {
        "port": 8082
    },
    "hikvision": {
        "ip": "192.168.8.5",
        "username": "admin",
        "password": "oborotni2447",
        "port": 554
    },
    "go2rtc": {
        "port": 1984,
        "enabled": true
    },
    "auth": {
        "enabled": false,
        "username": "admin",
        "password": "password"
    },
    "channels": [
        {
            "id": "1",
            "name": "Общий план",
            "url": "rtsp://admin:oborotni2447@192.168.8.5:554/Streaming/Channels/1"
        },
        {
            "id": "201",
            "name": "Камера 1 (HD)",
            "url": "rtsp://admin:oborotni2447@192.168.8.5:554/Streaming/Channels/201"
        },
        {
            "id": "202",
            "name": "Камера 1 (SD)",
            "url": "rtsp://admin:oborotni2447@192.168.8.5:554/Streaming/Channels/202"
        }
    ]
}
EOF
        print_status "Файл config.json создан"
        print_warning "Отредактируйте config.json для настройки ваших камер!"
    else
        print_status "Конфигурация найдена: config.json"
    fi
}

# Сборка приложения
build_app() {
    print_info "Проверка необходимости сборки..."
    
    # Проверяем, есть ли бинарник и актуален ли он
    if [ -f "teleoko" ]; then
        # Проверяем время модификации исходников
        NEWEST_SOURCE=$(find . -name "*.go" -newer "teleoko" 2>/dev/null | head -1)
        if [ -z "$NEWEST_SOURCE" ]; then
            print_status "Бинарник актуален, сборка не требуется"
            return 0
        fi
    fi
    
    print_info "Сборка приложения..."
    
    # Загружаем зависимости
    print_info "Загрузка зависимостей..."
    go mod download
    
    # Собираем
    print_info "Компиляция..."
    if go build -ldflags="-s -w" -o teleoko ./cmd/server; then
        print_status "Приложение успешно собрано"
    else
        print_error "Ошибка сборки приложения!"
        exit 1
    fi
    
    # Устанавливаем права на выполнение
    chmod +x teleoko
}

# Проверка подключения к камере
test_camera_connection() {
    print_info "Тестирование подключения к камере..."
    
    # Читаем IP из конфигурации
    if command -v jq &> /dev/null; then
        CAMERA_IP=$(jq -r '.hikvision.ip' config.json 2>/dev/null || echo "192.168.8.5")
    else
        CAMERA_IP="192.168.8.5"  # По умолчанию
    fi
    
    if ping -c 1 -W 3 "$CAMERA_IP" &> /dev/null; then
        print_status "Камера $CAMERA_IP доступна"
    else
        print_warning "Камера $CAMERA_IP недоступна"
        print_info "Убедитесь, что:"
        print_info "  • Камера включена и подключена к сети"
        print_info "  • IP-адрес в config.json правильный"
        print_info "  • Нет блокировки файрволом"
    fi
}

# Отображение информации о запуске
show_startup_info() {
    local LOCAL_IP=$(hostname -I | awk '{print $1}' 2>/dev/null || echo "localhost")
    
    echo ""
    echo "🎉 TeleOko v2.0 готов к запуску!"
    echo "================================"
    echo ""
    echo "📱 Веб-интерфейс будет доступен по адресам:"
    echo "   • Локально:  http://localhost:8082"
    echo "   • По сети:   http://$LOCAL_IP:8082"
    echo ""
    echo "🔧 Конфигурация:"
    echo "   • Файл:      config.json"
    echo "   • go2rtc:    http://localhost:1984"
    echo ""
    echo "📖 Полезные команды:"
    echo "   • Остановить: Ctrl+C"
    echo "   • Логи:       tail -f teleoko.log"
    echo "   • Статус:     curl http://localhost:8082/api/info"
    echo ""
    print_info "Запуск сервера..."
    echo ""
}

# Основная функция
main() {
    echo ""
    print_info "Проверка системных требований..."
    check_go
    check_ports
    
    echo ""
    print_info "Подготовка конфигурации..."
    create_default_config
    
    echo ""
    print_info "Сборка приложения..."
    build_app
    
    echo ""
    test_camera_connection
    
    echo ""
    show_startup_info
    
    # Запуск приложения с логированием
    if [ -t 1 ]; then
        # Интерактивный режим
        ./teleoko 2>&1 | tee teleoko.log
    else
        # Неинтерактивный режим (например, systemd)
        ./teleoko >> teleoko.log 2>&1
    fi
}

# Обработка сигналов завершения
trap 'echo -e "\n👋 Завершение TeleOko..."; exit 0' INT TERM

# Запуск
main "$@"