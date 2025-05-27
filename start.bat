@echo off
echo Запуск системы видеонаблюдения TeleOko...

:: Проверяем наличие бинарного файла
if not exist teleoko.exe (
    echo Сборка приложения...
    go build -o teleoko.exe .\cmd\server
    if %ERRORLEVEL% neq 0 (
        echo Ошибка при сборке приложения!
        pause
        exit /b 1
    )
)

:: Запускаем сервер
echo Запуск сервера на порту 8082...
start http://localhost:8082
teleoko.exe

pause