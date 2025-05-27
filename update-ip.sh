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

echo "🔧 Простая утилита обновления IP камеры в TeleOko"
echo "================================================"

# Проверяем наличие config.json
if [ ! -f "config.json" ]; then
    print_error "Файл config.json не найден!"
    echo
    echo "Убедитесь, что запускаете скрипт в папке с TeleOko"
    exit 1
fi

# Получаем текущий IP (простой парсинг)
current_ip=$(grep -o '"ip": *"[^"]*"' config.json | head -1 | sed 's/"ip": *"//' | sed 's/"//')

echo
print_info "Текущий IP камеры: $current_ip"
echo

# Запрашиваем новый IP
read -p "Введите новый IP камеры (например, 192.168.8.6): " new_ip

# Проверяем формат IP
if [[ ! $new_ip =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
    print_error "Неверный формат IP адреса!"
    echo "Используйте формат: xxx.xxx.xxx.xxx"
    exit 1
fi

# Создаем резервную копию
cp config.json config.json.backup
print_success "Создана резервная копия: config.json.backup"

echo
print_info "Обновляем IP в config.json..."

# Заменяем IP в config.json
sed -i.tmp "s/\"ip\": \"$current_ip\"/\"ip\": \"$new_ip\"/g" config.json
sed -i.tmp "s/$current_ip/$new_ip/g" config.json
rm -f config.json.tmp

print_success "config.json обновлен"

# Обновляем go2rtc.yaml если он существует
if [ -f "go2rtc.yaml" ]; then
    print_info "Обновляем go2rtc.yaml..."
    cp go2rtc.yaml go2rtc.yaml.backup
    sed -i.tmp "s/$current_ip/$new_ip/g" go2rtc.yaml
    rm -f go2rtc.yaml.tmp
    print_success "go2rtc.yaml обновлен"
fi

echo
print_success "🎉 Конфигурация успешно обновлена!"
echo
print_info "Изменения:"
print_info "  Старый IP: $current_ip"
print_info "  Новый IP:  $new_ip"
echo
print_info "Обновленные файлы:"
print_info "  ✅ config.json"
if [ -f "go2rtc.yaml.backup" ]; then
    print_info "  ✅ go2rtc.yaml"
fi
echo
print_info "Резервные копии:"
print_info "  📁 config.json.backup"
if [ -f "go2rtc.yaml.backup" ]; then
    print_info "  📁 go2rtc.yaml.backup"
fi
echo

# Проверяем доступность новой камеры
print_info "🔍 Проверка доступности камеры $new_ip..."
if ping -c 1 -W 3 "$new_ip" &> /dev/null; then
    print_success "Камера $new_ip доступна по сети"
else
    print_warning "Камера $new_ip недоступна по сети"
    echo "Убедитесь, что:"
    echo "  • Камера включена"
    echo "  • IP адрес правильный"  
    echo "  • Камера находится в той же сети"
fi

echo
print_success "✅ Готово! Теперь можете запустить TeleOko:"
echo "   ./start.sh"
echo
print_info "💡 Если что-то пошло не так, восстановите из резервной копии:"
echo "   cp config.json.backup config.json"

# Показываем содержимое обновленного config.json
echo
print_info "📄 Проверьте обновленную конфигурацию:"
echo "   Hikvision IP: $(grep -o '"ip": *"[^"]*"' config.json | head -1 | sed 's/"ip": *"//' | sed 's/"//')"
echo "   Количество обновленных URL: $(grep -c "$new_ip" config.json)"