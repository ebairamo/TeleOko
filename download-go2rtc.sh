#!/bin/bash

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

# Заголовок
echo "📥 Ручная загрузка go2rtc для macOS"
echo "===================================="

# Проверяем, есть ли уже go2rtc
if [ -f "go2rtc" ]; then
    print_success "go2rtc уже существует"
    echo
    read -p "Заменить существующий файл? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Выход..."
        exit 0
    fi
    rm -f go2rtc
fi

echo
print_info "🌐 Пробуем несколько источников загрузки..."
echo

# Определяем архитектуру
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH_NAME="amd64"
elif [ "$ARCH" = "arm64" ]; then
    ARCH_NAME="arm64"
else
    print_error "Неподдерживаемая архитектура: $ARCH"
    echo
    echo "📋 Поддерживаемые архитектуры:"
    echo "• Intel Mac (x86_64) - amd64"
    echo "• Apple Silicon (arm64) - arm64"
    exit 1
fi

print_info "Определена архитектура: $ARCH -> $ARCH_NAME"

# Попробуем несколько версий и источников
versions=("1.9.9" "1.9.8" "1.9.7" "1.9.6" "1.9.5")
found=0

for version in "${versions[@]}"; do
    if [ $found -eq 0 ]; then
        print_info "Попытка загрузки версии $version..."
        
        # Список возможных имен файлов для macOS
        filenames=(
            "go2rtc_darwin_${ARCH_NAME}.zip"
            "go2rtc_darwin_${ARCH_NAME}"
            "go2rtc_mac_${ARCH_NAME}.zip"
            "go2rtc_mac_${ARCH_NAME}"
            "go2rtc-darwin-${ARCH_NAME}.zip"
            "go2rtc-darwin-${ARCH_NAME}"
        )
        
        # Пробуем каждое имя файла
        for filename in "${filenames[@]}"; do
            if [ $found -eq 0 ]; then
                url="https://github.com/AlexxIT/go2rtc/releases/download/v${version}/${filename}"
                print_info "Пробуем: $filename"
                
                if [[ "$filename" == *.zip ]]; then
                    # ZIP архив
                    if curl -L -f -o "go2rtc.zip" "$url" 2>/dev/null; then
                        print_success "Загружен go2rtc v$version ($filename)"
                        found=1
                        break
                    fi
                else
                    # Прямой бинарник
                    if curl -L -f -o "go2rtc" "$url" 2>/dev/null; then
                        print_success "Загружен go2rtc v$version ($filename)"
                        chmod +x go2rtc
                        found=1
                        break
                    fi
                fi
            fi
        done
        
        if [ $found -eq 0 ]; then
            print_warning "Версия $version недоступна для macOS $ARCH_NAME"
        fi
    fi
done

# Извлекаем из архива, если был загружен ZIP
if [ -f "go2rtc.zip" ] && [ $found -eq 1 ]; then
    print_info "📦 Извлечение из архива..."
    
    if command -v unzip >/dev/null 2>&1; then
        # Используем unzip
        unzip -q go2rtc.zip
        
        # Ищем исполняемый файл
        if [ -f "go2rtc" ]; then
            print_success "go2rtc извлечен"
            chmod +x go2rtc
            rm go2rtc.zip
        else
            # Ищем в подпапках
            found_binary=$(find . -name "go2rtc" -type f 2>/dev/null | head -1)
            if [ -n "$found_binary" ]; then
                cp "$found_binary" "go2rtc"
                chmod +x go2rtc
                print_success "go2rtc скопирован из архива"
                rm go2rtc.zip
                # Удаляем временные папки
                find . -type d -name "*go2rtc*" -exec rm -rf {} + 2>/dev/null || true
            else
                print_error "go2rtc не найден в архиве"
                rm go2rtc.zip
                found=0
            fi
        fi
    else
        print_error "unzip не найден! Установите: brew install unzip"
        rm go2rtc.zip
        found=0
    fi
fi

# Проверяем результат
if [ $found -eq 0 ]; then
    echo
    print_warning "Файлы go2rtc для macOS не найдены в официальных релизах"
    echo
    
    # Пробуем Homebrew
    if command -v brew >/dev/null 2>&1; then
        print_info "🍺 Homebrew найден! Пробуем установить go2rtc..."
        echo
        
        if brew install go2rtc 2>/dev/null; then
            print_success "go2rtc успешно установлен через Homebrew!"
            
            # Создаем символическую ссылку в текущей директории
            if command -v go2rtc >/dev/null 2>&1; then
                go2rtc_path=$(which go2rtc)
                ln -sf "$go2rtc_path" ./go2rtc
                print_success "Создана ссылка на go2rtc в текущей директории"
                found=1
            fi
        else
            print_warning "Не удалось установить go2rtc через Homebrew"
        fi
    fi
    
    if [ $found -eq 0 ]; then
        print_error "Автоматическая установка не удалась"
        echo
        print_info "🔍 Альтернативные способы установки go2rtc на macOS:"
        echo
        echo "1️⃣ **Homebrew (РЕКОМЕНДУЕТСЯ):**"
        if command -v brew >/dev/null 2>&1; then
            echo "   brew install go2rtc"
        else
            echo "   • Установите Homebrew: https://brew.sh"
            echo "   • Затем выполните: brew install go2rtc"
        fi
        echo
        echo "2️⃣ **Компиляция из исходников:**"
        echo "   git clone https://github.com/AlexxIT/go2rtc.git"
        echo "   cd go2rtc"
        echo "   go build -o go2rtc"
        echo "   cp go2rtc /путь/к/TeleOko/"
        echo
        echo "3️⃣ **Docker:**"
        echo "   docker pull alexxit/go2rtc"
        echo "   # Запуск в контейнере"
        echo
        echo "4️⃣ **Отключить go2rtc (простейший способ):**"
        echo "   • В config.json установите: \"enabled\": false"
        echo "   • TeleOko будет работать с RTSP напрямую"
        echo
        echo "💡 **Архитектура вашего Mac:** $ARCH_NAME"
        echo
        echo "📋 **Для быстрого старта без go2rtc:**"
        echo "   • Откройте config.json"
        echo "   • Найдите секцию \"go2rtc\""
        echo "   • Установите \"enabled\": false"
        echo "   • TeleOko будет работать с прямыми RTSP-ссылками"
        echo
        exit 1
    fi
fi

# Успешное завершение
echo
print_success "🎉 go2rtc успешно загружен!"
echo

# Проверка версии
print_info "ℹ️  Проверка версии:"
if ./go2rtc --version 2>/dev/null; then
    print_success "go2rtc работает корректно"
else
    print_warning "Не удалось получить версию, но файл готов к использованию"
fi

echo
print_success "✅ Теперь можете запустить TeleOko:"
echo "   ./start.sh"
echo
print_info "🔧 Дополнительная информация:"
echo "• Файл go2rtc помещен в текущую директорию"
echo "• Права на выполнение установлены автоматически"
echo "• Архитектура: $ARCH_NAME"
echo "• Для обновления повторно запустите этот скрипт"
echo