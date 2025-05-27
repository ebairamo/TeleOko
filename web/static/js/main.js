/**
 * TeleOko v2.0 - Полностью исправленный JavaScript
 * Система видеонаблюдения с поддержкой WebRTC
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
    
    // Текущее состояние приложения
    let currentVideoElement = null;
    let currentRTCPeerConnection = null;
    let currentStream = null;
    let recordings = [];
    let connectionStatus = 'offline';
    
    // Установка текущей даты по умолчанию (формат dd.mm.yyyy)
    const today = new Date();
    const dd = String(today.getDate()).padStart(2, '0');
    const mm = String(today.getMonth() + 1).padStart(2, '0');
    const yyyy = today.getFullYear();
    archiveDate.value = dd + '.' + mm + '.' + yyyy;
    
    /**
     * Отображение индикатора загрузки
     */
    function showLoading(message) {
        if (!message) message = 'Загрузка...';
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
        const statusElement = document.querySelector('.connection-status');
        if (statusElement) {
            statusElement.className = 'connection-status ' + status;
            statusElement.textContent = status === 'online' ? 'Подключено' : 'Не подключено';
        }
    }
    
    /**
     * Отображение ошибки
     */
    function showError(container, message) {
        container.innerHTML = '<div class="error"><p>❌ ' + message + '</p></div>';
    }
    
    /**
     * Форматирование даты и времени
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
     * Расчет продолжительности записи
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
                return hours + ':' + minutes.toString().padStart(2, '0') + ':' + seconds.toString().padStart(2, '0');
            }
            return minutes + ':' + seconds.toString().padStart(2, '0');
        } catch (e) {
            return '00:00';
        }
    }
    
    /**
     * Запуск прямого эфира
     */
    async function startLiveStream() {
        const channelId = cameraSelect.value;
        if (!channelId) {
            alert('Выберите канал для просмотра');
            return;
        }
        
        showLoading('Подключение к камере...');
        stopCurrentStream();
        
        try {
            // Получаем информацию о потоке
            const streamResponse = await fetch('/api/stream/' + channelId);
            if (!streamResponse.ok) {
                throw new Error('HTTP ' + streamResponse.status);
            }
            
            const streamData = await streamResponse.json();
            
            if (streamData.type === 'webrtc') {
                await startWebRTCStream(channelId, streamData);
            } else {
                // Для RTSP показываем сообщение
                showError(videoContainer, 'WebRTC недоступен. Используйте VLC для просмотра RTSP: ' + streamData.rtsp_url);
            }
            
        } catch (error) {
            console.error('Ошибка запуска прямого эфира:', error);
            showError(videoContainer, 'Не удалось подключиться к камере: ' + error.message);
            updateConnectionStatus('offline');
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Запуск WebRTC потока
     */
    async function startWebRTCStream(channelId, streamData) {
        try {
            // Создаем видео элемент
            const videoElement = document.createElement('video');
            videoElement.autoplay = true;
            videoElement.playsInline = true;
            videoElement.muted = true;
            videoElement.style.width = '100%';
            videoElement.style.height = '100%';
            videoElement.style.objectFit = 'contain';
            
            // Настройка WebRTC с несколькими STUN серверами
            const pc = new RTCPeerConnection({
                iceServers: [
                    { urls: 'stun:stun.l.google.com:19302' },
                    { urls: 'stun:stun1.l.google.com:19302' },
                    { urls: 'stun:stun2.l.google.com:19302' }
                ],
                iceCandidatePoolSize: 10
            });
            
            currentRTCPeerConnection = pc;
            
            // Обработчики WebRTC событий
            pc.ontrack = function(event) {
                console.log('📺 Получен медиа-трек:', event.track.kind);
                if (event.streams && event.streams[0]) {
                    videoElement.srcObject = event.streams[0];
                    currentStream = event.streams[0];
                    updateConnectionStatus('online');
                    
                    // Добавляем обработчик загрузки метаданных
                    videoElement.onloadedmetadata = function() {
                        console.log('📐 Видео размер: ' + videoElement.videoWidth + 'x' + videoElement.videoHeight);
                    };
                }
            };
            
            pc.oniceconnectionstatechange = function() {
                console.log('🔌 ICE состояние:', pc.iceConnectionState);
                const connectionStates = {
                    'checking': 'Подключение...',
                    'connected': 'Подключено',
                    'completed': 'Подключено', 
                    'disconnected': 'Отключено',
                    'failed': 'Ошибка подключения',
                    'closed': 'Соединение закрыто'
                };
                
                const statusText = connectionStates[pc.iceConnectionState] || pc.iceConnectionState;
                console.log('📡 Статус: ' + statusText);
                
                if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
                    updateConnectionStatus('online');
                } else if (pc.iceConnectionState === 'disconnected' || pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'closed') {
                    updateConnectionStatus('offline');
                }
            };
            
            pc.onicecandidate = function(event) {
                if (event.candidate) {
                    console.log('🧊 ICE кандидат:', event.candidate.type);
                }
            };
            
            // Добавляем трансивер для получения видео
            pc.addTransceiver('video', { direction: 'recvonly' });
            
            // Создаем SDP offer
            const offer = await pc.createOffer({
                offerToReceiveVideo: true,
                offerToReceiveAudio: false,
                voiceActivityDetection: false
            });
            
            await pc.setLocalDescription(offer);
            console.log('📋 SDP Offer создан');
            
            // Отправляем offer на сервер
            const response = await fetch('/api/webrtc/offer?channel=' + channelId, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    type: offer.type,
                    sdp: offer.sdp
                })
            });
            
            if (!response.ok) {
                throw new Error('HTTP ' + response.status + ': ' + response.statusText);
            }
            
            const answer = await response.json();
            console.log('📨 Получен SDP Answer');
            
            if (answer.error) {
                throw new Error(answer.error);
            }
            
            // Устанавливаем удаленное описание
            if (answer.sdp) {
                await pc.setRemoteDescription(new RTCSessionDescription({
                    type: answer.type || 'answer',
                    sdp: answer.sdp
                }));
                console.log('✅ WebRTC соединение настроено');
            }
            
            // Добавляем видео в контейнер
            videoContainer.innerHTML = '';
            videoContainer.appendChild(videoElement);
            currentVideoElement = videoElement;
            
            // Добавляем информационную панель
            const infoPanel = document.createElement('div');
            infoPanel.className = 'video-info-panel';
            infoPanel.innerHTML = 
                '<div class="video-info">' +
                    '<span>📺 ' + (streamData.channel_name || 'Канал ' + channelId) + '</span>' +
                    '<span>🔴 Прямой эфир</span>' +
                    '<span id="video-quality">📐 Загрузка...</span>' +
                '</div>';
            videoContainer.appendChild(infoPanel);
            
            // Обновляем информацию о качестве видео
            videoElement.addEventListener('loadedmetadata', function() {
                const qualityInfo = document.getElementById('video-quality');
                if (qualityInfo) {
                    qualityInfo.textContent = '📐 ' + videoElement.videoWidth + 'x' + videoElement.videoHeight;
                }
            });
            
            // Добавляем обработчик ошибок видео
            videoElement.addEventListener('error', function(e) {
                console.error('❌ Ошибка видео:', e);
                updateConnectionStatus('offline');
            });
            
        } catch (error) {
            console.error('❌ WebRTC ошибка:', error);
            throw new Error('WebRTC ошибка: ' + error.message);
        }
    }
    
    /**
     * Остановка текущего потока
     */
    function stopCurrentStream() {
        if (currentRTCPeerConnection) {
            currentRTCPeerConnection.close();
            currentRTCPeerConnection = null;
        }
        
        if (currentStream) {
            currentStream.getTracks().forEach(function(track) {
                track.stop();
            });
            currentStream = null;
        }
        
        if (currentVideoElement) {
            currentVideoElement.srcObject = null;
            currentVideoElement = null;
        }
        
        updateConnectionStatus('offline');
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
            const response = await fetch('/api/recordings?channel=' + channelId + '&start=' + date + '&end=' + date);
            
            if (!response.ok) {
                throw new Error('HTTP ' + response.status);
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
            showError(recordingsList, 'Не удалось найти записи: ' + error.message);
            showError(timeline, 'Ошибка загрузки временной шкалы: ' + error.message);
        } finally {
            hideLoading();
        }
    }
    
    /**
     * Отображение списка записей
     */
    function displayRecordings(recordings) {
        recordingsList.innerHTML = '';
        
        if (!recordings || recordings.length === 0) {
            recordingsList.innerHTML = '<div class="recordings-empty">📁 Записи не найдены</div>';
            return;
        }
        
        // Сортируем записи по времени (новые сначала)
        recordings.sort(function(a, b) {
            return new Date(b.StartTime) - new Date(a.StartTime);
        });
        
        recordings.forEach(function(recording) {
            const recordingItem = document.createElement('div');
            recordingItem.className = 'recording-item';
            
            const startTime = formatDateTime(recording.StartTime);
            const endTime = formatDateTime(recording.EndTime);
            const duration = calculateDuration(recording.StartTime, recording.EndTime);
            
            recordingItem.innerHTML = 
                '<div class="recording-info">' +
                    '<span class="recording-time">📅 ' + startTime + '</span>' +
                    '<span class="recording-duration">⏱️ ' + duration + '</span>' +
                    '<span class="recording-end">🏁 ' + endTime + '</span>' +
                '</div>' +
                '<div class="recording-actions">' +
                    '<button class="play-btn primary-btn" onclick="playRecording(\'' + recording.StartTime + '\', \'' + recording.EndTime + '\', \'' + recording.Channel + '\')">' +
                        '▶️ Воспроизвести' +
                    '</button>' +
                '</div>';
            
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
        
        // Создаем контейнер временной шкалы
        const timelineContainer = document.createElement('div');
        timelineContainer.className = 'timeline-container';
        timelineContainer.style.position = 'relative';
        timelineContainer.style.height = '80px';
        timelineContainer.style.background = '#f5f5f5';
        timelineContainer.style.borderRadius = '4px';
        timelineContainer.style.overflow = 'hidden';
        
        // Добавляем часовые метки
        for (let hour = 0; hour < 24; hour += 2) {
            const timeLabel = document.createElement('div');
            timeLabel.className = 'time-label';
            timeLabel.style.position = 'absolute';
            timeLabel.style.left = ((hour / 24) * 100) + '%';
            timeLabel.style.top = '5px';
            timeLabel.style.fontSize = '10px';
            timeLabel.style.color = '#666';
            timeLabel.style.transform = 'translateX(-50%)';
            timeLabel.textContent = hour.toString().padStart(2, '0') + ':00';
            timelineContainer.appendChild(timeLabel);
        }
        
        // Отображаем записи на шкале
        const dateParts = date.split('.');
        const dayStart = new Date(dateParts[2] + '-' + dateParts[1] + '-' + dateParts[0] + 'T00:00:00');
        const dayEnd = new Date(dateParts[2] + '-' + dateParts[1] + '-' + dateParts[0] + 'T23:59:59');
        const dayDuration = dayEnd - dayStart;
        
        recordings.forEach(function(recording, index) {
            const startTime = new Date(recording.StartTime);
            const endTime = new Date(recording.EndTime);
            
            // Рассчитываем позицию и ширину сегмента
            const startPosition = ((startTime - dayStart) / dayDuration) * 100;
            const width = ((endTime - startTime) / dayDuration) * 100;
            
            if (startPosition >= 0 && startPosition <= 100) {
                const segment = document.createElement('div');
                segment.className = 'timeline-segment';
                segment.style.position = 'absolute';
                segment.style.left = Math.max(0, startPosition) + '%';
                segment.style.width = Math.min(width, 100 - startPosition) + '%';
                segment.style.height = '30px';
                segment.style.top = '35px';
                segment.style.background = 'hsl(' + ((index * 137.5) % 360) + ', 70%, 50%)';
                segment.style.cursor = 'pointer';
                segment.style.borderRadius = '2px';
                segment.style.boxShadow = '0 1px 3px rgba(0,0,0,0.3)';
                
                // Добавляем всплывающую подсказку
                segment.title = formatDateTime(recording.StartTime) + ' - ' + formatDateTime(recording.EndTime);
                
                // Обработчик клика
                segment.onclick = function() {
                    playRecording(recording.StartTime, recording.EndTime, recording.Channel);
                };
                
                timelineContainer.appendChild(segment);
            }
        });
        
        timeline.appendChild(timelineContainer);
    }
    
    /**
     * Воспроизведение архивной записи
     */
    window.playRecording = async function(startTime, endTime, channelId) {
        showLoading('Загрузка архивной записи...');
        stopCurrentStream();
        
        try {
            // Получаем URL для воспроизведения
            const response = await fetch('/api/playback-url?channel=' + channelId + '&start=' + startTime + '&end=' + endTime);
            
            if (!response.ok) {
                throw new Error('HTTP ' + response.status);
            }
            
            const data = await response.json();
            
            if (data.error) {
                throw new Error(data.error);
            }
            
            // Показываем информацию об RTSP URL
            videoContainer.innerHTML = 
                '<div class="playback-info-container">' +
                    '<div class="playback-info">' +
                        '<h3>📼 Архивная запись</h3>' +
                        '<p><strong>Время:</strong> ' + formatDateTime(startTime) + ' - ' + formatDateTime(endTime) + '</p>' +
                        '<p><strong>Канал:</strong> ' + channelId + '</p>' +
                        '<p><strong>RTSP URL:</strong></p>' +
                        '<code style="word-break: break-all; background: #f5f5f5; padding: 10px; border-radius: 4px; display: block; margin: 10px 0;">' +
                            data.url +
                        '</code>' +
                        '<p><em>💡 Используйте VLC Player для воспроизведения этого URL</em></p>' +
                        '<button onclick="copyToClipboard(\'' + data.url + '\')" class="primary-btn" style="margin-top: 10px;">' +
                            '📋 Копировать URL' +
                        '</button>' +
                    '</div>' +
                '</div>';
            
        } catch (error) {
            console.error('Ошибка воспроизведения архива:', error);
            showError(videoContainer, 'Не удалось загрузить запись: ' + error.message);
        } finally {
            hideLoading();
        }
    };
    
    /**
     * Копирование в буфер обмена
     */
    window.copyToClipboard = function(text) {
        if (navigator.clipboard) {
            navigator.clipboard.writeText(text).then(function() {
                alert('URL скопирован в буфер обмена!');
            }).catch(function() {
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
            alert('Не удалось скопировать URL. Скопируйте вручную.');
        }
        document.body.removeChild(textArea);
    }
    
    /**
     * Получение снимка с камеры
     */
    async function takeSnapshot() {
        const channelId = cameraSelect.value;
        if (!channelId) {
            alert('Выберите канал для создания снимка');
            return;
        }
        
        try {
            showLoading('Создание снимка...');
            
            const response = await fetch('/api/snapshot/' + channelId);
            if (!response.ok) {
                throw new Error('HTTP ' + response.status);
            }
            
            // Создаем blob из ответа
            const blob = await response.blob();
            const imageUrl = URL.createObjectURL(blob);
            
            // Создаем ссылку для скачивания
            const link = document.createElement('a');
            link.href = imageUrl;
            link.download = 'snapshot_' + channelId + '_' + new Date().toISOString().replace(/[:.]/g, '-') + '.jpg';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            
            // Освобождаем память
            URL.revokeObjectURL(imageUrl);
            
            alert('Снимок сохранен!');
            
        } catch (error) {
            console.error('Ошибка создания снимка:', error);
            alert('Не удалось создать снимок: ' + error.message);
        } finally {
            hideLoading();
        }
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
            } else {
                updateConnectionStatus('offline');
            }
        } catch (error) {
            updateConnectionStatus('offline');
        }
    }
    
    /**
     * Инициализация обработчиков событий
     */
    function initEventHandlers() {
        // Кнопка прямого эфира
        if (liveBtn) {
            liveBtn.addEventListener('click', startLiveStream);
        }
        
        // Кнопка снимка
        if (snapshotBtn) {
            snapshotBtn.addEventListener('click', takeSnapshot);
        }
        
        // Кнопка поиска записей
        if (searchBtn) {
            searchBtn.addEventListener('click', searchRecordings);
        }
        
        // Обработка закрытия страницы
        window.addEventListener('beforeunload', function() {
            stopCurrentStream();
        });
        
        // Предотвращение разрыва соединения при неактивности
        setInterval(function() {
            if (currentRTCPeerConnection && connectionStatus === 'online') {
                fetch('/api/ping').catch(function() {});
            }
        }, 30000);
    }
    
    // Инициализация приложения
    function init() {
        console.log('🚀 TeleOko v2.0 инициализирован');
        
        // Проверяем статус системы
        checkSystemStatus();
        
        // Инициализируем обработчики
        initEventHandlers();
        
        // Периодическая проверка статуса
        setInterval(checkSystemStatus, 30000);
        
        // Показываем начальное сообщение
        if (videoContainer) {
            videoContainer.innerHTML = 
                '<div class="placeholder">' +
                    '<div class="placeholder-icon">📹</div>' +
                    '<h3>Добро пожаловать в TeleOko</h3>' +
                    '<p>Выберите канал и нажмите "Прямой эфир" для начала просмотра</p>' +
                    '<p><small>Или выберите дату и нажмите "Поиск записей" для просмотра архива</small></p>' +
                '</div>';
        }
    }
    
    // Запуск приложения
    init();
});