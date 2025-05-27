@echo off
chcp 65001 >nul
title Ручная загрузка go2rtc

echo 📥 Ручная загрузка go2rtc для Windows
echo ====================================

:: Проверяем, есть ли уже go2rtc
if exist "go2rtc.exe" (
    echo ✅ go2rtc.exe уже существует
    echo.
    choice /M "Заменить существующий файл"
    if errorlevel 2 goto :end
    del go2rtc.exe
)

echo.
echo 🌐 Пробуем несколько источников загрузки...
echo.

:: Попробуем несколько версий и источников
set "versions=1.9.9 1.9.8 1.9.7 1.9.6 1.9.5"
set "found=0"

for %%v in (%versions%) do (
    if !found! equ 0 (
        echo ℹ️  Попытка загрузки версии %%v...
        
        :: Пробуем разные форматы имен файлов
        powershell -Command "try { Invoke-WebRequest -Uri 'https://github.com/AlexxIT/go2rtc/releases/download/v%%v/go2rtc_windows_amd64.zip' -OutFile 'go2rtc.zip' -ErrorAction Stop; Write-Host 'Успешно загружен ZIP'; exit 0 } catch { Write-Host 'ZIP не найден' }"
        if exist "go2rtc.zip" (
            echo ✅ Загружен go2rtc v%%v (ZIP)
            set "found=1"
            goto :extract
        )
        
        powershell -Command "try { Invoke-WebRequest -Uri 'https://github.com/AlexxIT/go2rtc/releases/download/v%%v/go2rtc_win64.zip' -OutFile 'go2rtc.zip' -ErrorAction Stop; Write-Host 'Успешно загружен ZIP'; exit 0 } catch { Write-Host 'Win64 ZIP не найден' }"
        if exist "go2rtc.zip" (
            echo ✅ Загружен go2rtc v%%v (Win64 ZIP)
            set "found=1"
            goto :extract
        )
        
        powershell -Command "try { Invoke-WebRequest -Uri 'https://github.com/AlexxIT/go2rtc/releases/download/v%%v/go2rtc.exe' -OutFile 'go2rtc.exe' -ErrorAction Stop; Write-Host 'Успешно загружен EXE'; exit 0 } catch { Write-Host 'EXE не найден' }"
        if exist "go2rtc.exe" (
            echo ✅ Загружен go2rtc v%%v (EXE)
            set "found=1"
            goto :success
        )
    )
)

:extract
if exist "go2rtc.zip" (
    echo ℹ️  Извлечение из архива...
    powershell -Command "Expand-Archive -Path 'go2rtc.zip' -DestinationPath '.' -Force"
    
    :: Ищем исполняемый файл
    if exist "go2rtc.exe" (
        echo ✅ go2rtc.exe извлечен
        del go2rtc.zip
        goto :success
    )
    
    :: Ищем в подпапках
    for /r %%f in (go2rtc.exe) do (
        if exist "%%f" (
            copy "%%f" "go2rtc.exe" >nul
            echo ✅ go2rtc.exe скопирован из архива
            del go2rtc.zip
            rmdir /s /q go2rtc 2>nul
            goto :success
        )
    )
    
    echo ❌ go2rtc.exe не найден в архиве
    del go2rtc.zip
)

:: Если автоматическая загрузка не удалась
echo.
echo ❌ Автоматическая загрузка не удалась
echo.
echo 📋 Ручная загрузка:
echo 1. Откройте: https://github.com/AlexxIT/go2rtc/releases/latest
echo 2. Найдите файл для Windows (go2rtc_windows_amd64.zip или go2rtc.exe)
echo 3. Скачайте и поместите go2rtc.exe в папку с TeleOko
echo.
echo 🔍 Альтернативные варианты:
echo • Используйте Docker: docker pull alexxit/go2rtc
echo • Отключите go2rtc в config.json: "enabled": false
echo.
pause
goto :end

:success
echo.
echo 🎉 go2rtc успешно загружен!
echo.
echo ℹ️  Проверка версии:
go2rtc.exe --version 2>nul || echo go2rtc.exe готов к использованию
echo.
echo ✅ Теперь можете запустить TeleOko:
echo    .\teleoko.exe
echo.

:end
pause