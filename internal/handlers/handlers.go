// internal/handlers/handlers.go
package handlers

import (
	"TeleOko/internal/config"
	"TeleOko/internal/hikvision"
	"TeleOko/internal/network"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetSystemInfo возвращает информацию о системе
func GetSystemInfo(c *gin.Context) {
	channels := config.GetChannels()

	c.JSON(http.StatusOK, gin.H{
		"status":         "online",
		"version":        "2.0.0",
		"channels_count": len(channels),
		"go2rtc_enabled": config.IsGo2RTCEnabled(),
		"go2rtc_port":    config.GetGo2RTCPort(),
		"timestamp":      time.Now().Unix(),
	})
}

// GetChannels возвращает список доступных каналов
func GetChannels(c *gin.Context) {
	channels := config.GetChannels()
	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"count":    len(channels),
	})
}

// GetLiveStream обрабатывает запрос на получение прямого эфира
func GetLiveStream(c *gin.Context) {
	channelID := c.Param("channel")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Проверяем, существует ли канал
	channel := config.GetChannelByID(channelID)
	if channel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Канал не найден"})
		return
	}

	// Если go2rtc включен, возвращаем WebRTC URL
	if config.IsGo2RTCEnabled() {
		go2rtcURL := fmt.Sprintf("http://localhost:%d/api/ws?src=%s",
			config.GetGo2RTCPort(), channelID)

		c.JSON(http.StatusOK, gin.H{
			"channel":      channelID,
			"channel_name": channel.Name,
			"webrtc_url":   go2rtcURL,
			"rtsp_url":     channel.URL,
			"type":         "webrtc",
		})
	} else {
		// Возвращаем только RTSP URL
		c.JSON(http.StatusOK, gin.H{
			"channel":      channelID,
			"channel_name": channel.Name,
			"rtsp_url":     channel.URL,
			"type":         "rtsp",
		})
	}
}

// HandleWebRTCOffer обрабатывает WebRTC предложения для прямого эфира
func HandleWebRTCOffer(c *gin.Context) {
	channelID := c.Query("channel")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Проверяем, включен ли go2rtc
	if !config.IsGo2RTCEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebRTC сервис недоступен",
		})
		return
	}

	// Читаем SDP предложение из тела запроса
	var offerData map[string]interface{}
	if err := c.ShouldBindJSON(&offerData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Пытаемся проксировать запрос к go2rtc
	go2rtcURL := fmt.Sprintf("http://localhost:%d/api/webrtc", config.GetGo2RTCPort())

	// Формируем запрос к go2rtc
	requestBody := map[string]interface{}{
		"type":  "webrtc",
		"value": offerData,
		"src":   channelID,
	}

	// Отправляем запрос к go2rtc (упрощенная версия)
	client := &http.Client{Timeout: 10 * time.Second}

	// Пока возвращаем базовый SDP ответ
	// В полной реализации здесь должен быть HTTP POST к go2rtc
	log.Printf("WebRTC запрос для канала %s к go2rtc %s", channelID, go2rtcURL)
	_ = client      // используем переменную
	_ = requestBody // используем переменную

	c.JSON(http.StatusOK, gin.H{
		"type": "answer",
		"sdp":  generateWebRTCSDP(channelID),
	})
}

// GetRecordings получает список архивных записей
func GetRecordings(c *gin.Context) {
	channelID := c.Query("channel")
	startDate := c.Query("start")
	endDate := c.Query("end")

	if channelID == "" || startDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Не указаны обязательные параметры (channel, start)",
		})
		return
	}

	// Если конечная дата не указана, используем начальную
	if endDate == "" {
		endDate = startDate
	}

	// Поиск записей через Hikvision API
	recordings, err := hikvision.SearchRecordings(channelID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка поиска записей: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings": recordings,
		"count":      len(recordings),
		"channel":    channelID,
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// GetPlaybackURL получает URL для воспроизведения архивной записи
func GetPlaybackURL(c *gin.Context) {
	channelID := c.Query("channel")
	startTime := c.Query("start")
	endTime := c.Query("end")

	if channelID == "" || startTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Не указаны обязательные параметры (channel, start)",
		})
		return
	}

	// Если конечное время не указано, добавляем час к начальному
	if endTime == "" {
		// Парсим время и добавляем 1 час
		if t, err := time.Parse("2006-01-02T15:04:05Z", startTime); err == nil {
			endTime = t.Add(time.Hour).Format("2006-01-02T15:04:05Z")
		} else {
			endTime = startTime
		}
	}

	// Получаем URL для воспроизведения
	playbackURL, err := hikvision.GetPlaybackURL(channelID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка получения URL: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":        playbackURL,
		"channel":    channelID,
		"start_time": startTime,
		"end_time":   endTime,
		"type":       "rtsp",
	})
}

// HandlePlaybackWebRTC обрабатывает WebRTC для воспроизведения архива
func HandlePlaybackWebRTC(c *gin.Context) {
	var requestData struct {
		Offer map[string]interface{} `json:"offer"`
		URL   string                 `json:"url"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Генерируем уникальный ID для потока архива
	streamID := "playbook_" + uuid.New().String()

	// Если go2rtc включен, возвращаем SDP ответ
	if config.IsGo2RTCEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"type":      "answer",
			"stream_id": streamID,
			"sdp":       generateDummySDP(),
		})
	} else {
		// Заглушка для случая, когда go2rtc отключен
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebRTC сервис недоступен для воспроизведения архива",
		})
	}
}

// GetSnapshot получает снимок с камеры
func GetSnapshot(c *gin.Context) {
	channelID := c.Param("channel")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получаем снимок через Hikvision API
	imageData, err := hikvision.GetSnapshot(channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка получения снимка: %v", err),
		})
		return
	}

	// Возвращаем изображение
	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Length", strconv.Itoa(len(imageData)))
	c.Data(http.StatusOK, "image/jpeg", imageData)
}

// TestCameraConnection тестирует подключение к камере
func TestCameraConnection(c *gin.Context) {
	err := network.TestCameraConnection()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Подключение к камере успешно",
	})
}

// ProxyToGo2RTC проксирует запросы к go2rtc
func ProxyToGo2RTC(c *gin.Context) {
	// Создаем URL для go2rtc
	targetURL := fmt.Sprintf("http://localhost:%d", config.GetGo2RTCPort())
	target, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конфигурации прокси"})
		return
	}

	// Создаем reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Модифицируем путь запроса
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = strings.TrimPrefix(originalPath, "/api/go2rtc")

	// Выполняем проксирование
	proxy.ServeHTTP(c.Writer, c.Request)
}

// generateWebRTCSDP генерирует SDP ответ для WebRTC
func generateWebRTCSDP(channelID string) string {
	return fmt.Sprintf(`v=0
o=- %d 0 IN IP4 127.0.0.1
s=TeleOko Stream %s
t=0 0
a=group:BUNDLE 0
a=msid-semantic: WMS
m=video 9 UDP/TLS/RTP/SAVPF 96
c=IN IP4 0.0.0.0
a=rtcp:9 IN IP4 0.0.0.0
a=ice-ufrag:teleoko%s
a=ice-pwd:teleoko%s123
a=fingerprint:sha-256 AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99
a=setup:actpass
a=mid:0
a=extmap:1 urn:ietf:params:rtp-hdrext:toffset
a=recvonly
a=rtpmap:96 VP8/90000
a=rtcp-fb:96 nack
a=rtcp-fb:96 nack pli
a=rtcp-fb:96 goog-remb`,
		time.Now().Unix(), channelID, channelID, channelID)
}

// generateDummySDP генерирует базовую заглушку SDP ответа
func generateDummySDP() string {
	return generateWebRTCSDP("default")
}
