package handlers

import (
	"log"
	"net/http"

	"TeleOko/internal/config"
	"TeleOko/internal/hikvision"
	"TeleOko/internal/streaming"

	"github.com/gin-gonic/gin"
)

// WebRTCOffer структура для получения SDP-оффера от клиента
type WebRTCOffer struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

// WebRTCPlaybackOffer структура для получения SDP-оффера от клиента с URL для воспроизведения
type WebRTCPlaybackOffer struct {
	Offer WebRTCOffer `json:"offer"`
	URL   string      `json:"url"`
}

// Кэш активных соединений WebRTC
var activeConnections = make(map[string]*streaming.WebRTCConnection)

// GetLiveStream обрабатывает запрос на получение прямого эфира
func GetLiveStream(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки конфигурации"})
		return
	}

	// Получение URL для RTSP
	rtspURL := hikvision.GetRTSPURL(
		cfg.Hikvision.IP,
		cfg.Hikvision.Username,
		cfg.Hikvision.Password,
		channel,
		false,
	)

	c.JSON(http.StatusOK, gin.H{
		"channel":  channel,
		"rtsp_url": rtspURL,
	})
}

// HandleWebRTCOffer обрабатывает WebRTC-оффер для прямого эфира
func HandleWebRTCOffer(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получение конфигурации
	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки конфигурации"})
		return
	}

	// Получаем SDP-оффер от клиента
	var offer WebRTCOffer
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат оффера"})
		return
	}

	// Закрываем существующее соединение для этого канала, если есть
	if conn, ok := activeConnections[channel]; ok {
		conn.Close()
		delete(activeConnections, channel)
	}

	// Создаем новое WebRTC-соединение
	webrtcConn, err := streaming.CreateWebRTCConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания WebRTC соединения: " + err.Error()})
		return
	}

	// Сохраняем соединение в кэше
	activeConnections[channel] = webrtcConn

	// Получаем RTSP URL
	rtspURL := hikvision.GetRTSPURL(
		cfg.Hikvision.IP,
		cfg.Hikvision.Username,
		cfg.Hikvision.Password,
		channel,
		false,
	)

	// Получаем RTSP-клиент
	rtspClient, err := streaming.GetRTSPConnection(
		rtspURL,
		cfg.Hikvision.Username,
		cfg.Hikvision.Password,
	)
	if err != nil {
		webrtcConn.Close()
		delete(activeConnections, channel)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подключения к RTSP: " + err.Error()})
		return
	}

	// Обрабатываем SDP-оффер от клиента
	sdpAnswer, err := webrtcConn.HandleOffer(offer.SDP)
	if err != nil {
		webrtcConn.Close()
		delete(activeConnections, channel)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки SDP-оффера: " + err.Error()})
		return
	}

	// Настраиваем проксирование RTSP в WebRTC
	if err := streaming.RTSPtoWebRTC(rtspClient, webrtcConn); err != nil {
		webrtcConn.Close()
		delete(activeConnections, channel)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка настройки проксирования: " + err.Error()})
		return
	}

	// Возвращаем SDP-ответ
	c.JSON(http.StatusOK, gin.H{
		"type": "answer",
		"sdp":  sdpAnswer,
	})
}

// HandlePlaybackOffer обрабатывает WebRTC-оффер для воспроизведения архива
func HandlePlaybackOffer(c *gin.Context) {
	// Получаем SDP-оффер и URL для воспроизведения от клиента
	var playbackOffer WebRTCPlaybackOffer
	if err := c.ShouldBindJSON(&playbackOffer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат оффера"})
		return
	}

	if playbackOffer.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL для воспроизведения не указан"})
		return
	}

	// Генерируем уникальный ключ для этого воспроизведения
	playbackKey := "playback-" + playbackOffer.URL

	// Закрываем существующее соединение для этого URL, если есть
	if conn, ok := activeConnections[playbackKey]; ok {
		conn.Close()
		delete(activeConnections, playbackKey)
	}

	// Создаем новое WebRTC-соединение
	webrtcConn, err := streaming.CreateWebRTCConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания WebRTC соединения: " + err.Error()})
		return
	}

	// Сохраняем соединение в кэше
	activeConnections[playbackKey] = webrtcConn

	// Получаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		webrtcConn.Close()
		delete(activeConnections, playbackKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки конфигурации"})
		return
	}

	// Получаем RTSP-клиент
	rtspClient, err := streaming.GetRTSPConnection(
		playbackOffer.URL,
		cfg.Hikvision.Username,
		cfg.Hikvision.Password,
	)
	if err != nil {
		webrtcConn.Close()
		delete(activeConnections, playbackKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подключения к RTSP: " + err.Error()})
		return
	}

	// Обрабатываем SDP-оффер от клиента
	sdpAnswer, err := webrtcConn.HandleOffer(playbackOffer.Offer.SDP)
	if err != nil {
		webrtcConn.Close()
		delete(activeConnections, playbackKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки SDP-оффера: " + err.Error()})
		return
	}

	// Настраиваем проксирование RTSP в WebRTC
	if err := streaming.RTSPtoWebRTC(rtspClient, webrtcConn); err != nil {
		webrtcConn.Close()
		delete(activeConnections, playbackKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка настройки проксирования: " + err.Error()})
		return
	}

	// Возвращаем SDP-ответ
	c.JSON(http.StatusOK, gin.H{
		"type": "answer",
		"sdp":  sdpAnswer,
	})
}

// CleanupConnections освобождает все соединения
func CleanupConnections() {
	for key, conn := range activeConnections {
		log.Printf("Закрытие соединения: %s", key)
		conn.Close()
	}

	// Очистка кэша соединений
	activeConnections = make(map[string]*streaming.WebRTCConnection)

	// Закрытие всех RTSP-соединений
	streaming.CloseAllConnections()
}
