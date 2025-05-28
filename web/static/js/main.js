/**
 * TeleOko v2.0 - JavaScript с полноэкранным режимом
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
    let isFullscreen = false;
    
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
     * Переключение полноэкранного режима
     */
    function toggleFullscreen() {
        if (!isFullscreen) {
            if (videoContainer.requestFullscreen) {
                videoContainer.requestFullscreen();
            } else if (videoContainer.webkitRequestFullscreen) {
                videoContainer.webkitRequestFullscreen();
            } else if (videoContainer.mozRequestFullScreen) {
                videoContainer.mozRequestFullScreen();
            } else if (videoContainer.msRequestFullscreen) {
                videoContainer.msRequestFullscreen();
            }
            isFullscreen = true;
        } else {
            if (document.exitFullscreen) {
                document.exitFullscreen();
            } else if (document.webkitExitFullscreen) {
                document.webkitExitFullscreen();
            } else if (document.mozCancelFullScreen) {
                document.mozCancelFullScreen();
            } else if (document.msExitFullscreen) {
                document.msExitFullscreen();
            }
            isFullscreen = false;
        }
    }
    
    /**
     * Добавление кнопок управления на видео
     */
    function addVideoControls(container) {
        // Создаем контейнер для кнопок
        const controlsDiv = document.createElement('div');
        controlsDiv.className = 'video-controls';
        controlsDiv.innerHTML = `
            <button class="control-btn fullscreen-btn" title="Полноэкранный режим">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="white">
                    <path d="M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z"/>
                </svg>
            </button>
            <button class="control-btn snapshot-btn" title="Сделать снимок">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="white">
                    <circle cx="12" cy="12" r="3.2"/>
                    <path d="M9 2L7.17 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2h-3.17L15 2H9zm3 15c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5z"/>
                </svg>
            </button>
        `;
        
        container.appendChild(controlsDiv);
        
        // Обработчики кнопок
        controlsDiv.querySelector('.fullscreen-btn').addEventListener('click', toggleFullscreen);
        controlsDiv.querySelector('.snapshot-btn').addEventListener('click', () => {
            const channelId = cameraSelect.value;
            if (channelId) takeSnapshot(channelId);
        });
        
        // Показываем/скрываем контролы при наведении
        let hideTimeout;
        container.addEventListener('mouseenter', () => {
            clearTimeout(hideTimeout);
            controlsDiv.style.opacity = '1';
        });
        
        container.addEventListener('mouseleave', () => {
            hideTimeout = setTimeout(() => {
                controlsDiv.style.opacity = '0';
            }, 2000);
        });
        
        // Двойной клик для полноэкранного режима
        if (currentVideoElement) {
            currentVideoElement.addEventListener('dblclick', toggleFullscreen);
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
        videoElement.style.backgroundColor = '#000';
        
        // Отображаем подробные логи для отладки
        console.log(`📹 Начинаем установку WebRTC для канала ${channelId}`);
        
        // Настройка WebRTC с большим количеством STUN серверов
        const pc = new RTCPeerConnection({
            iceServers: [
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                { urls: 'stun:stun2.l.google.com:19302' },
                { urls: 'stun:stun3.l.google.com:19302' },
                { urls: 'stun:stun4.l.google.com:19302' },
                // Добавляем TURN сервер для обхода NAT
                {
                    urls: 'turn:numb.viagenie.ca',
                    username: 'webrtc@live.com',
                    credential: 'muazkh'
                }
            ],
            iceCandidatePoolSize: 10,
            // Принудительно используем relay для обхода NAT
            iceTransportPolicy: 'relay'
        });
        
        currentRTCPeerConnection = pc;
        
        // Подробное логгирование всех событий WebRTC
        pc.addEventListener('negotiationneeded', e => console.log('📢 negotiationneeded', e));
        pc.addEventListener('signalingstatechange', e => console.log('📢 signalingstatechange', pc.signalingState));
        pc.addEventListener('iceconnectionstatechange', e => console.log('📢 iceconnectionstatechange', pc.iceConnectionState));
        pc.addEventListener('icegatheringstatechange', e => console.log('📢 icegatheringstatechange', pc.iceGatheringState));
        pc.addEventListener('icecandidate', e => console.log('📢 icecandidate', e.candidate));
        pc.addEventListener('connectionstatechange', e => console.log('📢 connectionstatechange', pc.connectionState));
        
        // Обработчики WebRTC событий
        pc.ontrack = function(event) {
            console.log('📺 Получен медиа-трек:', event.track.kind);
            if (event.streams && event.streams[0]) {
                console.log('💫 Установка источника видео:', event.streams[0]);
                videoElement.srcObject = event.streams[0];
                currentStream = event.streams[0];
                updateConnectionStatus('online');
            }
        };
        
        pc.oniceconnectionstatechange = function() {
            console.log('🔌 ICE состояние:', pc.iceConnectionState);
            
            if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
                updateConnectionStatus('online');
            } else if (pc.iceConnectionState === 'disconnected' || pc.iceConnectionState === 'failed') {
                updateConnectionStatus('offline');
                console.error('❌ ICE соединение разорвано или не удалось установить');
            }
        };
        
        // Добавляем трансивер для получения видео
        console.log('🔄 Добавляем видео трансивер');
        pc.addTransceiver('video', { direction: 'recvonly' });
        
        // Создаем SDP offer
        console.log('📝 Создаем SDP offer');
        const offer = await pc.createOffer({
            offerToReceiveVideo: true,
            offerToReceiveAudio: false
        });
        
        console.log('📝 SDP offer создан:', offer);
        
        await pc.setLocalDescription(offer);
        console.log('📝 Установлен локальный SDP');
        
        // Отправляем offer на сервер
        console.log('📤 Отправляем offer на сервер');
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
            throw new Error('HTTP ' + response.status);
        }
        
        const answer = await response.json();
        console.log('📥 Получен SDP answer:', answer);
        
        if (answer.error) {
            throw new Error(answer.error);
        }
        
        // Устанавливаем удаленное описание
        if (answer.sdp) {
            console.log('📝 Устанавливаем удаленный SDP');
            await pc.setRemoteDescription(new RTCSessionDescription({
                type: answer.type || 'answer',
                sdp: answer.sdp
            }));
            console.log('📝 Удаленный SDP установлен');
        } else {
            console.error('❌ SDP отсутствует в ответе');
        }
        
        // Добавляем обработчик для проверки состояния соединения через некоторое время
        setTimeout(() => {
            if (pc.iceConnectionState !== 'connected' && pc.iceConnectionState !== 'completed') {
                console.error('⏱️ Время ожидания ICE соединения истекло:', pc.iceConnectionState);
                // Можно добавить автоматическую перезагрузку
            }
        }, 10000); // 10 секунд ожидания
        
        // Очищаем контейнер и добавляем видео
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
            '</div>';
        videoContainer.appendChild(infoPanel);
        
        // Добавляем контролы
        addVideoControls(videoContainer);
        
        // Добавляем кнопку перезагрузки видео
        const reloadButton = document.createElement('button');
        reloadButton.className = 'reload-btn primary-btn';
        reloadButton.style.position = 'absolute';
        reloadButton.style.bottom = '20px';
        reloadButton.style.left = '20px';
        reloadButton.style.zIndex = '100';
        reloadButton.textContent = '🔄 Перезагрузить видео';
        reloadButton.onclick = () => {
            stopCurrentStream();
            startLiveStream();
        };
        videoContainer.appendChild(reloadButton);
        
        // Добавляем индикатор соединения
        const connectionIndicator = document.createElement('div');
        connectionIndicator.className = 'connection-indicator';
        connectionIndicator.style.position = 'absolute';
        connectionIndicator.style.top = '10px';
        connectionIndicator.style.right = '10px';
        connectionIndicator.style.padding = '5px 10px';
        connectionIndicator.style.borderRadius = '4px';
        connectionIndicator.style.fontSize = '12px';
        connectionIndicator.style.color = 'white';
        connectionIndicator.style.background = 'rgba(0, 0, 0, 0.5)';
        connectionIndicator.style.zIndex = '100';
        connectionIndicator.textContent = '🔄 Установка соединения...';
        videoContainer.appendChild(connectionIndicator);
        
        // Обновляем индикатор при изменении состояния
        const updateIndicator = () => {
            if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
                connectionIndicator.textContent = '✅ Соединение установлено';
                connectionIndicator.style.background = 'rgba(46, 204, 113, 0.5)';
            } else if (pc.iceConnectionState === 'checking') {
                connectionIndicator.textContent = '🔄 Проверка соединения...';
                connectionIndicator.style.background = 'rgba(241, 196, 15, 0.5)';
            } else if (pc.iceConnectionState === 'disconnected' || pc.iceConnectionState === 'failed') {
                connectionIndicator.textContent = '❌ Ошибка соединения';
                connectionIndicator.style.background = 'rgba(231, 76, 60, 0.5)';
            }
        };
        
        pc.addEventListener('iceconnectionstatechange', updateIndicator);
        
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
            showError(timeline, 'Ошибка загрузки временной шкалы');
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
        
        // Сортируем записи по времени
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
                '</div>' +
                '<div class="recording-actions">' +
                    '<button class="play-btn primary-btn" onclick="playRecording(\'' + 
                        recording.StartTime + '\', \'' + recording.EndTime + '\', \'' + recording.Channel + '\')">' +
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
            hourLine.style.left = ((hour / 24) * 100) + '%';
            hourLine.style.top = '0';
            hourLine.style.bottom = '0';
            hourLine.style.width = '1px';
            hourLine.style.background = '#ddd';
            timelineContainer.appendChild(hourLine);
            
            // Метки времени каждые 3 часа
            if (hour % 3 === 0) {
                const timeLabel = document.createElement('div');
                timeLabel.className = 'time-label';
                timeLabel.style.position = 'absolute';
                timeLabel.style.left = ((hour / 24) * 100) + '%';
                timeLabel.style.top = '-20px';
                timeLabel.style.fontSize = '11px';
                timeLabel.style.color = '#666';
                timeLabel.style.transform = 'translateX(-50%)';
                timeLabel.textContent = hour.toString().padStart(2, '0') + ':00';
                timelineContainer.appendChild(timeLabel);
            }
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
                segment.style.top = '15px';
                segment.style.background = '#3498db';
                segment.style.cursor = 'pointer';
                segment.style.borderRadius = '3px';
                segment.style.transition = 'all 0.2s';
                
                // Всплывающая подсказка
                const tooltip = formatDateTime(recording.StartTime) + ' - ' + formatDateTime(recording.EndTime);
                segment.title = tooltip;
                
                // Эффект при наведении
                segment.addEventListener('mouseenter', function() {
                    segment.style.background = '#2980b9';
                    segment.style.transform = 'scaleY(1.2)';
                });
                
                segment.addEventListener('mouseleave', function() {
                    segment.style.background = '#3498db';
                    segment.style.transform = 'scaleY(1)';
                });
                
                // Обработчик клика
                segment.onclick = function() {
                    playRecording(recording.StartTime, recording.EndTime, recording.Channel);
                };
                
                timelineContainer.appendChild(segment);
            }
        });
        
        // Индикатор текущего времени
        const now = new Date();
        if (now.toDateString() === dayStart.toDateString()) {
            const currentTimePosition = ((now - dayStart) / dayDuration) * 100;
            const currentTimeIndicator = document.createElement('div');
            currentTimeIndicator.className = 'current-time-indicator';
            currentTimeIndicator.style.position = 'absolute';
            currentTimeIndicator.style.left = currentTimePosition + '%';
            currentTimeIndicator.style.top = '0';
            currentTimeIndicator.style.bottom = '0';
            currentTimeIndicator.style.width = '2px';
            currentTimeIndicator.style.background = '#e74c3c';
            currentTimeIndicator.style.zIndex = '10';
            timelineContainer.appendChild(currentTimeIndicator);
        }
        
        timeline.appendChild(timelineContainer);
    }
    
    /**
     * Воспроизведение архивной записи
     */
    window.playRecording = async function(startTime, endTime, channelId) {
        showLoading('Загрузка архивной записи...');
        
        // Пока показываем RTSP URL, так как WebRTC для архива требует дополнительной настройки
        try {
            const response = await fetch('/api/playback-url?channel=' + channelId + 
                '&start=' + startTime + '&end=' + endTime);
            
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
                        '<p><em>💡 Используйте VLC Player для воспроизведения</em></p>' +
                        '<button onclick="copyToClipboard(\'' + data.url + '\')" class="primary-btn" style="margin-top: 10px;">' +
                            '📋 Копировать URL' +
                        '</button>' +
                        '<button onclick="location.reload()" class="secondary-btn" style="margin-top: 10px; margin-left: 10px;">' +
                            '🔙 Вернуться' +
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
            alert('Не удалось скопировать URL');
        }
        document.body.removeChild(textArea);
    }
    
    /**
     * Получение снимка с камеры
     */
    async function takeSnapshot(channelId) {
        if (!channelId) {
            channelId = cameraSelect.value;
            if (!channelId) {
                alert('Выберите канал для создания снимка');
                return;
            }
        }
        
        try {
            showLoading('Создание снимка...');
            
            const response = await fetch('/api/snapshot/' + channelId);
            if (!response.ok) {
                throw new Error('HTTP ' + response.status);
            }
            
            const blob = await response.blob();
            const imageUrl = URL.createObjectURL(blob);
            
            const link = document.createElement('a');
            link.href = imageUrl;
            link.download = 'snapshot_' + channelId + '_' + new Date().toISOString().replace(/[:.]/g, '-') + '.jpg';
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            
            URL.revokeObjectURL(imageUrl);
            
            hideLoading();
            
        } catch (error) {
            console.error('Ошибка создания снимка:', error);
            alert('Не удалось создать снимок: ' + error.message);
            hideLoading();
        }
    }
    
    /**
     * Обработчик изменения полноэкранного режима
     */
    document.addEventListener('fullscreenchange', function() {
        isFullscreen = !!document.fullscreenElement;
    });
    
    document.addEventListener('webkitfullscreenchange', function() {
        isFullscreen = !!document.webkitFullscreenElement;
    });
    
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
            snapshotBtn.addEventListener('click', () => takeSnapshot());
        }
        
        // Кнопка поиска записей
        if (searchBtn) {
            searchBtn.addEventListener('click', searchRecordings);
        }
        
        // Enter в поле даты
        if (archiveDate) {
            archiveDate.addEventListener('keypress', function(e) {
                if (e.key === 'Enter') {
                    searchRecordings();
                }
            });
        }
        
        // Обработка закрытия страницы
        window.addEventListener('beforeunload', function() {
            stopCurrentStream();
        });
        
        // Клавиатурные сокращения
        document.addEventListener('keydown', function(e) {
            // F - полноэкранный режим
            if (e.key === 'f' || e.key === 'F') {
                if (currentVideoElement && !e.target.matches('input, textarea')) {
                    toggleFullscreen();
                }
            }
            // Escape - выход из полноэкранного режима
            if (e.key === 'Escape' && isFullscreen) {
                toggleFullscreen();
            }
            // Space - пауза/воспроизведение
            if (e.key === ' ' && currentVideoElement && !e.target.matches('input, textarea')) {
                e.preventDefault();
                if (currentVideoElement.paused) {
                    currentVideoElement.play();
                } else {
                    currentVideoElement.pause();
                }
            }
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
            } else {
                updateConnectionStatus('offline');
            }
        } catch (error) {
            updateConnectionStatus('offline');
        }
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