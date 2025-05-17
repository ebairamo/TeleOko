/**
 * TeleOko - Скрипт для управления системой видеонаблюдения
 * Version: 1.0.0
 */

document.addEventListener('DOMContentLoaded', function() {
    // Основные элементы интерфейса
    const videoContainer = document.getElementById('video-container');
    const cameraSelect = document.getElementById('cameraSelect');
    const liveBtn = document.getElementById('liveBtn');
    const snapshotBtn = document.getElementById('snapshotBtn');
    const archiveDate = document.getElementById('archiveDate');
    const searchBtn = document.getElementById('searchBtn');
    const timeline = document.getElementById('timeline');
    const recordingsList = document.getElementById('recordingsList');
    const snapshotModal = document.getElementById('snapshotModal');
    const snapshotImage = document.getElementById('snapshotImage');
    const downloadBtn = document.getElementById('downloadBtn');
    const closeSnapshotBtn = document.getElementById('closeSnapshotBtn');
    const closeBtn = document.querySelector('.close-btn');
    const loadingOverlay = document.getElementById('loadingOverlay');
    const loadingMessage = document.getElementById('loadingMessage');
    
    // Текущее состояние
    let currentVideoElement = null;
    let currentRTCPeerConnection = null;
    let currentMediaStream = null;
    let connectionStatus = 'online';
    let connectionStatusElement = document.querySelector('.connection-status');
    
    // Установка текущей даты по умолчанию
    archiveDate.valueAsDate = new Date();
    
    /**
     * Отображение индикатора загрузки
     */
    function showLoading(message = 'Загрузка...') {
        loadingMessage.textContent = message;
        loadingOverlay.style.display = 'flex';
    }
    
    /**
     * Скрытие индикатора загрузки
     */
    function hideLoading() {
        loadingOverlay.style.display = 'none';
    }
    
    /**
     * Обновление статуса подключения
     */
    function updateConnectionStatus(status) {
        connectionStatus = status;
        connectionStatusElement.className = 'connection-status ' + status;
        connectionStatusElement.textContent = status === 'online' ? 'Подключено' : 'Не подключено';
    }
    
    /**
     * Отображение ошибки в контейнере
     */
    function showError(container, message) {
        container.innerHTML = `<div class="error">
            <p>${message}</p>
        </div>`;
    }
    
    /**
     * Форматирование даты и времени
     */
    function formatDateTime(dateTimeString) {
        const options = { 
            day: '2-digit', 
            month: '2-digit', 
            year: 'numeric', 
            hour: '2-digit', 
            minute: '2-digit' 
        };
        return new Date(dateTimeString).toLocaleString('ru-RU', options);
    }
    
    /**
     * Рассчет продолжительности записи
     */
    function calculateDuration(startTime, endTime) {
        const start = new Date(startTime);
        const end = new Date(endTime);
        const diffMs = end - start;
        
        const minutes = Math.floor(diffMs / (1000 * 60));
        const seconds = Math.floor((diffMs % (1000 * 60)) / 1000);
        
        return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    
    /**
     * Запуск прямого эфира
     */
    async function startLiveStream() {
        const channel = cameraSelect.value;
        
        // Отображение загрузки
        showLoading('Подключение к камере...');
        videoContainer.innerHTML = '<div class="loading">Подключение к камере...</div>';
        
        // Остановка предыдущего видеопотока, если есть
        stopCurrentStream();
        
        try {
            // Создание видеоэлемента
            const videoElement = document.createElement('video');
            videoElement.autoplay = true;
            videoElement.controls = true;
            videoElement.playsInline = true;
            videoElement.style.width = '100%';
            videoElement.style.height = '100%';
            
            // Настройка WebRTC-соединения
            const pc = new RTCPeerConnection({
                iceServers: [
                    { urls: 'stun:stun.l.google.com:19302' }
                ]
            });
            
            // Сохраняем соединение для возможности закрытия позже
            currentRTCPeerConnection = pc;
            
            // Обработчик получения трека
            pc.ontrack = function(event) {
                videoElement.srcObject = event.streams[0];
                currentMediaStream = event.streams[0];
            };
            
            // Обработчик состояния подключения
            pc.oniceconnectionstatechange = function() {
                console.log('ICE connection state:', pc.iceConnectionState);
                if (pc.iceConnectionState === 'disconnected' || 
                    pc.iceConnectionState === 'failed' || 
                    pc.iceConnectionState === 'closed') {
                    updateConnectionStatus('offline');
                } else if (pc.iceConnectionState === 'connected' || 
                          pc.iceConnectionState === 'completed') {
                    updateConnectionStatus('online');
                }
            };
            
            // Создание оффера
            const offer = await pc.createOffer({
                offerToReceiveVideo: true,
                offerToReceiveAudio: true
            });
            
            await pc.setLocalDescription(offer);
            
            // Отправка оффера на сервер
            const response = await fetch(`/api/webrtc/offer?channel=${channel}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(offer)
            });
            
            if (!response.ok) {
                throw new Error(`Ошибка HTTP: ${response.status}`);
            }
            
            const answer = await response.json();
            
            if (answer.error) {
                throw new Error(answer.error);
            }
            
            // Установка удаленного описания
            await pc.setRemoteDescription(new RTCSessionDescription(answer));
            
            // Очистка контейнера и добавление видео
            videoContainer.innerHTML = '';
            videoContainer.appendChild(videoElement);
            
            // Сохраняем видеоэлемент
            currentVideoElement = videoElement;
            
            // Обновляем статус подключения
            updateConnectionStatus('online');
            
        } catch (error) {
            console.error('Ошибка при запуске прямого эфира:', error);
            showError(videoContainer, `Не удалось подключиться к камере: ${error.message}`);
            updateConnectionStatus('offline');
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Остановка текущего видеопотока
     */
    function stopCurrentStream() {
        // Закрыть WebRTC-соединение
        if (currentRTCPeerConnection) {
            currentRTCPeerConnection.close();
            currentRTCPeerConnection = null;
        }
        
        // Остановить медиапотоки
        if (currentMediaStream) {
            currentMediaStream.getTracks().forEach(track => track.stop());
            currentMediaStream = null;
        }
        
        // Очистить видеоэлемент
        if (currentVideoElement) {
            currentVideoElement.srcObject = null;
            currentVideoElement = null;
        }
    }
    
    /**
     * Сделать снимок с камеры
     */
    function takeSnapshot() {
        if (!currentVideoElement || !currentVideoElement.srcObject) {
            alert('Нет активного видеопотока. Сначала запустите прямой эфир.');
            return;
        }
        
        try {
            // Создать канвас и скопировать на него текущий кадр видео
            const canvas = document.createElement('canvas');
            canvas.width = currentVideoElement.videoWidth;
            canvas.height = currentVideoElement.videoHeight;
            
            const ctx = canvas.getContext('2d');
            ctx.drawImage(currentVideoElement, 0, 0, canvas.width, canvas.height);
            
            // Получить изображение как URL
            const imageDataUrl = canvas.toDataURL('image/png');
            
            // Отобразить в модальном окне
            snapshotImage.src = imageDataUrl;
            snapshotModal.style.display = 'flex';
            
            // Настроить кнопку скачивания
            downloadBtn.onclick = function() {
                const link = document.createElement('a');
                link.href = imageDataUrl;
                link.download = `snapshot_${new Date().toISOString().replace(/[:.]/g, '-')}.png`;
                document.body.appendChild(link);
                link.click();
                document.body.removeChild(link);
            };
            
        } catch (error) {
            console.error('Ошибка при создании снимка:', error);
            alert(`Не удалось сделать снимок: ${error.message}`);
        }
    }
    
    /**
     * Поиск записей
     */
    async function searchRecordings() {
        const channel = cameraSelect.value;
        const date = archiveDate.value;
        
        // Показать индикатор загрузки
        showLoading('Поиск записей...');
        recordingsList.innerHTML = '<div class="loading">Поиск записей...</div>';
        timeline.innerHTML = '<div class="loading">Загрузка временной шкалы...</div>';
        
        try {
            // Запрос к API
            const response = await fetch(`/api/recordings?channel=${channel}&start=${date}`, {
                headers: {
                    'Cache-Control': 'no-cache'
                }
            });
            
            if (!response.ok) {
                throw new Error(`Ошибка HTTP: ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.error) {
                throw new Error(data.error);
            }
            
            // Отображение результатов
            displayTimeline(data.recordings);
            displayRecordingsList(data.recordings);
            
        } catch (error) {
            console.error('Ошибка при поиске записей:', error);
            showError(recordingsList, `Не удалось найти записи: ${error.message}`);
            showError(timeline, `Не удалось загрузить временную шкалу: ${error.message}`);
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Отображение временной шкалы
     */
    function displayTimeline(recordings) {
        timeline.innerHTML = '';
        
        if (!recordings || recordings.length === 0) {
            timeline.innerHTML = '<div class="timeline-empty">Записи не найдены</div>';
            return;
        }
        
        // Создание контейнера для шкалы
        const timelineContainer = document.createElement('div');
        timelineContainer.className = 'timeline-container';
        
        // Добавление меток времени
        for (let hour = 0; hour < 24; hour++) {
            const timeLabel = document.createElement('div');
            timeLabel.className = 'time-label';
            timeLabel.style.left = `${(hour / 24) * 100}%`;
            timeLabel.textContent = `${hour}:00`;
            timelineContainer.appendChild(timeLabel);
        }
        
        // Отображение записей на шкале
        recordings.forEach(recording => {
            const startTime = new Date(recording.StartTime);
            const endTime = new Date(recording.EndTime);
            
            // Расчет положения и ширины сегмента
            const startOfDay = new Date(startTime);
            startOfDay.setHours(0, 0, 0, 0);
            
            const dayDuration = 24 * 60 * 60 * 1000; // 24 часа в мс
            
            const startPosition = ((startTime - startOfDay) / dayDuration) * 100;
            const width = ((endTime - startTime) / dayDuration) * 100;
            
            // Создание сегмента
            const segment = document.createElement('div');
            segment.className = 'timeline-segment';
            segment.style.left = `${startPosition}%`;
            segment.style.width = `${width}%`;
            
            // Добавление всплывающей подсказки
            segment.title = `${formatDateTime(recording.StartTime)} - ${formatDateTime(recording.EndTime)}`;
            
            // Обработчик клика для воспроизведения
            segment.onclick = () => playRecording(recording);
            
            timelineContainer.appendChild(segment);
        });
        
        timeline.appendChild(timelineContainer);
    }
   function displayRecordingsList(recordings) {
    recordingsList.innerHTML = '';
    
    if (!recordings || recordings.length === 0) {
        recordingsList.innerHTML = '<div class="recordings-empty">Записи не найдены</div>';
        return;
    }
    
    // Сортировка записей по времени (сначала новые)
    recordings.sort((a, b) => new Date(b.StartTime) - new Date(a.StartTime));
    
    // Отображение каждой записи
    recordings.forEach(recording => {
        const startTime = new Date(recording.StartTime);
        const endTime = new Date(recording.EndTime);
        
        const formattedStart = formatDateTime(recording.StartTime);
        const formattedEnd = formatDateTime(recording.EndTime);
        const duration = calculateDuration(recording.StartTime, recording.EndTime);
        
        const recordingItem = document.createElement('div');
        recordingItem.className = 'recording-item';
        
        recordingItem.innerHTML = `
            <div class="recording-info">
                <span class="recording-time">${formattedStart}</span>
                <span class="recording-duration">Длительность: ${duration}</span>
            </div>
            <button class="play-btn primary-btn">Воспроизвести</button>
        `;
        
        // Обработчик клика на кнопку воспроизведения
        recordingItem.querySelector('.play-btn').onclick = () => playRecording(recording);
        
        recordingsList.appendChild(recordingItem);
    });
}

/**
 * Воспроизведение записи из архива
 */
async function playRecording(recording) {
    // Отображение загрузки
    showLoading('Загрузка архивной записи...');
    videoContainer.innerHTML = '<div class="loading">Загрузка архивной записи...</div>';
    
    // Остановка предыдущего видеопотока, если есть
    stopCurrentStream();
    
    try {
        // Получение URL для воспроизведения
        const response = await fetch(`/api/playback-url?channel=${recording.Channel}&start=${recording.StartTime}&end=${recording.EndTime}`);
        
        if (!response.ok) {
            throw new Error(`Ошибка HTTP: ${response.status}`);
        }
        
        const data = await response.json();
        
        if (data.error) {
            throw new Error(data.error);
        }
        
        // Создание видеоэлемента
        const videoElement = document.createElement('video');
        videoElement.autoplay = true;
        videoElement.controls = true;
        videoElement.playsInline = true;
        videoElement.style.width = '100%';
        videoElement.style.height = '100%';
        
        // Настройка WebRTC-соединения для архива
        const pc = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' }
            ]
        });
        
        // Сохраняем соединение для возможности закрытия позже
        currentRTCPeerConnection = pc;
        
        // Обработчик получения трека
        pc.ontrack = function(event) {
            videoElement.srcObject = event.streams[0];
            currentMediaStream = event.streams[0];
        };
        
        // Обработчик состояния подключения
        pc.oniceconnectionstatechange = function() {
            console.log('ICE connection state:', pc.iceConnectionState);
            if (pc.iceConnectionState === 'disconnected' || 
                pc.iceConnectionState === 'failed' || 
                pc.iceConnectionState === 'closed') {
                updateConnectionStatus('offline');
            } else if (pc.iceConnectionState === 'connected' || 
                      pc.iceConnectionState === 'completed') {
                updateConnectionStatus('online');
            }
        };
        
        // Создание оффера
        const offer = await pc.createOffer({
            offerToReceiveVideo: true,
            offerToReceiveAudio: true
        });
        
        await pc.setLocalDescription(offer);
        
        // Отправка оффера на сервер
        const webrtcResponse = await fetch(`/api/webrtc/offer/playback`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                offer: offer,
                url: data.url
            })
        });
        
        if (!webrtcResponse.ok) {
            throw new Error(`Ошибка HTTP: ${webrtcResponse.status}`);
        }
        
        const answer = await webrtcResponse.json();
        
        if (answer.error) {
            throw new Error(answer.error);
        }
        
        // Установка удаленного описания
        await pc.setRemoteDescription(new RTCSessionDescription(answer));
        
        // Очистка контейнера и добавление видео
        videoContainer.innerHTML = '';
        videoContainer.appendChild(videoElement);
        
        // Добавление информации о воспроизводимой записи
        const infoOverlay = document.createElement('div');
        infoOverlay.className = 'playback-info';
        infoOverlay.innerHTML = `
            <div class="playback-details">
                <span>Архив: ${formatDateTime(recording.StartTime)}</span>
            </div>
        `;
        videoContainer.appendChild(infoOverlay);
        
        // Сохраняем видеоэлемент
        currentVideoElement = videoElement;
        
        // Обновляем статус подключения
        updateConnectionStatus('online');
        
    } catch (error) {
        console.error('Ошибка при воспроизведении архива:', error);
        showError(videoContainer, `Не удалось воспроизвести запись: ${error.message}`);
        updateConnectionStatus('offline');
    } finally {
        hideLoading();
    }
}

/**
 * Периодическая проверка статуса сервера
 */
async function checkServerStatus() {
    try {
        const response = await fetch('/api/info', {
            headers: {
                'Cache-Control': 'no-cache'
            }
        });
        
        if (response.ok) {
            updateConnectionStatus('online');
        } else {
            updateConnectionStatus('offline');
        }
    } catch (error) {
        console.error('Ошибка при проверке статуса сервера:', error);
        updateConnectionStatus('offline');
    }
}

/**
 * Инициализация функций масштабирования и навигации по временной шкале
 */
function initTimelineControls() {
    // Масштабирование временной шкалы будет реализовано в будущих версиях
    console.log('Инициализация контролов временной шкалы');
}

/**
 * Обработчики событий
 */

// Кнопка прямого эфира
liveBtn.addEventListener('click', startLiveStream);

// Кнопка снимка
snapshotBtn.addEventListener('click', takeSnapshot);

// Кнопка поиска записей
searchBtn.addEventListener('click', searchRecordings);

// Закрытие модального окна со снимком
closeBtn.addEventListener('click', () => {
    snapshotModal.style.display = 'none';
});

closeSnapshotBtn.addEventListener('click', () => {
    snapshotModal.style.display = 'none';
});

// Закрытие модального окна при клике вне его содержимого
window.addEventListener('click', (event) => {
    if (event.target === snapshotModal) {
        snapshotModal.style.display = 'none';
    }
});

// Инициализация при загрузке страницы
initTimelineControls();

// Периодическая проверка статуса сервера каждые 30 секунд
checkServerStatus();
setInterval(checkServerStatus, 30000);

// Обработка закрытия страницы
window.addEventListener('beforeunload', () => {
    stopCurrentStream();
});

// Предотвращение разрыва соединения при неактивности
setInterval(() => {
    if (currentRTCPeerConnection && connectionStatus === 'online') {
        fetch('/api/ping', { method: 'GET' }).catch(() => {});
    }
}, 60000);

// Добавление стилей для динамических элементов
const style = document.createElement('style');
style.textContent = `
    .playback-info {
        position: absolute;
        top: 10px;
        left: 10px;
        background-color: rgba(0, 0, 0, 0.7);
        color: white;
        padding: 5px 10px;
        border-radius: 4px;
        font-size: 12px;
        pointer-events: none;
    }
`;
document.head.appendChild(style);
});