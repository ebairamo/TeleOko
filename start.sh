#!/bin/bash

# TeleOko v2.0 - Скрипт запуска с поддержкой ngrok
# ================================

set -e  # Остановка при ошибках

echo "🚀 Запуск TeleOko v2.0 - Система видеонаблюдения с поддержкой ngrok"
echo "=================================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для цветного вывода
print_success() {
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
    print_success "Go версия: $GO_VERSION"
}

# Проверка ngrok
check_ngrok() {
    if ! command -v ngrok &> /dev/null; then
        print_warning "ngrok не установлен! Установите ngrok с https://ngrok.com/"
        print_warning "Внешний доступ через интернет будет недоступен."
        USE_NGROK=false
    else
        print_success "ngrok найден и будет использован для внешнего доступа"
        USE_NGROK=true
    fi
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
            elif [ "$port" = "1984" ]; then
                print_error "Порт 1984 (go2rtc) занят! Остановите другие службы или измените порт в config.json"
                exit 1
            fi
        else
            print_success "Порт $port свободен"
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
        "ip": "192.168.8.10",
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
            "name": "🎥 Общий план",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1"
        },
        {
            "id": "102",
            "name": "📹 Камера 1",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/102"
        },
        {
            "id": "202",
            "name": "📹 Камера 2",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/202"
        },
        {
            "id": "302",
            "name": "📹 Камера 3",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/302"
        },
        {
            "id": "402",
            "name": "📹 Камера 4",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/402"
        },
        {
            "id": "502",
            "name": "📹 Камера 5",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/502"
        },
        {
            "id": "602",
            "name": "📹 Камера 6",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/602"
        },
        {
            "id": "702",
            "name": "📹 Камера 7",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/702"
        },
        {
            "id": "802",
            "name": "📹 Камера 8",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/802"
        },
        {
            "id": "902",
            "name": "📹 Камера 9",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/902"
        },
        {
            "id": "1002",
            "name": "📹 Камера 10",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1002"
        },
        {
            "id": "1102",
            "name": "📹 Камера 11",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1102"
        },
        {
            "id": "1202",
            "name": "📹 Камера 12",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1202"
        },
        {
            "id": "1302",
            "name": "📹 Камера 13",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1302"
        },
        {
            "id": "1402",
            "name": "📹 Камера 14",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1402"
        },
        {
            "id": "1502",
            "name": "📹 Камера 15",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1502"
        },
        {
            "id": "1602",
            "name": "📹 Камера 16",
            "url": "rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1602"
        }
    ]
}
EOF
        print_success "Файл config.json создан"
        print_warning "Отредактируйте config.json для настройки ваших камер!"
    else
        print_success "Конфигурация найдена: config.json"
    fi
}

# Обновление конфигурации go2rtc
update_go2rtc_config() {
    print_info "Обновление конфигурации go2rtc для работы через интернет..."
    
    cat > go2rtc.yaml << 'EOF'
# Универсальная конфигурация go2rtc для работы через любые сети
streams:
  # Общий канал
  1: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1
  # Стандартные каналы
  102: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/102
  202: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/202
  302: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/302
  402: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/402
  502: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/502
  602: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/602
  702: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/702
  802: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/802
  902: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/902
  1002: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1002
  1102: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1102
  1202: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1202
  1302: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1302
  1402: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1402
  1502: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1502
  1602: rtsp://admin:oborotni2447@192.168.8.10:554/Streaming/Channels/1602

# API настройки - разрешаем доступ с любых доменов
api:
  listen: :1984
  origin: "*"
  cors: true

# WebRTC настройки для универсальной работы
webrtc:
  listen: :1984
  
  # Настройки для работы через любые сети
  tcp: true
  drop_late: true
  
  # Активируем все типы кандидатов
  candidates:
    - stun:stun.l.google.com:19302
    - stun:stun1.l.google.com:19302
    - stun:stun.stunprotocol.org:3478
    
  # ICE серверы для всех типов сетей  
  ice_servers:
    # STUN серверы
    - urls:
      - stun:stun.l.google.com:19302
      - stun:stun1.l.google.com:19302
      - stun:stun2.l.google.com:19302
      - stun:stun3.l.google.com:19302
      - stun:stun4.l.google.com:19302
      
    # TURN серверы через TCP и UDP
    - urls:
      - turn:openrelay.metered.ca:80?transport=tcp
      - turn:openrelay.metered.ca:443?transport=tcp
      - turn:openrelay.metered.ca:80
      - turn:openrelay.metered.ca:443
      username: openrelayproject
      credential: openrelayproject
EOF
    print_success "Конфигурация go2rtc обновлена для работы через интернет"
}

# Сборка приложения
build_app() {
    print_info "Проверка необходимости сборки..."
    
    # Проверяем, есть ли бинарник и актуален ли он
    if [ -f "teleoko" ] || [ -f "teleoko.exe" ]; then
        # Проверяем время модификации исходников
        NEWEST_SOURCE=$(find . -name "*.go" -newer "teleoko" 2>/dev/null | head -1)
        if [ -z "$NEWEST_SOURCE" ]; then
            print_success "Бинарник актуален, сборка не требуется"
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
        print_success "Приложение успешно собрано"
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
        CAMERA_IP=$(jq -r '.hikvision.ip' config.json 2>/dev/null || echo "192.168.8.10")
    else
        CAMERA_IP="192.168.8.10"  # По умолчанию
    fi
    
    if ping -c 1 -W 3 "$CAMERA_IP" &> /dev/null; then
        print_success "Камера $CAMERA_IP доступна"
    else
        print_warning "Камера $CAMERA_IP недоступна"
        print_info "Убедитесь, что:"
        print_info "  • Камера включена и подключена к сети"
        print_info "  • IP-адрес в config.json правильный"
        print_info "  • Нет блокировки файрволом"
    fi
}

# Запуск ngrok в фоновом режиме и сохранение URL в файл
start_ngrok() {
    if [ "$USE_NGROK" = true ]; then
        print_info "Запуск ngrok для внешнего доступа..."
        
        # Останавливаем предыдущие сессии ngrok, если есть
        pkill -f ngrok || true
        
        # Запускаем ngrok для веб-сервера
        ngrok http 8082 > /dev/null &
        NGROK_WEB_PID=$!
        print_success "ngrok запущен для порта 8082 (веб-интерфейс)"
        
        # Запускаем ngrok для go2rtc
        ngrok http 1984 > /dev/null &
        NGROK_GO2RTC_PID=$!
        print_success "ngrok запущен для порта 1984 (go2rtc WebRTC)"
        
        # Ждем немного, чтобы ngrok успел запуститься и получить URL
        sleep 5
        
        # Получаем URL для веб-интерфейса
        WEB_URL=$(curl -s http://localhost:4040/api/tunnels | grep -o '"public_url":"[^"]*"' | grep -o 'http[^"]*' | head -1)
        if [ -n "$WEB_URL" ]; then
            print_info "Внешний URL для веб-интерфейса: $WEB_URL"
            # Сохраняем URL в файл для передачи в приложение
            echo "$WEB_URL" > web_url.txt
        else
            print_warning "Не удалось получить URL для веб-интерфейса"
        fi
        
        # Получаем URL для go2rtc
        sleep 2  # Ждем, чтобы ngrok API обновился
        GO2RTC_URL=$(curl -s http://localhost:4041/api/tunnels | grep -o '"public_url":"[^"]*"' | grep -o 'http[^"]*' | head -1)
        if [ -n "$GO2RTC_URL" ]; then
            print_info "Внешний URL для go2rtc: $GO2RTC_URL"
            # Сохраняем URL в переменную окружения для передачи в приложение
            export GO2RTC_URL="$GO2RTC_URL"
            # Также сохраняем в файл
            echo "$GO2RTC_URL" > go2rtc_url.txt
        else
            print_warning "Не удалось получить URL для go2rtc"
        fi
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
    
    if [ "$USE_NGROK" = true ]; then
        echo "   • Через интернет: $WEB_URL"
    fi
    
    echo ""
    echo "🔧 Конфигурация:"
    echo "   • Файл:      config.json"
    echo "   • go2rtc:    http://localhost:1984"
    
    if [ "$USE_NGROK" = true ]; then
        echo "   • go2rtc URL: $GO2RTC_URL"
    fi
    
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
    check_ngrok
    check_ports
    
    echo ""
    print_info "Подготовка конфигурации..."
    create_default_config
    update_go2rtc_config
    
    echo ""
    print_info "Сборка приложения..."
    build_app
    
    echo ""
    test_camera_connection
    
    echo ""
    print_info "Настройка внешнего доступа..."
    start_ngrok
    
    echo ""
    show_startup_info
    
    # Запуск приложения с передачей URL go2rtc через переменную окружения
    # Экспортируем GO2RTC_URL перед запуском, если он доступен
    if [ -n "$GO2RTC_URL" ]; then
        GO2RTC_URL="$GO2RTC_URL" ./teleoko 2>&1 | tee teleoko.log
    else
        ./teleoko 2>&1 | tee teleoko.log
    fi
}

# Обработка сигналов завершения
trap 'echo -e "\n👋 Завершение TeleOko..."; if [ "$USE_NGROK" = true ]; then kill $NGROK_WEB_PID $NGROK_GO2RTC_PID 2>/dev/null || true; fi; exit 0' INT TERM

# Запуск
main "$@"