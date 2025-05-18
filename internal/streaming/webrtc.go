package streaming

import (
	"log"
	"sync"

	"github.com/deepch/vdk/format/rtsp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type WebRTCConnection struct {
	PeerConnection *webrtc.PeerConnection
	VideoTrack     *webrtc.TrackLocalStaticSample
	AudioTrack     *webrtc.TrackLocalStaticSample
	stopChan       chan struct{}
	mutex          sync.Mutex
	isConnected    bool
}

type RTSPStream struct {
	URL      string
	Client   *rtsp.Client
	StopChan chan bool
}

// CreateWebRTCConnection создает новое WebRTC-соединение
func CreateWebRTCConnection() (*WebRTCConnection, error) {
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
	}

	// Обработчик ICE-соединения
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s\n", connectionState.String())

		if connectionState == webrtc.ICEConnectionStateConnected {
			conn.mutex.Lock()
			conn.isConnected = true
			conn.mutex.Unlock()
		} else if connectionState == webrtc.ICEConnectionStateDisconnected ||
			connectionState == webrtc.ICEConnectionStateFailed ||
			connectionState == webrtc.ICEConnectionStateClosed {
			conn.mutex.Lock()
			conn.isConnected = false
			conn.mutex.Unlock()

			// Можно добавить логику для закрытия соединения
		}
	})

	return conn, nil
}

// TODO: Реализовать создание WebRTC-соединения
// 1. Создать конфигурацию WebRTC
// 2. Создать peer connection
// 3. Настроить обработчики событий
// 4. Вернуть соединение

// RTSPtoWebRTC проксирует RTSP-поток в WebRTC
func RTSPtoWebRTC(rtspURL string, conn *WebRTCConnection) error {
	// Подключаемся к RTSP-потоку
	rtspClient, err := rtsp.Dial(rtspURL)
	if err != nil {
		return err
	}

	// Создаем RTSP-поток
	stream := &RTSPStream{
		URL:      rtspURL,
		Client:   rtspClient,
		StopChan: make(chan bool),
	}

	// Запускаем горутину для чтения пакетов из RTSP
	go func() {
		defer rtspClient.Close()

		// Бесконечный цикл чтения пакетов
		for {
			select {
			case <-stream.StopChan:
				// Получен сигнал остановки
				return
			default:
				// Читаем пакет из RTSP
				pkt, err := rtspClient.ReadPacket()
				if err != nil {
					log.Printf("Ошибка чтения RTSP-пакета: %v", err)
					return
				}

				// Если это видеопакет
				if pkt.IsKeyFrame || pkt.IsFrame {
					// Создаем RTP пакет
					sample := &media.Sample{
						Data:      pkt.Data,
						Duration:  pkt.Duration,
						Timestamp: pkt.Time,
					}

					// Отправляем пакет в WebRTC-трек
					if conn.VideoTrack != nil {
						if err := conn.VideoTrack.WriteSample(*sample); err != nil {
							log.Printf("Ошибка отправки видеопакета: %v", err)
						}
					}
				}
			}
		}
	}()

	return nil
}

// TODO: Реализовать проксирование RTSP в WebRTC
// 1. Получить RTSP-соединение
// 2. Создать видео- и аудиотреки
// 3. Добавить треки в peer connection
// 4. Запустить горутину для чтения пакетов из RTSP и записи в WebRTC

// HandleOffer обрабатывает SDP-предложение от клиента и создает ответ
func (conn *WebRTCConnection) HandleOffer(offerSDP string) (string, error) {
	// Устанавливаем удаленное описание (offer)
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerSDP,
	}

	if err := conn.PeerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}

	// Создаем трек для видео
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264},
		"video", "pion",
	)
	if err != nil {
		return "", err
	}
	conn.VideoTrack = videoTrack

	// Добавляем трек в peer connection
	if _, err := conn.PeerConnection.AddTrack(videoTrack); err != nil {
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
