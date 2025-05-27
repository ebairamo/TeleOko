/**
 * TeleOko - Модуль для работы с камерами
 * Управление обнаружением и выбором камер
 */

// Функция для загрузки информации о камерах
function loadCamerasInfo() {
    const camerasInfo = document.getElementById('camerasInfo');
    
    if (!camerasInfo) return; // Проверка на существование элемента
    
    // Показываем индикатор загрузки
    camerasInfo.innerHTML = '<div class="cameras-empty">Загрузка информации о камерах...</div>';
    
    fetch('/api/cameras')
        .then(response => response.json())
        .then(data => {
            if (!data.cameras || data.cameras.length === 0) {
                camerasInfo.innerHTML = '<div class="cameras-empty">Камеры не обнаружены</div>';
                return;
            }
            
            camerasInfo.innerHTML = '';
            
            // Получаем предпочтительную камеру из localStorage, если есть
            const preferredCameraIP = localStorage.getItem('preferredCameraIP');
            
            // Отображаем информацию о каждой камере
            data.cameras.forEach(camera => {
                const cameraItem = document.createElement('div');
                cameraItem.className = 'camera-item';
                
                // Определяем, выбрана ли эта камера как предпочтительная
                const isPreferred = camera.IP === preferredCameraIP || camera.IP === data.current_camera;
                
                const lastSeen = new Date(camera.LastSeen).toLocaleString('ru-RU');
                
                cameraItem.innerHTML = `
                    <div class="camera-info">
                        <span class="camera-ip">${camera.IP} ${isPreferred ? '<span class="camera-preferred">(Активная)</span>' : ''}</span>
                        <span class="camera-status ${camera.Status}">${camera.Status === 'online' ? 'В сети' : 'Не в сети'}</span>
                        <span class="camera-timestamp">Последняя активность: ${lastSeen}</span>
                    </div>
                    <div class="camera-actions">
                        <button class="camera-action-btn camera-primary" onclick="selectCamera('${camera.IP}')">Выбрать</button>
                    </div>
                `;
                
                camerasInfo.appendChild(cameraItem);
            });
        })
        .catch(error => {
            console.error('Ошибка при получении информации о камерах:', error);
            camerasInfo.innerHTML = '<div class="cameras-empty">Ошибка загрузки информации о камерах</div>';
        });
}

// Функция для запуска ручного сканирования камер
function triggerCameraScan() {
    const scanBtn = document.getElementById('scanCamerasBtn');
    if (scanBtn) {
        scanBtn.disabled = true;
        scanBtn.textContent = 'Сканирование...';
    }
    
    fetch('/api/scan_cameras', {
        method: 'POST',
    })
        .then(response => response.json())
        .then(data => {
            if (scanBtn) {
                scanBtn.disabled = false;
                scanBtn.textContent = 'Сканировать сеть';
            }
            
            alert('Запущено сканирование камер. Обновление списка через 5 секунд.');
            
            // Обновляем список через 5 секунд
            setTimeout(loadCamerasInfo, 5000);
        })
        .catch(error => {
            console.error('Ошибка при запуске сканирования:', error);
            alert('Ошибка при запуске сканирования камер');
            
            if (scanBtn) {
                scanBtn.disabled = false;
                scanBtn.textContent = 'Сканировать сеть';
            }
        });
}

// Функция для выбора камеры
function selectCamera(ip) {
    // Сохраняем выбранную камеру в localStorage
    localStorage.setItem('preferredCameraIP', ip);
    
    // Также отправляем на сервер, чтобы сохранить в конфигурации
    const formData = new FormData();
    formData.append('camera_ip', ip);
    
    fetch('/api/set_preferred_camera', {
        method: 'POST',
        body: formData
    })
        .then(response => response.json())
        .then(data => {
            if (data.status === 'success') {
                alert(`Камера с IP ${ip} выбрана в качестве основной`);
                // Обновляем список камер, чтобы отобразить новую активную камеру
                loadCamerasInfo();
            } else {
                alert(`Ошибка при выборе камеры: ${data.error || 'Неизвестная ошибка'}`);
            }
        })
        .catch(error => {
            console.error('Ошибка при выборе камеры:', error);
            alert(`Ошибка при выборе камеры: ${error.message}`);
        });
}

// Инициализация при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    // Загружаем информацию о камерах
    loadCamerasInfo();
    
    // Подключаем обработчик кнопки сканирования
    const scanCamerasBtn = document.getElementById('scanCamerasBtn');
    if (scanCamerasBtn) {
        scanCamerasBtn.addEventListener('click', triggerCameraScan);
    }
    
    // Периодическое обновление информации о камерах (каждые 30 секунд)
    setInterval(loadCamerasInfo, 30000);
});