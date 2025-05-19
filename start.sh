#!/bin/bash

echo "Запуск системы видеонаблюдения TeleOko..."

# Проверяем наличие бинарного файла
if [ ! -f ./teleoko ]; then
    echo "Сборка приложения..."
    go build -o teleoko ./cmd/server
    if [ $? -ne 0 ]; then
        echo "Ошибка при сборке приложения!"
        exit 1
    fi
fi

# Выдаем права на исполнение
chmod +x ./teleoko

# Запускаем сервер
echo "Запуск сервера на порту 8082..."
echo "Откройте браузер и перейдите по адресу http://localhost:8082"

./teleoko