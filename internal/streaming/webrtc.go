package streaming

import (
	"log"
	"sync"
	"time"

	"github.com/deepch/vdk/format/rtsp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

// WebRTCConnection представляет WebRTC-соединение
type WebRTCConnection struct {
	PeerConnection *webrtc.PeerConnection
	VideoTrack     *webrtc.TrackLocalStaticSample
	AudioTrack     *webrtc.TrackLocalStaticSample
	stopChan       chan struct{}
	mutex          sync.Mutex
	isConnected    bool
	lastActivity   time.Time
	ID             string
}

// Хранилище WebRTC-соединений
var (
	webrtcConnections = make(map[string]*WebRTCConnection)
	webrtcMutex       sync.Mutex
)

// CreateWebRTCConnection создает новое WebRTC-соединение
func CreateWebRTCConnection(id string) (*WebRTCConnection, error) {
	// Создаем конфигурацию WebRTC с публичными STUN-серверами
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Создаем новое peer connection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	// Создаем соединение
	conn := &WebRTCConnection{
		PeerConnection: peerConnection,
		stopChan:       make(chan struct{}),
		isConnected:    false,
		lastActivity:   time.Now(),
		ID:             id,
	}

	// Обработчик ICE-соединения
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("[%s] ICE Connection State has changed: %s\n", id, connectionState.String())

		conn.mutex.Lock()
		defer conn.mutex.Unlock()

		// Обновляем время последней активности
		conn.lastActivity = time.Now()

		if connectionState == webrtc.ICEConnectionStateConnected {
			conn.isConnected = true
		} else if connectionState == webrtc.ICEConnectionStateDisconnected ||
			connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateClosed {
			conn.isConnected = false

			// Закрываем соединение при длительной неактивности
			select {
			case conn.stopChan <- struct{}{}:
				// Сигнал отправлен
			default:
				// Канал уже закрыт или заполнен
			}
		}
	})

	// Сохраняем соединение в хранилище
	webrtcMutex.Lock()
	webrtcConnections[id] = conn
	webrtcMutex.Unlock()

	return conn, nil
}

// GetWebRTCConnection получает существующее WebRTC-соединение по ID
func GetWebRTCConnection(id string) *WebRTCConnection {
	webrtcMutex.Lock()
	defer webrtcMutex.Unlock()

	return webrtcConnections[id]
}

// RTSPtoWebRTC проксирует RTSP-поток в WebRTC
func RTSPtoWebRTC(rtspURL string, conn *WebRTCConnection) error {
	// Подключаемся к RTSP-потоку
	rtspClient, err := rtsp.Dial(rtspURL)
	if err != nil {
		return err
	}

	// Создаем видеотрек, если его еще нет
	if conn.VideoTrack == nil {
		videoTrack, err := webrtc.NewTrackLocalStaticSample(
			webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
			"video", "teleoko",
		)
		if err != nil {
			rtspClient.Close()
			return err
		}

		// Добавляем трек в peer connection
		if _, err := conn.PeerConnection.AddTrack(videoTrack); err != nil {
			rtspClient.Close()
			return err
		}

		conn.VideoTrack = videoTrack
	}

	// Запускаем горутину для чтения пакетов из RTSP
	go func() {
		defer rtspClient.Close()

		// Создаем таймер для проверки активности
		keepAliveTimer := time.NewTicker(5 * time.Second)
		defer keepAliveTimer.Stop()

		// Бесконечный цикл чтения пакетов
		for {
			select {
			case <-conn.stopChan:
				// Получен сигнал остановки
				log.Printf("[%s] Остановка RTSP-потока", conn.ID)
				return

			case <-keepAliveTimer.C:
				// Проверка активности соединения
				conn.mutex.Lock()
				isActive := conn.isConnected
				conn.mutex.Unlock()

				if !isActive {
					log.Printf("[%s] WebRTC-соединение не активно, остановка RTSP-потока", conn.ID)
					return
				}

			default:
				// Читаем пакет из RTSP
				pkt, err := rtspClient.ReadPacket()
				if err != nil {
					log.Printf("[%s] Ошибка чтения RTSP-пакета: %v", conn.ID, err)
					return
				}

				// Обрабатываем пакет в зависимости от типа
				// Учитываем, что структура пакета может отличаться в зависимости от библиотеки
				// Предполагаем, что это видеопакет, если он имеет данные
				if len(pkt.Data) > 0 {
					// Создаем медиа-сэмпл
					// Обратите внимание, что поля могут отличаться от предполагаемых
					sample := &media.Sample{
						Data:      pkt.Data,
						Duration:  50 * time.Millisecond, // Примерная длительность кадра при 20 FPS
						Timestamp: time.Now(),            // Текущее время вместо pkt.Time
					}

					// Отправляем пакет в WebRTC-трек
					if conn.VideoTrack != nil {
						if err := conn.VideoTrack.WriteSample(*sample); err != nil {
							log.Printf("[%s] Ошибка отправки видеопакета: %v", conn.ID, err)
						}
					}
				}
			}
		}
	}()

	return nil
}

// HandleOffer обрабатывает SDP-предложение от клиента и создает ответ
func (conn *WebRTCConnection) HandleOffer(offerSDP string) (string, error) {
	// Обновляем время последней активности
	conn.mutex.Lock()
	conn.lastActivity = time.Now()
	conn.mutex.Unlock()

	// Устанавливаем удаленное описание (offer)
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}

	if err := conn.PeerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// Создаем ответ (answer)
	answer, err := conn.PeerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}

	// Устанавливаем локальное описание (answer)
	if err := conn.PeerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}

	return answer.SDP, nil
}

// Close закрывает WebRTC-соединение
func (conn *WebRTCConnection) Close() {
	// Отправляем сигнал остановки
	select {
	case conn.stopChan <- struct{}{}:
		// Сигнал отправлен
	default:
		// Канал уже закрыт или заполнен
	}

	// Закрываем peer connection
	if conn.PeerConnection != nil {
		conn.PeerConnection.Close()
	}

	// Удаляем соединение из хранилища
	webrtcMutex.Lock()
	delete(webrtcConnections, conn.ID)
	webrtcMutex.Unlock()

	log.Printf("[%s] WebRTC-соединение закрыто", conn.ID)
}

// CloseAllWebRTCConnections закрывает все WebRTC-соединения
func CloseAllWebRTCConnections() {
	webrtcMutex.Lock()
	connections := make([]*WebRTCConnection, 0, len(webrtcConnections))
	for _, conn := range webrtcConnections {
		connections = append(connections, conn)
	}
	webrtcMutex.Unlock()

	// Закрываем все соединения
	for _, conn := range connections {
		conn.Close()
	}

	log.Println("Все WebRTC-соединения закрыты")
}

// CleanupInactiveConnections закрывает неактивные WebRTC-соединения
func CleanupInactiveConnections(maxInactivity time.Duration) {
	webrtcMutex.Lock()

	var inactiveConnections []string
	now := time.Now()

	// Находим неактивные соединения
	for id, conn := range webrtcConnections {
		conn.mutex.Lock()
		inactive := now.Sub(conn.lastActivity) > maxInactivity
		conn.mutex.Unlock()

		if inactive {
			inactiveConnections = append(inactiveConnections, id)
		}
	}

	webrtcMutex.Unlock()

	// Закрываем неактивные соединения
	for _, id := range inactiveConnections {
		if conn := GetWebRTCConnection(id); conn != nil {
			log.Printf("[%s] Закрытие неактивного WebRTC-соединения", id)
			conn.Close()
		}
	}
}

// StartWebRTCCleanupRoutine запускает периодическую очистку неактивных соединений
func StartWebRTCCleanupRoutine(interval, maxInactivity time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			CleanupInactiveConnections(maxInactivity)
		}
	}()

	log.Printf("Запущена периодическая очистка WebRTC-соединений (интервал: %v, макс. неактивность: %v)", interval, maxInactivity)
}
