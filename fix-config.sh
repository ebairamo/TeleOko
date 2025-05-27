#!/bin/bash

echo "🔧 Исправление IP камеры в конфигурации"
echo "======================================"

# Останавливаем TeleOko если запущен
echo "⏹️ Остановка TeleOko..."
pkill -f teleoko 2>/dev/null || true

# Ждем немного
sleep 2

# Исправляем config.json
if [ -f "config.json" ]; then
    echo "📝 Обновление config.json..."
    cp config.json config.json.backup
    
    # Заменяем IP в hikvision секции
    sed -i.tmp 's/"ip": "192\.168\.8\.[0-9]*"/"ip": "192.168.8.10"/g' config.json
    
    # Заменяем IP во всех URL каналов
    sed -i.tmp 's/192\.168\.8\.[0-9]*/192.168.8.10/g' config.json
    
    rm -f config.json.tmp
    echo "✅ config.json обновлен"
else
    echo "❌ config.json не найден"
fi

# Исправляем go2rtc.yaml если есть
if [ -f "go2rtc.yaml" ]; then
    echo "📝 Обновление go2rtc.yaml..."
    cp go2rtc.yaml go2rtc.yaml.backup
    sed -i.tmp 's/192\.168\.8\.[0-9]*/192.168.8.10/g' go2rtc.yaml
    rm -f go2rtc.yaml.tmp
    echo "✅ go2rtc.yaml обновлен"
fi

echo ""
echo "🎉 Конфигурация исправлена!"
echo "IP камеры теперь: 192.168.8.10"
echo ""
echo "🚀 Запускайте TeleOko:"
echo "./teleoko"