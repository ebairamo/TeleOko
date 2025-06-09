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
 * Запуск WebRTC потока - с поддержкой прямого подключения к go2rtc
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
        
        // Показываем индикатор загрузки
        showLoadingIndicator(videoContainer, 'Подключение к камере...');
        
        // Определяем тип соединения
        const isMobile = detectMobileNetwork();
        console.log(`📱 Тип соединения: ${isMobile ? 'Мобильное' : 'WiFi/Проводное'}`);
        
        console.log(`📹 Начинаем подключение к камере ${channelId}`);
        
        // Настройки WebRTC с учетом типа соединения
        const pcConfig = {
            iceServers: [
                // STUN серверы
                { urls: 'stun:stun.l.google.com:19302' },
                { urls: 'stun:stun1.l.google.com:19302' },
                
                // TURN серверы для надежности
                {
                    urls: [
                        'turn:openrelay.metered.ca:80?transport=tcp',
                        'turn:openrelay.metered.ca:443?transport=tcp',
                        'turn:openrelay.metered.ca:80',
                        'turn:openrelay.metered.ca:443'
                    ],
                    username: 'openrelayproject',
                    credential: 'openrelayproject'
                }
            ],
            iceCandidatePoolSize: 10,
            // Для мобильных сетей только relay
            iceTransportPolicy: isMobile ? 'relay' : 'all',
            bundlePolicy: 'max-bundle',
            rtcpMuxPolicy: 'require'
        };
        
        // Создаем RTCPeerConnection
        const pc = new RTCPeerConnection(pcConfig);
        currentRTCPeerConnection = pc;
        
        // Базовые обработчики событий
        setupWebRTCEventHandlers(pc, videoElement);
        
        // Добавляем трансивер для получения видео
        pc.addTransceiver('video', { direction: 'recvonly' });
        
        // Создаем SDP offer
        console.log('📝 Создаем SDP offer');
        const offer = await pc.createOffer({
            offerToReceiveVideo: true,
            offerToReceiveAudio: false,
            iceRestart: true
        });
        
        // Модифицируем SDP для мобильных сетей при необходимости
        if (isMobile) {
            let sdp = offer.sdp;
            // Понижаем битрейт для мобильных сетей
            sdp = sdp.replace(/a=rtpmap:(96|97|98) VP8\/90000\r\n/g, 
                'a=rtpmap:$1 VP8/90000\r\na=fmtp:$1 x-google-min-bitrate=100;x-google-max-bitrate=500;x-google-start-bitrate=300\r\n');
            offer.sdp = sdp;
        }
        
        await pc.setLocalDescription(offer);
        console.log('📝 Установлен локальный SDP');
        
        // Пытаемся получить URL для прямого подключения к go2rtc
        let go2rtcURL = null;
        try {
            // Запрашиваем системную информацию
            const infoResponse = await fetch('/api/info');
            if (infoResponse.ok) {
                const infoData = await infoResponse.json();
                go2rtcURL = infoData.go2rtc_url;
                console.log('🌐 Получен URL для go2rtc:', go2rtcURL);
            }
        } catch (error) {
            console.warn('⚠️ Не удалось получить URL для go2rtc:', error);
        }
        
        let useDirectConnection = false;
        let response = null;
        
        // Пробуем прямое подключение к go2rtc, если URL доступен
        if (go2rtcURL) {
            try {
                console.log('🔄 Пробуем прямое подключение к go2rtc');
                const directUrl = `${go2rtcURL}/api/webrtc?src=${channelId}`;
                console.log('URL для прямого подключения:', directUrl);
                
                response = await fetch(directUrl, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        type: offer.type,
                        sdp: offer.sdp
                    })
                });
                
                if (response.ok) {
                    useDirectConnection = true;
                    console.log('✅ Прямое подключение к go2rtc успешно');
                } else {
                    console.warn('⚠️ Ошибка прямого подключения, статус:', response.status);
                }
            } catch (error) {
                console.warn('⚠️ Ошибка прямого подключения:', error);
            }
        }
        
        // Если прямое подключение не удалось, используем прокси через наш сервер
        if (!useDirectConnection) {
            try {
                console.log('🔄 Используем подключение через прокси');
                response = await fetch('/api/webrtc/offer?channel=' + channelId, {
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
                    throw new Error('Ошибка HTTP: ' + response.status);
                }
            } catch (error) {
                console.error('❌ Ошибка при подключении через прокси:', error);
                showErrorOverlay(videoContainer, 'Не удалось подключиться к серверу');
                return;
            }
        }
        
        // Обрабатываем ответ (одинаково для обоих методов)
        try {
            const answer = await response.json();
            
            if (answer.error) {
                throw new Error(answer.error);
            }
            
            if (!answer.sdp) {
                throw new Error('SDP отсутствует в ответе');
            }
            
            console.log('📝 Устанавливаем удаленный SDP');
            await pc.setRemoteDescription(new RTCSessionDescription(answer));
            console.log('✅ Удаленный SDP установлен');
            
        } catch (error) {
            console.error('❌ Ошибка при обработке SDP ответа:', error);
            showErrorOverlay(videoContainer, 'Ошибка соединения: ' + error.message);
            return;
        }
        
        // Очищаем контейнер и добавляем видео
        videoContainer.innerHTML = '';
        videoContainer.appendChild(videoElement);
        currentVideoElement = videoElement;
        
        // Для мобильных устройств добавляем обработчик клика для автозапуска видео
        videoElement.addEventListener('click', () => {
            if (videoElement.paused) {
                videoElement.play().catch(e => console.error('Ошибка при воспроизведении:', e));
            }
        });
        
        // Добавляем информационную панель с типом соединения
        const infoPanel = document.createElement('div');
        infoPanel.className = 'video-info-panel';
        infoPanel.innerHTML = 
            '<div class="video-info">' +
                '<span>📺 ' + (streamData.channel_name || 'Канал ' + channelId) + '</span>' +
                '<span>🔴 Прямой эфир' + (useDirectConnection ? ' (Прямое соединение)' : '') + '</span>' +
                '<span>' + (isMobile ? '📱 Мобильное соединение' : '🖥️ WiFi/Проводное соединение') + '</span>' +
            '</div>';
        videoContainer.appendChild(infoPanel);
        
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
        
        // Кнопка переключения качества (если это не общий канал)
        if (channelId !== "1") {
            // Определяем текущее качество
            const isHighQuality = channelId.endsWith('01');
            const qualityButton = document.createElement('button');
            qualityButton.style.position = 'absolute';
           qualityButton.style.bottom = '20px';
           qualityButton.style.right = '20px';
           qualityButton.style.padding = '8px 16px';
           qualityButton.style.borderRadius = '4px';
           qualityButton.style.fontSize = '12px';
           qualityButton.style.color = 'white';
           qualityButton.style.background = 'rgba(52, 152, 219, 0.7)';
           qualityButton.style.border = 'none';
           qualityButton.style.cursor = 'pointer';
           qualityButton.style.zIndex = '100';
           
           qualityButton.textContent = isHighQuality ? 
               '🔎 Высокое качество' : '🔍 Низкое качество';
           
           qualityButton.onclick = () => {
               let newChannelId = channelId;
               if (channelId.endsWith('01')) {
                   // Переключаемся с HD на SD
                   newChannelId = channelId.slice(0, -2) + '02';
               } else if (channelId.endsWith('02')) {
                   // Переключаемся с SD на HD
                   newChannelId = channelId.slice(0, -2) + '01';
               }
               
               if (newChannelId !== channelId) {
                   console.log(`🔄 Переключение качества с ${channelId} на ${newChannelId}`);
                   cameraSelect.value = newChannelId;
                   stopCurrentStream();
                   startLiveStream();
               }
           };
           
           videoContainer.appendChild(qualityButton);
       }
       
       // Статус соединения
       const connectionStatus = document.createElement('div');
       connectionStatus.className = 'connection-status';
       connectionStatus.style.position = 'absolute';
       connectionStatus.style.top = '10px';
       connectionStatus.style.right = '10px';
       connectionStatus.style.padding = '5px 10px';
       connectionStatus.style.borderRadius = '4px';
       connectionStatus.style.fontSize = '12px';
       connectionStatus.style.color = 'white';
       connectionStatus.style.background = 'rgba(0, 0, 0, 0.5)';
       connectionStatus.style.zIndex = '100';
       connectionStatus.textContent = '🔄 Установка соединения...';
       videoContainer.appendChild(connectionStatus);
       
       // Обновляем статус соединения при изменении
       pc.addEventListener('iceconnectionstatechange', () => {
           if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
               connectionStatus.textContent = '✅ Соединение установлено';
               connectionStatus.style.background = 'rgba(46, 204, 113, 0.5)';
           } else if (pc.iceConnectionState === 'checking') {
               connectionStatus.textContent = '🔄 Проверка соединения...';
               connectionStatus.style.background = 'rgba(241, 196, 15, 0.5)';
           } else if (pc.iceConnectionState === 'disconnected') {
               connectionStatus.textContent = '⚠️ Соединение нестабильно';
               connectionStatus.style.background = 'rgba(230, 126, 34, 0.5)';
           } else if (pc.iceConnectionState === 'failed') {
               connectionStatus.textContent = '❌ Ошибка соединения';
               connectionStatus.style.background = 'rgba(231, 76, 60, 0.5)';
           }
       });
       
       // Устанавливаем таймер для проверки успешности соединения
       setTimeout(() => {
           if (pc.iceConnectionState !== 'connected' && pc.iceConnectionState !== 'completed') {
               console.warn('⏱️ Медленное соединение:', pc.iceConnectionState);
               
               if (pc.iceConnectionState === 'failed') {
                   console.error('❌ Соединение не установлено');
                   showErrorOverlay(videoContainer, 'Не удалось установить соединение. Попробуйте перезагрузить видео');
               } else if (pc.iceConnectionState === 'checking') {
                   // Продолжаем ждать, но обновляем статус
                   showLoadingIndicator(videoContainer, 'Соединение устанавливается медленно. Пожалуйста, подождите...');
               }
           }
       }, 15000); // 15 секунд
       
   } catch (error) {
       console.error('❌ Ошибка WebRTC:', error);
       showErrorOverlay(videoContainer, 'Ошибка подключения: ' + error.message);
   }
}

/**
* Определение мобильной сети
*/
function detectMobileNetwork() {
   // Проверка через Network Information API
   if (navigator.connection) {
       const connection = navigator.connection;
       const isMobile = 
           connection.type === 'cellular' || 
           connection.effectiveType === '3g' || 
           connection.effectiveType === '2g' ||
           connection.downlink < 2;
       
       console.log('📱 Тип соединения:', connection.type);
       console.log('📡 Эффективный тип:', connection.effectiveType);
       console.log('⚡ Пропускная способность:', connection.downlink, 'Мбит/с');
       
       return isMobile;
   }
   
   // Резервный вариант: проверка через User-Agent
   const userAgent = navigator.userAgent;
   const mobileRegex = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i;
   const isMobileBrowser = mobileRegex.test(userAgent);
   
   return isMobileBrowser;
}

/**
* Настройка базовых обработчиков событий WebRTC
*/
function setupWebRTCEventHandlers(pc, videoElement) {
   // Основные события соединения
   pc.addEventListener('iceconnectionstatechange', () => {
       console.log('📢 Состояние ICE соединения:', pc.iceConnectionState);
       
       if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
           hideLoadingIndicator(videoContainer);
           updateConnectionStatus('online');
       } else if (pc.iceConnectionState === 'checking') {
           showLoadingIndicator(videoContainer, 'Проверка соединения...');
           updateConnectionStatus('connecting');
       } else if (pc.iceConnectionState === 'disconnected') {
           showLoadingIndicator(videoContainer, 'Переподключение...');
           updateConnectionStatus('connecting');
       } else if (pc.iceConnectionState === 'failed') {
           showErrorOverlay(videoContainer, 'Не удалось установить соединение');
           updateConnectionStatus('offline');
       }
   });
   
   pc.addEventListener('connectionstatechange', () => {
       console.log('📢 Состояние соединения:', pc.connectionState);
   });
   
   // Обработчик получения медиа потока
   pc.ontrack = (event) => {
       console.log('📺 Получен медиа трек:', event.track.kind);
       
       if (event.streams && event.streams[0]) {
           videoElement.srcObject = event.streams[0];
           currentStream = event.streams[0];
           
           // Обработчики событий для видео
           videoElement.onloadedmetadata = () => {
               console.log('📐 Размер видео:', videoElement.videoWidth, 'x', videoElement.videoHeight);
               videoElement.play().catch(err => {
                   console.warn('⚠️ Автовоспроизведение не удалось:', err);
                   // Добавляем подсказку о необходимости нажать на видео
                   const playHint = document.createElement('div');
                   playHint.style.position = 'absolute';
                   playHint.style.top = '50%';
                   playHint.style.left = '50%';
                   playHint.style.transform = 'translate(-50%, -50%)';
                   playHint.style.background = 'rgba(0, 0, 0, 0.7)';
                   playHint.style.color = 'white';
                   playHint.style.padding = '15px 20px';
                   playHint.style.borderRadius = '5px';
                   playHint.style.fontSize = '16px';
                   playHint.style.zIndex = '100';
                   playHint.textContent = 'Нажмите для воспроизведения';
                   videoContainer.appendChild(playHint);
                   
                   // Удаляем подсказку при клике
                   videoContainer.addEventListener('click', () => {
                       videoElement.play().catch(e => console.error('Ошибка воспроизведения:', e));
                       if (playHint.parentNode) {
                           playHint.parentNode.removeChild(playHint);
                       }
                   }, { once: true });
               });
           };
           
           videoElement.onplaying = () => {
               console.log('▶️ Видео воспроизводится');
               hideLoadingIndicator(videoContainer);
           };
           
           videoElement.onerror = (error) => {
               console.error('❌ Ошибка видео:', error);
               showErrorOverlay(videoContainer, 'Ошибка воспроизведения видео');
           };
       }
   };
   
   // Логирование ICE кандидатов
   pc.onicecandidate = (event) => {
       if (event.candidate) {
           console.log('🧊 ICE кандидат:', 
               event.candidate.type, 
               event.candidate.protocol, 
               event.candidate.address);
       } else {
           console.log('🧊 Сбор ICE кандидатов завершен');
       }
   };
}

function stopCurrentStream() {
    // Останавливаем WebRTC соединение
    if (currentRTCPeerConnection) {
        try {
            currentRTCPeerConnection.close();
        } catch (e) {
            console.error('Ошибка при закрытии WebRTC соединения:', e);
        }
        currentRTCPeerConnection = null;
    }
    
    // Останавливаем медиа-потоки
    if (currentStream) {
        try {
            currentStream.getTracks().forEach(track => track.stop());
        } catch (e) {
            console.error('Ошибка при остановке медиа-потоков:', e);
        }
        currentStream = null;
    }
    
    // Очищаем видео-элемент
    if (currentVideoElement) {
        try {
            if (currentVideoElement.srcObject) {
                currentVideoElement.srcObject = null;
            }
        } catch (e) {
            console.error('Ошибка при очистке видео-элемента:', e);
        }
        currentVideoElement = null;
    }
    
    // Обновляем статус соединения
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
/**
 * Показывает упрощенный индикатор загрузки для мобильных устройств
 */
function showLoadingIndicator(container, message) {
    // Удаляем предыдущие индикаторы, если они есть
    hideLoadingIndicator(container);
    
    const loadingIndicator = document.createElement('div');
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
    spinner.style.width = '20px';
    spinner.style.height = '20px';
    spinner.style.margin = '0 auto 10px';
    spinner.style.border = '3px solid rgba(255, 255, 255, 0.3)';
    spinner.style.borderRadius = '50%';
    spinner.style.borderTop = '3px solid #fff';
    spinner.style.animation = 'spin 1s linear infinite';
    
    // Добавляем стиль анимации, если его еще нет
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
    
    const text = document.createElement('div');
    text.textContent = message || 'Загрузка...';
    
    loadingIndicator.appendChild(spinner);
    loadingIndicator.appendChild(text);
    
    container.appendChild(loadingIndicator);
}

/**
 * Скрывает индикатор загрузки
 */
function hideLoadingIndicator(container) {
    const indicators = container.querySelectorAll('.video-loading-indicator');
    indicators.forEach(indicator => {
        if (indicator.parentNode === container) {
            container.removeChild(indicator);
        }
    });
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
 * Показывает простое сообщение об ошибке без лишних элементов
 */
function showErrorOverlay(container, message) {
    // Упрощенный вариант для мобильных устройств
    let errorOverlay = document.createElement('div');
    errorOverlay.style.position = 'absolute';
    errorOverlay.style.top = '50%';
    errorOverlay.style.left = '50%';
    errorOverlay.style.transform = 'translate(-50%, -50%)';
    errorOverlay.style.background = 'rgba(231, 76, 60, 0.9)';
    errorOverlay.style.color = 'white';
    errorOverlay.style.padding = '15px 20px';
    errorOverlay.style.borderRadius = '5px';
    errorOverlay.style.textAlign = 'center';
    errorOverlay.style.zIndex = '1000';
    errorOverlay.style.maxWidth = '80%';
    
    const text = document.createElement('div');
    text.textContent = message || 'Произошла ошибка';
    text.style.marginBottom = '10px';
    
    const retryButton = document.createElement('button');
    retryButton.textContent = '🔄 Перезагрузить';
    retryButton.style.padding = '8px 16px';
    retryButton.style.border = 'none';
    retryButton.style.borderRadius = '4px';
    retryButton.style.background = 'white';
    retryButton.style.color = '#e74c3c';
    retryButton.style.cursor = 'pointer';
    
    retryButton.onclick = () => {
        // Удаляем элемент ошибки
        container.removeChild(errorOverlay);
        
        // Перезапускаем поток
        const channelId = cameraSelect.value;
        if (channelId) {
            stopCurrentStream();
            startLiveStream();
        }
    };
    
    errorOverlay.appendChild(text);
    errorOverlay.appendChild(retryButton);
    
    // Удаляем предыдущие ошибки, если они есть
    const oldErrors = container.querySelectorAll('[class^="error"]');
    oldErrors.forEach(el => container.removeChild(el));
    
    container.appendChild(errorOverlay);
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
function addVideoUI(container, channelId, streamData, isMobile) {
    // Добавляем информационную панель
    const infoPanel = document.createElement('div');
    infoPanel.className = 'video-info-panel';
    infoPanel.innerHTML = 
        '<div class="video-info">' +
            '<span>📺 ' + (streamData.channel_name || 'Канал ' + channelId) + '</span>' +
            '<span>🔴 Прямой эфир' + (isMobile ? ' (Мобильное соединение)' : '') + '</span>' +
        '</div>';
    container.appendChild(infoPanel);
    
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
    
    // Добавляем индикатор типа соединения
    const connectionType = document.createElement('div');
    connectionType.style.position = 'absolute';
    connectionType.style.top = '50px';
    connectionType.style.right = '10px';
    connectionType.style.padding = '5px 10px';
    connectionType.style.borderRadius = '4px';
    connectionType.style.fontSize = '12px';
    connectionType.style.color = 'white';
    connectionType.style.background = isMobile ? 
        'rgba(255, 165, 0, 0.7)' : 'rgba(46, 204, 113, 0.7)';
    connectionType.style.zIndex = '100';
    connectionType.textContent = isMobile ? '📱 Мобильное соединение' : '🖥️ WiFi соединение';
    container.appendChild(connectionType);
    
    // Кнопка качества (HD/SD)
    if (channelId.length > 2) { // Только для каналов с возможностью переключения качества
        const qualityButton = document.createElement('button');
        qualityButton.style.position = 'absolute';
        qualityButton.style.top = '90px';
        qualityButton.style.right = '10px';
        qualityButton.style.padding = '5px 10px';
        qualityButton.style.borderRadius = '4px';
        qualityButton.style.fontSize = '12px';
        qualityButton.style.color = 'white';
        qualityButton.style.background = 'rgba(52, 152, 219, 0.7)';
        qualityButton.style.border = 'none';
        qualityButton.style.cursor = 'pointer';
        qualityButton.style.zIndex = '100';
        
        // Определяем текущее качество
        const isHighQuality = channelId.endsWith('01');
        qualityButton.textContent = isHighQuality ? 
            '🔎 Высокое качество' : '🔍 Низкое качество';
        
        qualityButton.onclick = () => {
            let newChannelId = channelId;
            if (channelId.endsWith('01')) {
                newChannelId = channelId.slice(0, -2) + '02'; // HD -> SD
                qualityButton.textContent = '🔍 Низкое качество';
            } else if (channelId.endsWith('02')) {
                newChannelId = channelId.slice(0, -2) + '01'; // SD -> HD
                qualityButton.textContent = '🔎 Высокое качество';
            }
            
            if (newChannelId !== channelId) {
                cameraSelect.value = newChannelId;
                stopCurrentStream();
                startLiveStream();
            }
        };
        
        container.appendChild(qualityButton);
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

/**
 * Определение мобильной сети
 */
function detectMobileNetwork() {
    // Проверка через Network Information API
    if (navigator.connection) {
        const connection = navigator.connection;
        const isMobile = 
            connection.type === 'cellular' || 
            connection.effectiveType === '3g' || 
            connection.effectiveType === '2g' ||
            connection.downlink < 2;
        
        console.log('📱 Тип соединения:', connection.type);
        console.log('📡 Эффективный тип:', connection.effectiveType);
        console.log('⚡ Пропускная способность:', connection.downlink, 'Мбит/с');
        
        return isMobile;
    }
    
    // Резервный вариант: проверка через User-Agent
    const userAgent = navigator.userAgent;
    const mobileRegex = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i;
    const isMobileBrowser = mobileRegex.test(userAgent);
    
    return isMobileBrowser;
}
});