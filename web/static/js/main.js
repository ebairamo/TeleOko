/**
 * TeleOko v2.0 HLS - JavaScript для HLS стриминга
 * Работает на ВСЕХ устройствах без WebRTC
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
    const loadingOverlay = document.getElementById('loadingOverlay');
    const loadingMessage = document.getElementById('loadingMessage');
    
    // Текущее состояние
    let currentVideoElement = null;
    let currentHls = null;
    let currentChannelID = null;
    let recordings = [];
    let connectionStatus = 'offline';
    
    // Установка текущей даты
    const today = new Date();
    const dd = String(today.getDate()).padStart(2, '0');
    const mm = String(today.getMonth() + 1).padStart(2, '0');
    const yyyy = today.getFullYear();
    archiveDate.value = dd + '.' + mm + '.' + yyyy;
    
    /**
     * Показать загрузку
     */
    function showLoading(message) {
        if (!message) message = 'Загрузка...';
        loadingMessage.textContent = message;
        loadingOverlay.style.display = 'flex';
    }
    
    /**
     * Скрыть загрузку
     */
    function hideLoading() {
        loadingOverlay.style.display = 'none';
    }
    
    /**
     * Обновить статус подключения
     */
    function updateConnectionStatus(status) {
        connectionStatus = status;
        const statusElement = document.querySelector('.connection-status');
        if (statusElement) {
            statusElement.className = 'connection-status ' + status;
            statusElement.textContent = status === 'online' ? '🟢 Подключено' : '🔴 Не подключено';
        }
    }
    
    /**
     * Показать ошибку
     */
    function showError(container, message) {
        container.innerHTML = `
            <div class="error">
                <p>❌ ${message}</p>
                <button onclick="location.reload()" class="secondary-btn" style="margin-top: 10px;">
                    🔄 Перезагрузить страницу
                </button>
            </div>
        `;
    }
    
    /**
     * Форматирование даты
     */
    function formatDateTime(dateTimeString) {
        try {
            const date = new Date(dateTimeString);
            return date.toLocaleString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
        } catch (e) {
            return dateTimeString;
        }
    }
    
    /**
     * Расчет продолжительности
     */
    function calculateDuration(startTime, endTime) {
        try {
            const start = new Date(startTime);
            const end = new Date(endTime);
            const diffMs = end - start;
            
            const hours = Math.floor(diffMs / (1000 * 60 * 60));
            const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
            const seconds = Math.floor((diffMs % (1000 * 60)) / 1000);
            
            if (hours > 0) {
                return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
            }
            return `${minutes}:${seconds.toString().padStart(2, '0')}`;
        } catch (e) {
            return '00:00';
        }
    }
    
    /**
     * Остановить текущий поток
     */
    function stopCurrentStream() {
        if (currentHls) {
            currentHls.destroy();
            currentHls = null;
        }
        
        if (currentVideoElement) {
            currentVideoElement.pause();
            currentVideoElement.src = '';
            currentVideoElement.load();
            currentVideoElement = null;
        }
        
        currentChannelID = null;
        updateConnectionStatus('offline');
        console.log('🛑 Поток остановлен');
    }
    
    /**
     * Запустить прямой эфир
     */
    async function startLiveStream() {
        const channelId = cameraSelect.value;
        if (!channelId) {
            alert('Выберите канал для просмотра');
            return;
        }
        
        showLoading('Запуск потока...');
        stopCurrentStream();
        
        try {
            // Получаем информацию о потоке
            const response = await fetch(`/api/stream/${channelId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            const streamData = await response.json();
            
            if (streamData.type !== 'hls') {
                throw new Error('Сервер не настроен для HLS стриминга');
            }
            
            currentChannelID = channelId;
            
            // Ждем несколько секунд, чтобы поток успел запуститься
            if (streamData.status === 'starting') {
                showLoading('Запуск потока, подождите 5 секунд...');
                await new Promise(resolve => setTimeout(resolve, 5000));
            }
            
            // Создаем видео элемент
            const video = document.createElement('video');
            video.id = 'videoPlayer';
            video.controls = true;
            video.autoplay = true;
            video.muted = true;
            video.playsInline = true;
            video.style.width = '100%';
            video.style.height = '100%';
            video.style.objectFit = 'contain';
            video.style.backgroundColor = '#000';
            
            // Проверяем поддержку HLS
            const streamUrl = streamData.stream_url;
            
            if (video.canPlayType('application/vnd.apple.mpegurl')) {
                // Нативная поддержка HLS (iOS Safari)
                console.log('📱 Используем нативную поддержку HLS');
                video.src = streamUrl;
                
                video.addEventListener('loadedmetadata', () => {
                    console.log('✅ Метаданные загружены');
                    hideLoading();
                    updateConnectionStatus('online');
                });
                
                video.addEventListener('error', (e) => {
                    console.error('❌ Ошибка видео:', e);
                    showError(videoContainer, 'Ошибка загрузки видео. Попробуйте еще раз.');
                    updateConnectionStatus('offline');
                });
                
            } else if (Hls && Hls.isSupported()) {
                // Используем HLS.js для остальных браузеров
                console.log('💻 Используем HLS.js');
                const hls = new Hls({
                    debug: false,
                    enableWorker: true,
                    lowLatencyMode: true,
                    backBufferLength: 90,
                    maxBufferSize: 60 * 1000 * 1000,
                    maxBufferLength: 60,
                    startLevel: -1
                });
                
                currentHls = hls;
                
                hls.loadSource(streamUrl);
                hls.attachMedia(video);
                
                hls.on(Hls.Events.MANIFEST_PARSED, () => {
                    console.log('✅ HLS манифест загружен');
                    video.play().catch(e => {
                        console.warn('Автовоспроизведение заблокировано:', e);
                        // Показываем кнопку play
                        showPlayButton(videoContainer, video);
                    });
                    hideLoading();
                    updateConnectionStatus('online');
                });
                
                hls.on(Hls.Events.ERROR, (event, data) => {
                    console.error('❌ HLS ошибка:', data);
                    if (data.fatal) {
                        switch (data.type) {
                            case Hls.ErrorTypes.NETWORK_ERROR:
                                console.error('Сетевая ошибка, пробуем восстановить...');
                                showLoading('Восстановление соединения...');
                                hls.startLoad();
                                break;
                            case Hls.ErrorTypes.MEDIA_ERROR:
                                console.error('Ошибка медиа, пробуем восстановить...');
                                hls.recoverMediaError();
                                break;
                            default:
                                showError(videoContainer, 'Критическая ошибка потока');
                                hls.destroy();
                                break;
                        }
                    }
                });
                
            } else {
                throw new Error('Ваш браузер не поддерживает HLS воспроизведение');
            }
            
            // Очищаем контейнер и добавляем видео
            videoContainer.innerHTML = '';
            videoContainer.appendChild(video);
            currentVideoElement = video;
            
            // Добавляем информационную панель
            const infoPanel = document.createElement('div');
            infoPanel.className = 'video-info-panel';
            infoPanel.innerHTML = `
                <div class="video-info">
                    <span>📺 ${streamData.channel_name || `Канал ${channelId}`}</span>
                    <span>🔴 Прямой эфир (HLS)</span>
                    <span>📱 Работает на всех устройствах</span>
                </div>
            `;
            videoContainer.appendChild(infoPanel);
            
            // Кнопка остановки потока
            const stopButton = document.createElement('button');
            stopButton.className = 'control-btn';
            stopButton.style.position = 'absolute';
            stopButton.style.bottom = '20px';
            stopButton.style.right = '20px';
            stopButton.style.padding = '10px 20px';
            stopButton.style.backgroundColor = '#e74c3c';
            stopButton.style.color = 'white';
            stopButton.style.border = 'none';
            stopButton.style.borderRadius = '5px';
            stopButton.style.cursor = 'pointer';
            stopButton.style.zIndex = '100';
            stopButton.textContent = '⏹️ Остановить';
            stopButton.onclick = () => {
                stopCurrentStream();
                videoContainer.innerHTML = `
                    <div class="placeholder">
                        <div class="placeholder-icon">📹</div>
                        <h3>Поток остановлен</h3>
                        <p>Выберите канал и нажмите "Прямой эфир" для начала просмотра</p>
                    </div>
                `;
            };
            videoContainer.appendChild(stopButton);
            
        } catch (error) {
            console.error('Ошибка запуска потока:', error);
            showError(videoContainer, `Не удалось запустить поток: ${error.message}`);
            updateConnectionStatus('offline');
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Показать кнопку воспроизведения
     */
    function showPlayButton(container, video) {
        const playOverlay = document.createElement('div');
        playOverlay.style.position = 'absolute';
        playOverlay.style.top = '0';
        playOverlay.style.left = '0';
        playOverlay.style.width = '100%';
        playOverlay.style.height = '100%';
        playOverlay.style.backgroundColor = 'rgba(0, 0, 0, 0.5)';
        playOverlay.style.display = 'flex';
        playOverlay.style.alignItems = 'center';
        playOverlay.style.justifyContent = 'center';
        playOverlay.style.cursor = 'pointer';
        playOverlay.style.zIndex = '50';
        
        const playButton = document.createElement('div');
        playButton.style.width = '80px';
        playButton.style.height = '80px';
        playButton.style.backgroundColor = 'rgba(255, 255, 255, 0.8)';
        playButton.style.borderRadius = '50%';
        playButton.style.display = 'flex';
        playButton.style.alignItems = 'center';
        playButton.style.justifyContent = 'center';
        playButton.innerHTML = `
            <svg width="40" height="40" viewBox="0 0 24 24" fill="#333">
                <path d="M8 5v14l11-7z"/>
            </svg>
        `;
        
        playOverlay.appendChild(playButton);
        
        playOverlay.addEventListener('click', () => {
            video.play();
            playOverlay.remove();
        });
        
        container.appendChild(playOverlay);
    }
    
    /**
     * Получить снимок
     */
    async function takeSnapshot() {
        const channelId = cameraSelect.value;
        if (!channelId) {
            alert('Выберите канал для создания снимка');
            return;
        }
        
        try {
            showLoading('Создание снимка...');
            
            const response = await fetch(`/api/snapshot/${channelId}`);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            const blob = await response.blob();
            const imageUrl = URL.createObjectURL(blob);
            
            const link = document.createElement('a');
            link.href = imageUrl;
            link.download = `snapshot_${channelId}_${new Date().toISOString().replace(/[:.]/g, '-')}.jpg`;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            
            URL.revokeObjectURL(imageUrl);
            
        } catch (error) {
            console.error('Ошибка создания снимка:', error);
            alert(`Не удалось создать снимок: ${error.message}`);
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Поиск архивных записей
     */
    async function searchRecordings() {
        const channelId = cameraSelect.value;
        const date = archiveDate.value;
        
        if (!channelId || !date) {
            alert('Выберите канал и дату для поиска');
            return;
        }
        
        showLoading('Поиск записей...');
        recordingsList.innerHTML = '<div class="loading">Поиск записей...</div>';
        timeline.innerHTML = '<div class="loading">Загрузка временной шкалы...</div>';
        
        try {
            const response = await fetch(`/api/recordings?channel=${channelId}&start=${date}&end=${date}`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.error) {
                throw new Error(data.error);
            }
            
            recordings = data.recordings || [];
            displayRecordings(recordings);
            displayTimeline(recordings, date);
            
        } catch (error) {
            console.error('Ошибка поиска записей:', error);
            showError(recordingsList, `Не удалось найти записи: ${error.message}`);
            showError(timeline, 'Ошибка загрузки временной шкалы');
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Отображение записей
     */
    function displayRecordings(recordings) {
        recordingsList.innerHTML = '';
        
        if (!recordings || recordings.length === 0) {
            recordingsList.innerHTML = '<div class="recordings-empty">📁 Записи не найдены</div>';
            return;
        }
        
        recordings.sort((a, b) => new Date(b.StartTime) - new Date(a.StartTime));
        
        recordings.forEach(recording => {
            const recordingItem = document.createElement('div');
            recordingItem.className = 'recording-item';
            
            const startTime = formatDateTime(recording.StartTime);
            const endTime = formatDateTime(recording.EndTime);
            const duration = calculateDuration(recording.StartTime, recording.EndTime);
            
            recordingItem.innerHTML = `
                <div class="recording-info">
                    <span class="recording-time">📅 ${startTime}</span>
                    <span class="recording-duration">⏱️ ${duration}</span>
                </div>
                <div class="recording-actions">
                    <button class="play-btn primary-btn" onclick="playRecording('${recording.StartTime}', '${recording.EndTime}', '${recording.Channel}')">
                        ▶️ Воспроизвести
                    </button>
                </div>
            `;
            
            recordingsList.appendChild(recordingItem);
        });
    }
    
    /**
     * Отображение временной шкалы
     */
    function displayTimeline(recordings, date) {
        timeline.innerHTML = '';
        
        if (!recordings || recordings.length === 0) {
            timeline.innerHTML = '<div class="timeline-empty">📊 Нет данных для отображения</div>';
            return;
        }
        
        const timelineContainer = document.createElement('div');
        timelineContainer.className = 'timeline-inner';
        timelineContainer.style.position = 'relative';
        timelineContainer.style.height = '60px';
        timelineContainer.style.background = '#f5f5f5';
        timelineContainer.style.borderRadius = '4px';
        
        // Добавляем сетку времени
        for (let hour = 0; hour < 24; hour++) {
            const hourLine = document.createElement('div');
            hourLine.className = 'hour-line';
            hourLine.style.position = 'absolute';
            hourLine.style.left = `${(hour / 24) * 100}%`;
            hourLine.style.top = '0';
            hourLine.style.bottom = '0';
            hourLine.style.width = '1px';
            hourLine.style.background = '#ddd';
            timelineContainer.appendChild(hourLine);
            
            if (hour % 3 === 0) {
                const timeLabel = document.createElement('div');
                timeLabel.className = 'time-label';
                timeLabel.style.position = 'absolute';
                timeLabel.style.left = `${(hour / 24) * 100}%`;
                timeLabel.style.top = '-20px';
                timeLabel.style.fontSize = '11px';
                timeLabel.style.color = '#666';
                timeLabel.style.transform = 'translateX(-50%)';
                timeLabel.textContent = `${hour.toString().padStart(2, '0')}:00`;
                timelineContainer.appendChild(timeLabel);
            }
        }
        
        // Отображаем записи
        const dateParts = date.split('.');
        const dayStart = new Date(`${dateParts[2]}-${dateParts[1]}-${dateParts[0]}T00:00:00`);
        const dayEnd = new Date(`${dateParts[2]}-${dateParts[1]}-${dateParts[0]}T23:59:59`);
        const dayDuration = dayEnd - dayStart;
        
        recordings.forEach((recording, index) => {
            const startTime = new Date(recording.StartTime);
            const endTime = new Date(recording.EndTime);
            
            const startPosition = ((startTime - dayStart) / dayDuration) * 100;
            const width = ((endTime - startTime) / dayDuration) * 100;
            
            if (startPosition >= 0 && startPosition <= 100) {
                const segment = document.createElement('div');
                segment.className = 'timeline-segment';
                segment.style.position = 'absolute';
                segment.style.left = `${Math.max(0, startPosition)}%`;
                segment.style.width = `${Math.min(width, 100 - startPosition)}%`;
                segment.style.height = '30px';
                segment.style.top = '15px';
                segment.style.background = '#3498db';
                segment.style.cursor = 'pointer';
                segment.style.borderRadius = '3px';
                segment.style.transition = 'all 0.2s';
                
                const tooltip = `${formatDateTime(recording.StartTime)} - ${formatDateTime(recording.EndTime)}`;
                segment.title = tooltip;
                
                segment.addEventListener('mouseenter', () => {
                    segment.style.background = '#2980b9';
                    segment.style.transform = 'scaleY(1.2)';
                });
                
                segment.addEventListener('mouseleave', () => {
                    segment.style.background = '#3498db';
                    segment.style.transform = 'scaleY(1)';
                });
                
                segment.onclick = () => {
                    playRecording(recording.StartTime, recording.EndTime, recording.Channel);
                };
                
                timelineContainer.appendChild(segment);
            }
        });
        
        timeline.appendChild(timelineContainer);
    }
    
    /**
     * Воспроизведение записи
     */
    window.playRecording = async function(startTime, endTime, channelId) {
        showLoading('Загрузка архивной записи...');
        
        try {
            const response = await fetch(`/api/playback-url?channel=${channelId}&start=${startTime}&end=${endTime}`);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.error) {
                throw new Error(data.error);
            }
            
            videoContainer.innerHTML = `
                <div class="playback-info-container">
                    <div class="playback-info">
                        <h3>📼 Архивная запись</h3>
                        <p><strong>Время:</strong> ${formatDateTime(startTime)} - ${formatDateTime(endTime)}</p>
                        <p><strong>Канал:</strong> ${channelId}</p>
                        <p><strong>RTSP URL:</strong></p>
                        <code style="word-break: break-all;">${data.url}</code>
                        <p><em>💡 Используйте VLC Player для воспроизведения</em></p>
                        <button onclick="copyToClipboard('${data.url}')" class="primary-btn" style="margin-top: 10px;">
                            📋 Копировать URL
                        </button>
                        <button onclick="location.reload()" class="secondary-btn" style="margin-top: 10px; margin-left: 10px;">
                            🔙 Вернуться
                        </button>
                    </div>
                </div>
            `;
            
        } catch (error) {
            console.error('Ошибка воспроизведения архива:', error);
            showError(videoContainer, `Не удалось загрузить запись: ${error.message}`);
        } finally {
            hideLoading();
        }
    };
    
    /**
     * Копирование в буфер обмена
     */
    window.copyToClipboard = function(text) {
        if (navigator.clipboard) {
            navigator.clipboard.writeText(text).then(() => {
                alert('URL скопирован в буфер обмена!');
            }).catch(() => {
                fallbackCopyToClipboard(text);
            });
        } else {
            fallbackCopyToClipboard(text);
        }
    };
    
    function fallbackCopyToClipboard(text) {
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        try {
            document.execCommand('copy');
            alert('URL скопирован в буфер обмена!');
        } catch (err) {
            alert('Не удалось скопировать URL');
        }
        document.body.removeChild(textArea);
    }
    
    /**
     * Инициализация событий
     */
    function initEventHandlers() {
        if (liveBtn) {
            liveBtn.addEventListener('click', startLiveStream);
        }
        
        if (snapshotBtn) {
            snapshotBtn.addEventListener('click', takeSnapshot);
        }
        
        if (searchBtn) {
            searchBtn.addEventListener('click', searchRecordings);
        }
        
        if (archiveDate) {
            archiveDate.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    searchRecordings();
                }
            });
        }
        
        window.addEventListener('beforeunload', () => {
            stopCurrentStream();
        });
    }
    
    /**
     * Проверка статуса системы
     */
    async function checkSystemStatus() {
        try {
            const response = await fetch('/api/info');
            if (response.ok) {
                const data = await response.json();
                updateConnectionStatus(data.status || 'online');
                
                // Проверяем, что сервер работает в HLS режиме
                if (data.streaming_type !== 'HLS') {
                    console.warn('⚠️ Сервер не в HLS режиме:', data.streaming_type);
                }
            } else {
                updateConnectionStatus('offline');
            }
        } catch (error) {
            updateConnectionStatus('offline');
        }
    }
    
    /**
     * Инициализация
     */
    function init() {
        console.log('🚀 TeleOko v2.0 HLS инициализирован');
        
        checkSystemStatus();
        initEventHandlers();
        
        // Периодическая проверка статуса
        setInterval(checkSystemStatus, 30000);
        
        // Начальное сообщение
        if (videoContainer) {
            videoContainer.innerHTML = `
                <div class="placeholder">
                    <div class="placeholder-icon">📹</div>
                    <h3>Добро пожаловать в TeleOko HLS</h3>
                    <p>Выберите канал и нажмите "Прямой эфир" для начала просмотра</p>
                    <p><small>✅ Работает на всех устройствах без настройки</small></p>
                </div>
            `;
        }
    }
    
    // Запуск
    init();
});