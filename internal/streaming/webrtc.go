package streaming

import (
	"log"
	"sync"
	"time"

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

	// Создаем настройки медиа для WebRTC
	mediaEngine := webrtc.MediaEngine{}

	// Регистрируем кодеки
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	// Настраиваем API для WebRTC
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	// Создаем новое peer connection
	peerConnection, err := api.NewPeerConnection(config)
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

			// Сигнал для остановки потока
			select {
			case conn.stopChan <- struct{}{}:
			default:
			}
		}
	})

	return conn, nil
}

// RTSPtoWebRTC проксирует RTSP-поток в WebRTC
func RTSPtoWebRTC(rtspClient *RTSPClient, conn *WebRTCConnection) error {
	// Создаем трек для видео - поддерживаем разные кодеки
	var videoTrack *webrtc.TrackLocalStaticSample
	var err error

	// Пробуем использовать H264
	videoTrack, err = webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264,
		},
		"video", "teleoko",
	)

	if err != nil {
		// Если H264 не поддерживается, пробуем VP8
		videoTrack, err = webrtc.NewTrackLocalStaticSample(
			webrtc.RTPCodecCapability{
				MimeType: webrtc.MimeTypeVP8,
			},
			"video", "teleoko",
		)

		if err != nil {
			return err
		}
	}

	conn.VideoTrack = videoTrack

	// Добавляем видеотрек в peer connection
	rtpSender, err := conn.PeerConnection.AddTrack(videoTrack)
	if err != nil {
		return err
	}

	// Обрабатываем RTCP-пакеты
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Запускаем горутину для проксирования RTSP в WebRTC
	go func() {
		defer func() {
			log.Println("Остановка проксирования RTSP в WebRTC")
		}()

		for {
			select {
			case <-conn.stopChan:
				return
			default:
				// Получаем пакет из RTSP
				packet, err := rtspClient.GetPacket()
				if err != nil {
					log.Printf("Ошибка получения пакета RTSP: %v", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				// Обрабатываем пакет
				sample := media.Sample{
					Data:      packet.Data,
					Duration:  packet.Duration,
					Timestamp: time.Now(),
				}

				// Отправляем в WebRTC-трек
				if err := videoTrack.WriteSample(sample); err != nil {
					log.Printf("Ошибка отправки видеопакета: %v", err)
				}
			}
		}
	}()

	return nil
}

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
	// Сигнал для остановки потока
	select {
	case conn.stopChan <- struct{}{}:
	default:
	}

	// Закрываем peer connection
	if conn.PeerConnection != nil {
		conn.PeerConnection.Close()
	}
}
