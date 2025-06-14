/**
 * Универсальный WebRTC клиент для TeleOko
 * Работает на всех устройствах и типах сетей
 */

class WebRTCClient {
    constructor(options = {}) {
        this.options = {
            go2rtcBaseUrl: null,  // URL для go2rtc
            onStatusChange: null, // Коллбэк изменения статуса
            onStream: null,       // Коллбэк получения потока
            onError: null,        // Коллбэк ошибки
            ...options
        };
        
        this.peerConnection = null;
        this.stream = null;
        this.isMobileNetwork = this.detectMobileNetwork();
        this.connectionState = 'disconnected';
        this.retryCount = 0;
        this.maxRetries = 3;
    }
    
    // Определить тип сети
    detectMobileNetwork() {
        if (navigator.connection) {
            return navigator.connection.type === 'cellular' || 
                   navigator.connection.effectiveType === '3g' || 
                   navigator.connection.effectiveType === '2g' ||
                   navigator.connection.downlink < 2;
        }
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    }
    
    // Получить URL для go2rtc
    async getGo2rtcUrl() {
        // Если URL уже задан в опциях, используем его
        if (this.options.go2rtcBaseUrl) {
            return this.options.go2rtcBaseUrl;
        }
        
        // Пробуем получить из localStorage
        const savedUrl = localStorage.getItem('go2rtcUrl');
        if (savedUrl) {
            return savedUrl;
        }
        
        // Пробуем получить из API
        try {
            const response = await fetch('/api/info');
            if (response.ok) {
                const data = await response.json();
                if (data.go2rtc_url) {
                    localStorage.setItem('go2rtcUrl', data.go2rtc_url);
                    return data.go2rtc_url;
                }
            }
        } catch (error) {
            console.warn('Не удалось получить URL из API:', error);
        }
        
        // Пробуем получить из файла
        try {
            const fileResponse = await fetch('/go2rtc_url.txt');
            if (fileResponse.ok) {
                const url = await fileResponse.text();
                const trimmedUrl = url.trim();
                localStorage.setItem('go2rtcUrl', trimmedUrl);
                return trimmedUrl;
            }
        } catch (error) {
            console.warn('Не удалось прочитать go2rtc_url.txt:', error);
        }
        
        return null;
    }
    
    // Установить статус соединения
    setConnectionState(state) {
        this.connectionState = state;
        if (this.options.onStatusChange) {
            this.options.onStatusChange(state);
        }
    }
    
    // Подключиться к камере
    async connect(channelId) {
        try {
            // Остановить предыдущее соединение
            this.disconnect();
            
            // Получить URL go2rtc
            const go2rtcUrl = await this.getGo2rtcUrl();
            if (!go2rtcUrl) {
                throw new Error('URL go2rtc не найден');
            }
            
            // Создать RTCPeerConnection с оптимальными настройками
            const pcConfig = {
                iceServers: [
                    { urls: 'stun:stun.l.google.com:19302' },
                    { urls: 'stun:stun1.l.google.com:19302' },
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
                iceTransportPolicy: this.isMobileNetwork ? 'relay' : 'all',
                bundlePolicy: 'max-bundle',
                rtcpMuxPolicy: 'require',
                iceCandidatePoolSize: 10
            };
            
            this.peerConnection = new RTCPeerConnection(pcConfig);
            
            // Обработчики событий
            this.peerConnection.oniceconnectionstatechange = () => {
                const state = this.peerConnection.iceConnectionState;
                console.log('ICE состояние:', state);
                
                if (state === 'connected' || state === 'completed') {
                    this.setConnectionState('connected');
                    this.retryCount = 0;
                } else if (state === 'checking') {
                    this.setConnectionState('connecting');
                } else if (state === 'disconnected') {
                    this.setConnectionState('disconnected');
                    // Пробуем перезапустить ICE
                    if (this.peerConnection.restartIce) {
                        this.peerConnection.restartIce();
                    }
                } else if (state === 'failed') {
                    this.setConnectionState('failed');
                    
                    // Автоматический повтор подключения
                    if (this.retryCount < this.maxRetries) {
                        this.retryCount++;
                        console.log(`Попытка переподключения ${this.retryCount}/${this.maxRetries}`);
                        this.reconnect(channelId);
                    } else if (this.options.onError) {
                        this.options.onError(new Error('Не удалось установить WebRTC соединение'));
                    }
                }
            };
            
            this.peerConnection.ontrack = (event) => {
                console.log('Получен медиа-трек');
                if (event.streams && event.streams[0]) {
                    this.stream = event.streams[0];
                    if (this.options.onStream) {
                        this.options.onStream(this.stream);
                    }
                }
            };
            
            // Добавляем трансивер для получения видео
            this.peerConnection.addTransceiver('video', { direction: 'recvonly' });
            
            // Создаем SDP offer
            const offer = await this.peerConnection.createOffer({
                offerToReceiveVideo: true,
                offerToReceiveAudio: false,
                iceRestart: true
            });
            
            // Для мобильных сетей оптимизируем битрейт
            if (this.isMobileNetwork) {
                let sdp = offer.sdp;
                sdp = sdp.replace(/a=rtpmap:(96|97|98) VP8\/90000\r\n/g, 
                    'a=rtpmap:$1 VP8/90000\r\na=fmtp:$1 x-google-min-bitrate=100;x-google-max-bitrate=500;x-google-start-bitrate=300\r\n');
                offer.sdp = sdp;
            }
            
            await this.peerConnection.setLocalDescription(offer);
            
            // Прямой запрос к go2rtc (без прокси через TeleOko)
            const requestUrl = `${go2rtcUrl}/api/webrtc?src=${channelId}`;
            console.log('URL запроса:', requestUrl);
            
            const response = await fetch(requestUrl, {
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
            
            const answer = await response.json();
            
            if (answer.error) {
                throw new Error(answer.error);
            }
            
            if (!answer.sdp) {
                throw new Error('SDP отсутствует в ответе');
            }
            
            await this.peerConnection.setRemoteDescription(new RTCSessionDescription({
                type: 'answer',
                sdp: answer.sdp
            }));
            
            return true;
            
        } catch (error) {
            console.error('Ошибка подключения:', error);
            this.setConnectionState('failed');
            if (this.options.onError) {
                this.options.onError(error);
            }
            return false;
        }
    }
    
    // Переподключение
    async reconnect(channelId) {
        // Короткая задержка перед переподключением
        await new Promise(resolve => setTimeout(resolve, 1000));
        return this.connect(channelId);
    }
    
    // Отключение
    disconnect() {
        if (this.peerConnection) {
            this.peerConnection.close();
            this.peerConnection = null;
        }
        
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
            this.stream = null;
        }
        
        this.setConnectionState('disconnected');
    }
}

// Экспортируем класс
window.WebRTCClient = WebRTCClient;