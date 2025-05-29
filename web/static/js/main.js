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
 * Запуск WebRTC потока с поддержкой мобильных сетей
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
        
        // Определяем, мобильное ли соединение
        const isMobileConnection = navigator.connection && 
            (navigator.connection.type === 'cellular' || 
             navigator.connection.effectiveType === '3g' || 
             navigator.connection.effectiveType === '2g' ||
             navigator.connection.downlink < 2);
        
        console.log(`📱 Тип соединения: ${isMobileConnection ? 'Мобильное' : 'Стационарное'}`);
        
        // Специальные настройки WebRTC для мобильных сетей
        const pc = new RTCPeerConnection({
            iceServers: [
                // STUN серверы, работающие через мобильные сети
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                { urls: 'stun:stun.stunprotocol.org:3478' },
                
                // TURN серверы с TCP - обязательно для мобильных сетей
                {
                    urls: [
                        'turn:openrelay.metered.ca:80?transport=tcp',
                        'turn:openrelay.metered.ca:443?transport=tcp'
                    ],
                    username: 'openrelayproject',
                    credential: 'openrelayproject'
                },
                {
                    urls: [
                        'turn:openrelay.metered.ca:80',
                        'turn:openrelay.metered.ca:443'
                    ],
                    username: 'openrelayproject',
                    credential: 'openrelayproject'
                },
                {
                    urls: 'turn:global.turn.twilio.com:3478?transport=tcp',
                    username: 'f4b4035eaa76f77e3b3f90eda80c6f9250e6072728a458fc1345000000000000',
                    credential: '/7OMV8e64wEVj/8H64d4vGKWG9Dj3kZTTI4rjZhzDvk='
                },
                {
                    urls: [
                        'turn:eu-turn1.xirsys.com:80?transport=tcp',
                        'turn:eu-turn2.xirsys.com:80?transport=tcp'
                    ],
                    username: '4kFU+JWgwVXiQ3Bf9pRLvRNlOCk=',
                    credential: '3+2+4WN9VxbdYpU9gcQuTmrlIc0='
                },
                {
                    urls: 'turn:turn.anyfirewall.com:443?transport=tcp',
                    username: 'webrtc',
                    credential: 'webrtc'
                }
            ],
            iceCandidatePoolSize: 20,
            // Принудительно используем только relay для мобильных сетей
            iceTransportPolicy: isMobileConnection ? 'relay' : 'all',
            bundlePolicy: 'max-bundle',
            rtcpMuxPolicy: 'require',
            sdpSemantics: 'unified-plan'
        });
        
        currentRTCPeerConnection = pc;
        
        // Добавляем все обработчики событий WebRTC
        setupWebRTCEventHandlers(pc, videoElement);
        
        // Добавляем трансивер для получения видео
        console.log('🔄 Добавляем видео трансивер');
        pc.addTransceiver('video', { 
            direction: 'recvonly',
            streams: [new MediaStream()]
        });
        
        // Создаем SDP offer с улучшенными настройками для мобильных сетей
        console.log('📝 Создаем SDP offer');
        const offerOptions = {
            offerToReceiveVideo: true,
            offerToReceiveAudio: false,
            voiceActivityDetection: false,
            iceRestart: true
        };
        
        const offer = await pc.createOffer(offerOptions);
        
        // Модифицируем SDP для улучшения работы через мобильные сети
        let sdp = offer.sdp;
        
        // Установка параметров видео
        if (isMobileConnection) {
            // Понижаем битрейт для мобильных сетей
            sdp = sdp.replace(/a=mid:video\r\n/g, 
                'a=mid:video\r\na=content:main\r\na=rtcp-fb:* nack\r\na=rtcp-fb:* nack pli\r\na=rtcp-fb:* ccm fir\r\n');
            
            // Устанавливаем параметры видео кодека для мобильных сетей
            sdp = sdp.replace(/a=rtpmap:(96|97|98) VP8\/90000\r\n/g, 
                'a=rtpmap:$1 VP8/90000\r\na=fmtp:$1 x-google-min-bitrate=100;x-google-max-bitrate=800;x-google-start-bitrate=300\r\n');
        }
        
        // Принудительно используем TCP для мобильных сетей
        if (isMobileConnection) {
            // Удаляем UDP кандидатов если мобильное соединение
            sdp = sdp.replace(/a=candidate:.*UDP.*\r\n/g, '');
        }
        
        offer.sdp = sdp;
        console.log('📝 SDP offer создан и модифицирован');
        
        await pc.setLocalDescription(offer);
        console.log('📝 Установлен локальный SDP');
        
        // Показываем индикатор загрузки
        showLoadingIndicator(videoContainer, 'Подключение к камере...');
        
        // Отправляем offer на сервер
        console.log('📤 Отправляем offer на сервер');
        try {
            const response = await fetch('/api/webrtc/offer?channel=' + channelId, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    type: offer.type,
                    sdp: offer.sdp
                }),
                // Увеличиваем таймаут для мобильных соединений
                timeout: 15000
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
                
                // Модифицируем SDP answer для лучшей работы с мобильными сетями
                let remoteSdp = answer.sdp;
                
                // Для мобильных сетей принудительно используем TCP
                if (isMobileConnection) {
                    remoteSdp = remoteSdp.replace(/a=candidate:.*UDP.*\r\n/g, '');
                }
                
                await pc.setRemoteDescription(new RTCSessionDescription({
                    type: answer.type || 'answer',
                    sdp: remoteSdp
                }));
                console.log('📝 Удаленный SDP установлен');
                
                // Перебираем активные трансиверы и проверяем состояние
                pc.getTransceivers().forEach(transceiver => {
                    console.log(`📡 Трансивер: ${transceiver.mid}, Направление: ${transceiver.direction}, Текущее направление: ${transceiver.currentDirection}`);
                });
                
                // Таймаут для проверки установки соединения
                setTimeout(() => {
                    if (pc.iceConnectionState !== 'connected' && pc.iceConnectionState !== 'completed') {
                        console.warn('⚠️ Медленное соединение, проверяем статус:', pc.iceConnectionState);
                        
                        // Если соединение в процессе - показываем сообщение
                        if (pc.iceConnectionState === 'checking') {
                            showLoadingIndicator(videoContainer, 'Соединение устанавливается... Ждите');
                        }
                        // Если проблема с соединением - пытаемся переподключиться
                        else if (pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'disconnected') {
                            console.error('⏱️ Таймаут соединения: ', pc.iceConnectionState);
                            // Для мобильных сетей увеличиваем тайминги
                            if (isMobileConnection) {
                                showLoadingIndicator(videoContainer, 'Мобильное соединение медленное. Подождите...');
                                setTimeout(() => {
                                    if (pc.iceConnectionState !== 'connected' && pc.iceConnectionState !== 'completed') {
                                        tryReconnect(channelId);
                                    }
                                }, 10000); // Увеличиваем задержку для мобильных сетей
                            } else {
                                tryReconnect(channelId);
                            }
                        }
                    }
                }, isMobileConnection ? 15000 : 10000); // Увеличиваем время ожидания для мобильных сетей
            } else {
                console.error('❌ SDP отсутствует в ответе');
                throw new Error('SDP отсутствует в ответе');
            }
        } catch (error) {
            console.error('❌ Ошибка при отправке/получении SDP:', error);
            throw error;
        }
        
        // Очищаем контейнер и добавляем видео
        videoContainer.innerHTML = '';
        videoContainer.appendChild(videoElement);
        currentVideoElement = videoElement;
        
        // Добавляем информационную панель и другие элементы интерфейса
        addVideoUIElements(videoContainer, channelId, streamData, isMobileConnection);
        
    } catch (error) {
        console.error('❌ WebRTC ошибка:', error);
        showErrorOverlay(videoContainer, 'WebRTC ошибка: ' + error.message);
        throw new Error('WebRTC ошибка: ' + error.message);
    }
}
   /**
 * Остановка текущего потока
 */
function stopCurrentStream() {
    // Остановка WebRTC соединения
    if (currentRTCPeerConnection) {
        try {
            // Закрываем WebRTC соединение
            currentRTCPeerConnection.getSenders().forEach(sender => {
                if (sender.track) {
                    sender.track.stop();
                }
            });
            
            currentRTCPeerConnection.getReceivers().forEach(receiver => {
                if (receiver.track) {
                    receiver.track.stop();
                }
            });
            
            currentRTCPeerConnection.close();
        } catch (e) {
            console.error('Ошибка при закрытии WebRTC соединения:', e);
        }
        
        currentRTCPeerConnection = null;
    }
    
    // Остановка медиа-потоков
    if (currentStream) {
        try {
            currentStream.getTracks().forEach(function(track) {
                track.stop();
            });
        } catch (e) {
            console.error('Ошибка при остановке медиа-потоков:', e);
        }
        
        currentStream = null;
    }
    
    // Очистка видео-элемента
    if (currentVideoElement) {
        try {
            if (currentVideoElement.srcObject) {
                currentVideoElement.srcObject.getTracks().forEach(track => track.stop());
                currentVideoElement.srcObject = null;
            }
            
            if (currentVideoElement.parentNode) {
                currentVideoElement.parentNode.removeChild(currentVideoElement);
            }
        } catch (e) {
            console.error('Ошибка при очистке видео-элемента:', e);
        }
        
        currentVideoElement = null;
    }
    
    // Обновление статуса соединения
    updateConnectionStatus('offline');
    
    console.log('🛑 Текущий поток остановлен');
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
    
    // Определяем тип соединения
    handleMobileConnection();
    
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
    /**
 * Показывает индикатор загрузки внутри указанного контейнера
 */
function showLoadingIndicator(container, message) {
    // Проверяем, существует ли уже индикатор
    let loadingIndicator = container.querySelector('.video-loading-indicator');
    
    if (!loadingIndicator) {
        loadingIndicator = document.createElement('div');
        loadingIndicator.className = 'video-loading-indicator';
        loadingIndicator.style.position = 'absolute';
        loadingIndicator.style.top = '50%';
        loadingIndicator.style.left = '50%';
        loadingIndicator.style.transform = 'translate(-50%, -50%)';
        loadingIndicator.style.background = 'rgba(0, 0, 0, 0.7)';
        loadingIndicator.style.color = 'white';
        loadingIndicator.style.padding = '15px 20px';
        loadingIndicator.style.borderRadius = '5px';
        loadingIndicator.style.textAlign = 'center';
        loadingIndicator.style.zIndex = '1000';
        
        const spinner = document.createElement('div');
        spinner.style.width = '30px';
        spinner.style.height = '30px';
        spinner.style.borderRadius = '50%';
        spinner.style.border = '3px solid rgba(255, 255, 255, 0.3)';
        spinner.style.borderTopColor = '#fff';
        spinner.style.margin = '0 auto 10px';
        spinner.style.animation = 'spin 1s linear infinite';
        
        if (!document.querySelector('style#loading-animation')) {
            const style = document.createElement('style');
            style.id = 'loading-animation';
            style.textContent = `
                @keyframes spin { 
                    to { transform: rotate(360deg); }
                }
            `;
            document.head.appendChild(style);
        }
        
        loadingIndicator.appendChild(spinner);
        
        const text = document.createElement('div');
        text.textContent = message || 'Загрузка...';
        loadingIndicator.appendChild(text);
        
        container.appendChild(loadingIndicator);
    } else {
        // Обновляем сообщение, если индикатор уже существует
        const textElement = loadingIndicator.querySelector('div:last-child');
        if (textElement) {
            textElement.textContent = message || 'Загрузка...';
        }
        
        // Показываем существующий индикатор
        loadingIndicator.style.display = 'block';
    }
}

/**
 * Скрывает индикатор загрузки
 */
function hideLoadingIndicator(container) {
    const loadingIndicator = container.querySelector('.video-loading-indicator');
    if (loadingIndicator) {
        loadingIndicator.style.display = 'none';
    }
}

/**
 * Показывает индикатор стриминга
 */
function showStreamingIndicator(container) {
    // Создаем индикатор потока
    let streamingIndicator = container.querySelector('.streaming-indicator');
    
    if (!streamingIndicator) {
        streamingIndicator = document.createElement('div');
        streamingIndicator.className = 'streaming-indicator';
        streamingIndicator.style.position = 'absolute';
        streamingIndicator.style.bottom = '60px';
        streamingIndicator.style.right = '20px';
        streamingIndicator.style.background = 'rgba(46, 204, 113, 0.7)';
        streamingIndicator.style.color = 'white';
        streamingIndicator.style.padding = '5px 10px';
        streamingIndicator.style.borderRadius = '4px';
        streamingIndicator.style.fontSize = '12px';
        streamingIndicator.style.zIndex = '100';
        streamingIndicator.textContent = '🔴 LIVE';
        
        // Мигающий эффект
        if (!document.querySelector('style#streaming-animation')) {
            const style = document.createElement('style');
            style.id = 'streaming-animation';
            style.textContent = `
                @keyframes pulse { 
                    0% { opacity: 1; }
                    50% { opacity: 0.7; }
                    100% { opacity: 1; }
                }
            `;
            document.head.appendChild(style);
        }
        
        streamingIndicator.style.animation = 'pulse 2s infinite';
        
        container.appendChild(streamingIndicator);
    } else {
        streamingIndicator.style.display = 'block';
    }
}

/**
 * Показывает сообщение об ошибке
 */
function showErrorOverlay(container, message) {
    // Проверяем, существует ли уже оверлей
    let errorOverlay = container.querySelector('.error-overlay');
    
    if (!errorOverlay) {
        errorOverlay = document.createElement('div');
        errorOverlay.className = 'error-overlay';
        errorOverlay.style.position = 'absolute';
        errorOverlay.style.top = '0';
        errorOverlay.style.left = '0';
        errorOverlay.style.width = '100%';
        errorOverlay.style.height = '100%';
        errorOverlay.style.background = 'rgba(231, 76, 60, 0.8)';
        errorOverlay.style.color = 'white';
        errorOverlay.style.display = 'flex';
        errorOverlay.style.flexDirection = 'column';
        errorOverlay.style.alignItems = 'center';
        errorOverlay.style.justifyContent = 'center';
        errorOverlay.style.textAlign = 'center';
        errorOverlay.style.zIndex = '1000';
        errorOverlay.style.padding = '20px';
        
        const icon = document.createElement('div');
        icon.textContent = '❌';
        icon.style.fontSize = '48px';
        icon.style.marginBottom = '10px';
        
        const text = document.createElement('div');
        text.textContent = message || 'Произошла ошибка';
        text.style.fontSize = '16px';
        
        const retryButton = document.createElement('button');
        retryButton.textContent = '🔄 Повторить';
        retryButton.style.marginTop = '20px';
        retryButton.style.padding = '10px 20px';
        retryButton.style.borderRadius = '5px';
        retryButton.style.border = 'none';
        retryButton.style.background = 'white';
        retryButton.style.color = '#e74c3c';
        retryButton.style.cursor = 'pointer';
        retryButton.style.fontWeight = 'bold';
        
        retryButton.onclick = () => {
            errorOverlay.style.display = 'none';
            const channelId = cameraSelect.value;
            if (channelId) {
                stopCurrentStream();
                startLiveStream();
            }
        };
        
        errorOverlay.appendChild(icon);
        errorOverlay.appendChild(text);
        errorOverlay.appendChild(retryButton);
        
        container.appendChild(errorOverlay);
    } else {
        // Обновляем сообщение, если оверлей уже существует
        const textElement = errorOverlay.querySelector('div:nth-child(2)');
        if (textElement) {
            textElement.textContent = message || 'Произошла ошибка';
        }
        
        // Показываем существующий оверлей
        errorOverlay.style.display = 'flex';
    }
}

/**
 * Пытается перезапустить ICE соединение
 */
async function tryRestartIce(pc) {
    try {
        if (pc.restartIce) {
            pc.restartIce();
            console.log('🔄 ICE перезапущен');
        } else {
            console.log('❌ Функция restartIce не поддерживается');
            
            // Если restartIce не поддерживается, создаем новый offer с iceRestart
            const offer = await pc.createOffer({ iceRestart: true });
            await pc.setLocalDescription(offer);
            console.log('🔄 ICE перезапущен через новый offer');
        }
    } catch (error) {
        console.error('❌ Ошибка перезапуска ICE:', error);
    }
}

/**
 * Повторное подключение к камере
 */
function tryReconnect(channelId) {
    console.log('🔄 Попытка переподключения к камере:', channelId);
    
    if (!channelId) {
        channelId = cameraSelect.value;
        if (!channelId) {
            console.error('❌ ID канала не указан для переподключения');
            return;
        }
    }
    
    // Останавливаем текущий поток
    stopCurrentStream();
    
    // Пытаемся подключиться заново
    showLoadingIndicator(videoContainer, 'Переподключение...');
    
    // Задержка перед переподключением
    setTimeout(() => {
        startLiveStream();
    }, 1000);
}

/**
 * Настройка обработчиков событий WebRTC
 */
function setupWebRTCEventHandlers(pc, videoElement) {
    // Логгирование событий WebRTC
    pc.addEventListener('negotiationneeded', e => console.log('📢 negotiationneeded'));
    pc.addEventListener('signalingstatechange', e => console.log('📢 signalingstatechange:', pc.signalingState));
    pc.addEventListener('iceconnectionstatechange', e => {
        console.log('📢 iceconnectionstatechange:', pc.iceConnectionState);
        
        // Перезапуск при ошибке
        if (pc.iceConnectionState === 'failed') {
            console.log('🔄 Пытаемся перезапустить ICE соединение');
            tryRestartIce(pc);
        }
    });
    pc.addEventListener('icegatheringstatechange', e => console.log('📢 icegatheringstatechange:', pc.iceGatheringState));
    pc.addEventListener('connectionstatechange', e => {
        console.log('📢 connectionstatechange:', pc.connectionState);
        
        // Перезапуск при ошибке
        if (pc.connectionState === 'failed') {
            console.log('🔄 Пытаемся перезапустить соединение');
            const channelId = cameraSelect.value;
            if (channelId) {
                tryReconnect(channelId);
            }
        }
    });
    
    // Логируем ICE кандидатов
    pc.addEventListener('icecandidate', e => {
        if (e.candidate) {
            console.log('🧊 ICE candidate:', e.candidate.type, e.candidate.protocol, e.candidate.address);
            
            // Приоритизация relay кандидатов для мобильных сетей
            if (e.candidate.type === 'relay') {
                console.log('🔼 Повышаем приоритет relay кандидата');
                // Можно модифицировать приоритет, но это зависит от браузера
            }
        } else {
            console.log('🧊 ICE gathering complete');
        }
    });
    
    // Обработчики WebRTC событий
    pc.ontrack = function(event) {
        console.log('📺 Получен медиа-трек:', event.track.kind);
        if (event.streams && event.streams[0]) {
            console.log('💫 Установка источника видео');
            videoElement.srcObject = event.streams[0];
            currentStream = event.streams[0];
            
            // Обработка метаданных видео
            videoElement.onloadedmetadata = () => {
                console.log('📐 Видео размер:', videoElement.videoWidth + 'x' + videoElement.videoHeight);
                updateConnectionStatus('online');
                
                // Показываем индикатор стриминга
                showStreamingIndicator(videoContainer);
            };
            
            // Событие воспроизведения
            videoElement.onplaying = () => {
                console.log('▶️ Видео воспроизводится');
                hideLoadingIndicator(videoContainer);
                updateConnectionStatus('online');
            };
            
            // Обрабатываем события паузы и буферизации
            videoElement.onwaiting = () => {
                console.log('⏸️ Видео буферизуется');
                showLoadingIndicator(videoContainer, 'Буферизация...');
            };
            
            videoElement.onstalled = () => {
                console.log('⚠️ Видео приостановлено');
                showLoadingIndicator(videoContainer, 'Соединение прервано. Восстановление...');
            };
        }
    };
    
    // Обработка ошибок видео
    videoElement.onerror = function(error) {
        console.error('❌ Ошибка видео:', error);
        showErrorOverlay(videoContainer, 'Ошибка воспроизведения видео');
    };
    
    // Отслеживаем статус соединения
    pc.oniceconnectionstatechange = function() {
        console.log('🔌 ICE состояние:', pc.iceConnectionState);
        
        if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
            updateConnectionStatus('online');
            hideLoadingIndicator(videoContainer);
        } else if (pc.iceConnectionState === 'disconnected') {
            showLoadingIndicator(videoContainer, 'Переподключение...');
            updateConnectionStatus('connecting');
        } else if (pc.iceConnectionState === 'failed') {
            updateConnectionStatus('offline');
            showErrorOverlay(videoContainer, 'Не удалось установить соединение. Пробуем переподключиться...');
            console.error('❌ ICE соединение не удалось установить');
            
            // Автоматическая попытка переподключения через 5 секунд
            setTimeout(() => {
                const channelId = cameraSelect.value;
                if (channelId) {
                    tryReconnect(channelId);
                }
            }, 5000);
        }
    };
}

/**
 * Добавляет элементы интерфейса для видео
 */
function addVideoUIElements(container, channelId, streamData, isMobile) {
    // Добавляем информационную панель
    const infoPanel = document.createElement('div');
    infoPanel.className = 'video-info-panel';
    infoPanel.innerHTML = 
        '<div class="video-info">' +
            '<span>📺 ' + (streamData.channel_name || 'Канал ' + channelId) + '</span>' +
            '<span>🔴 Прямой эфир' + (isMobile ? ' (Мобильное соединение)' : '') + '</span>' +
        '</div>';
    container.appendChild(infoPanel);
    
    // Добавляем контролы
    addVideoControls(container);
    
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
    container.appendChild(reloadButton);
    
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
    container.appendChild(connectionIndicator);
    
    // Добавляем для мобильных специальные элементы
    if (isMobile) {
        const mobileInfo = document.createElement('div');
        mobileInfo.className = 'mobile-info';
        mobileInfo.style.position = 'absolute';
        mobileInfo.style.top = '40px';
        mobileInfo.style.right = '10px';
        mobileInfo.style.padding = '5px 10px';
        mobileInfo.style.borderRadius = '4px';
        mobileInfo.style.fontSize = '12px';
        mobileInfo.style.color = 'white';
        mobileInfo.style.background = 'rgba(255, 165, 0, 0.7)';
        mobileInfo.style.zIndex = '100';
        mobileInfo.textContent = '📱 Мобильное соединение';
        container.appendChild(mobileInfo);
        
        // Уменьшаем качество для мобильных
        const qualitySwitch = document.createElement('button');
        qualitySwitch.className = 'quality-switch';
        qualitySwitch.style.position = 'absolute';
        qualitySwitch.style.top = '70px';
        qualitySwitch.style.right = '10px';
        qualitySwitch.style.padding = '5px 10px';
        qualitySwitch.style.borderRadius = '4px';
        qualitySwitch.style.fontSize = '12px';
        qualitySwitch.style.color = 'white';
        qualitySwitch.style.background = 'rgba(52, 152, 219, 0.7)';
        qualitySwitch.style.border = 'none';
        qualitySwitch.style.cursor = 'pointer';
        qualitySwitch.style.zIndex = '100';
        qualitySwitch.textContent = '🔍 Низкое качество';
        qualitySwitch.onclick = () => {
            // Изменение ID канала для переключения качества
            // Например, с HD (x01) на SD (x02)
            let newChannelId = channelId;
            if (channelId.endsWith('01')) {
                newChannelId = channelId.slice(0, -2) + '02';
            } else if (channelId.endsWith('02')) {
                newChannelId = channelId.slice(0, -2) + '01';
            }
            
            if (newChannelId !== channelId) {
                cameraSelect.value = newChannelId;
                stopCurrentStream();
                startLiveStream();
            }
        };
        container.appendChild(qualitySwitch);
    }
    
    // Обновляем индикатор при изменении состояния
    if (currentRTCPeerConnection) {
        const updateIndicator = () => {
            if (currentRTCPeerConnection.iceConnectionState === 'connected' || 
                currentRTCPeerConnection.iceConnectionState === 'completed') {
                connectionIndicator.textContent = '✅ Соединение установлено';
                connectionIndicator.style.background = 'rgba(46, 204, 113, 0.5)';
            } else if (currentRTCPeerConnection.iceConnectionState === 'checking') {
                connectionIndicator.textContent = '🔄 Проверка соединения...';
                connectionIndicator.style.background = 'rgba(241, 196, 15, 0.5)';
            } else if (currentRTCPeerConnection.iceConnectionState === 'disconnected') {
                connectionIndicator.textContent = '⚠️ Соединение нестабильно';
                connectionIndicator.style.background = 'rgba(230, 126, 34, 0.5)';
            } else if (currentRTCPeerConnection.iceConnectionState === 'failed') {
                connectionIndicator.textContent = '❌ Ошибка соединения';
                connectionIndicator.style.background = 'rgba(231, 76, 60, 0.5)';
            }
        };
        
        currentRTCPeerConnection.addEventListener('iceconnectionstatechange', updateIndicator);
    }
}

//**

function detectConnectionType() {
   if (navigator.connection) {
       // Получаем информацию о соединении
       const connection = navigator.connection;
       
       console.log('Тип соединения:', connection.type);
       console.log('Эффективный тип:', connection.effectiveType);
       console.log('Пропускная способность:', connection.downlink, 'Мбит/с');
       console.log('RTT:', connection.rtt, 'мс');
       
       // Определяем, мобильное ли соединение
       const isMobile = 
           connection.type === 'cellular' || 
           connection.effectiveType === '3g' || 
           connection.effectiveType === '2g' ||
           connection.downlink < 2;
       
       return {
           type: connection.type,
           effectiveType: connection.effectiveType,
           downlink: connection.downlink,
           rtt: connection.rtt,
           isMobile: isMobile
       };
   }
   
   // Если API недоступен, проверяем через User-Agent
   const userAgent = navigator.userAgent;
   const mobileRegex = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i;
   const isMobile = mobileRegex.test(userAgent);
   
   return {
       type: isMobile ? 'probably cellular' : 'probably wifi',
       effectiveType: 'unknown',
       downlink: -1,
       rtt: -1,
       isMobile: isMobile
   };
}

/**
* Обработка мобильного соединения
*/
function handleMobileConnection() {
   // Добавляем обработчик изменения соединения, если API доступен
   if (navigator.connection) {
       navigator.connection.addEventListener('change', function() {
           console.log('🔄 Соединение изменилось, перепроверяем...');
           const connectionInfo = detectConnectionType();
           
           // Если изменился тип соединения, перезагружаем активный поток
           const channelId = cameraSelect.value;
           if (channelId && currentRTCPeerConnection) {
               console.log('🔄 Перезагружаем поток из-за изменения соединения');
               stopCurrentStream();
               startLiveStream();
           }
       });
   }
   
   // Определяем тип соединения при загрузке
   const connectionInfo = detectConnectionType();
   console.log('📱 Информация о соединении:', connectionInfo);
   
   // Добавляем индикатор типа соединения в интерфейс
   const connectionTypeElement = document.createElement('div');
   connectionTypeElement.style.position = 'fixed';
   connectionTypeElement.style.bottom = '10px';
   connectionTypeElement.style.right = '10px';
   connectionTypeElement.style.padding = '5px 10px';
   connectionTypeElement.style.borderRadius = '4px';
   connectionTypeElement.style.fontSize = '12px';
   connectionTypeElement.style.color = 'white';
   connectionTypeElement.style.background = connectionInfo.isMobile ? 
       'rgba(231, 76, 60, 0.7)' : 'rgba(46, 204, 113, 0.7)';
   connectionTypeElement.style.zIndex = '9999';
   connectionTypeElement.textContent = connectionInfo.isMobile ? 
       '📱 Мобильное соединение' : '🖥️ WiFi соединение';
   
   document.body.appendChild(connectionTypeElement);
}

});