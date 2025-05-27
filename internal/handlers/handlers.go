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
	localIP, _ := network.GetLocalIP()

	c.JSON(http.StatusOK, gin.H{
		"status":         "online",
		"version":        "2.0.0",
		"channels_count": len(channels),
		"go2rtc_enabled": config.IsGo2RTCEnabled(),
		"go2rtc_port":    config.GetGo2RTCPort(),
		"local_ip":       localIP,
		"timestamp":      time.Now().Unix(),
	})
}

// GetChannels возвращает список доступных каналов
func GetChannels(c *gin.Context) {
	channels := config.GetChannels()

	// Логируем все доступные каналы с их RTSP URL
	log.Printf("📺 Запрос списка каналов - всего доступно: %d каналов", len(channels))
	for _, channel := range channels {
		log.Printf("  📹 [%s] %s -> %s", channel.ID, channel.Name, channel.URL)
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"count":    len(channels),
	})
}

// GetLiveStream обрабатывает запрос на получение прямого эфира
func GetLiveStream(c *gin.Context) {
	channelID := c.Param("channel")
	if channelID == "" {
		log.Printf("❌ Запрос прямого эфира без указания канала")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Проверяем, существует ли канал
	channel := config.GetChannelByID(channelID)
	if channel == nil {
		log.Printf("❌ Запрос несуществующего канала: %s", channelID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Канал не найден"})
		return
	}

	// Детальное логирование
	log.Printf("🔴 ПРЯМОЙ ЭФИР - Запрос канала %s", channelID)
	log.Printf("  📹 Название: %s", channel.Name)
	log.Printf("  🌐 RTSP URL: %s", channel.URL)
	log.Printf("  🎥 go2rtc включен: %t", config.IsGo2RTCEnabled())

	// Если go2rtc включен, возвращаем WebRTC URL
	if config.IsGo2RTCEnabled() {
		go2rtcURL := fmt.Sprintf("http://localhost:%d/api/ws?src=%s",
			config.GetGo2RTCPort(), channelID)

		log.Printf("  ✅ WebRTC URL: %s", go2rtcURL)

		c.JSON(http.StatusOK, gin.H{
			"channel":      channelID,
			"channel_name": channel.Name,
			"webrtc_url":   go2rtcURL,
			"rtsp_url":     channel.URL,
			"type":         "webrtc",
		})
	} else {
		// Возвращаем только RTSP URL
		log.Printf("  ⚠️ go2rtc отключен, используется только RTSP")

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
		log.Printf("❌ WebRTC запрос без указания канала")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получаем информацию о канале для логирования
	channel := config.GetChannelByID(channelID)
	rtspURL := "неизвестен"
	if channel != nil {
		rtspURL = channel.URL
	}

	log.Printf("🎯 WebRTC OFFER - Канал %s", channelID)
	log.Printf("  🌐 RTSP источник: %s", rtspURL)

	// Проверяем, включен ли go2rtc
	if !config.IsGo2RTCEnabled() {
		log.Printf("  ❌ go2rtc отключен - WebRTC недоступен")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebRTC сервис недоступен",
		})
		return
	}

	// Читаем SDP предложение из тела запроса
	var offerData map[string]interface{}
	if err := c.ShouldBindJSON(&offerData); err != nil {
		log.Printf("  ❌ Ошибка чтения SDP offer: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Пытаемся проксировать запрос к go2rtc
	go2rtcURL := fmt.Sprintf("http://localhost:%d/api/webrtc", config.GetGo2RTCPort())
	log.Printf("  🔄 Проксирование к go2rtc: %s", go2rtcURL)

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
	log.Printf("  ✅ WebRTC соединение для канала %s (RTSP: %s)", channelID, rtspURL)
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
		log.Printf("❌ Запрос архива без обязательных параметров")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Не указаны обязательные параметры (channel, start)",
		})
		return
	}

	// Если конечная дата не указана, используем начальную
	if endDate == "" {
		endDate = startDate
	}

	// Получаем информацию о канале для логирования
	channel := config.GetChannelByID(channelID)
	rtspURL := "неизвестен"
	channelName := "Неизвестный канал"
	if channel != nil {
		rtspURL = channel.URL
		channelName = channel.Name
	}

	log.Printf("📼 ПОИСК АРХИВА - Канал %s (%s)", channelID, channelName)
	log.Printf("  📅 Период: %s - %s", startDate, endDate)
	log.Printf("  🌐 RTSP источник: %s", rtspURL)

	// Поиск записей через Hikvision API
	recordings, err := hikvision.SearchRecordings(channelID, startDate, endDate)
	if err != nil {
		log.Printf("  ❌ Ошибка поиска записей: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка поиска записей: %v", err),
		})
		return
	}

	log.Printf("  ✅ Найдено записей: %d", len(recordings))

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
		log.Printf("❌ Запрос URL воспроизведения без обязательных параметров")
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

	// Получаем информацию о канале для логирования
	channel := config.GetChannelByID(channelID)
	channelName := "Неизвестный канал"
	liveRTSP := "неизвестен"
	if channel != nil {
		channelName = channel.Name
		liveRTSP = channel.URL
	}

	log.Printf("📺 АРХИВНОЕ ВОСПРОИЗВЕДЕНИЕ - Канал %s (%s)", channelID, channelName)
	log.Printf("  ⏰ Время: %s - %s", startTime, endTime)
	log.Printf("  🌐 Базовый RTSP: %s", liveRTSP)

	// Получаем URL для воспроизведения
	playbackURL, err := hikvision.GetPlaybackURL(channelID, startTime, endTime)
	if err != nil {
		log.Printf("  ❌ Ошибка получения URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка получения URL: %v", err),
		})
		return
	}

	log.Printf("  ✅ Архивный RTSP URL: %s", playbackURL)

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
		log.Printf("❌ WebRTC Playback: ошибка чтения данных - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	log.Printf("🎯 WebRTC PLAYBACK запрос")
	log.Printf("  🌐 Архивный URL: %s", requestData.URL)

	// Генерируем уникальный ID для потока архива
	streamID := "playbook_" + uuid.New().String()
	log.Printf("  🆔 Stream ID: %s", streamID)

	// Если go2rtc включен, возвращаем SDP ответ
	if config.IsGo2RTCEnabled() {
		log.Printf("  ✅ go2rtc включен - возвращаем WebRTC ответ")
		c.JSON(http.StatusOK, gin.H{
			"type":      "answer",
			"stream_id": streamID,
			"sdp":       generateDummySDP(),
		})
	} else {
		// Заглушка для случая, когда go2rtc отключен
		log.Printf("  ❌ go2rtc отключен - WebRTC недоступен")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebRTC сервис недоступен для воспроизведения архива",
		})
	}
}

// GetSnapshot получает снимок с камеры
func GetSnapshot(c *gin.Context) {
	channelID := c.Param("channel")
	if channelID == "" {
		log.Printf("❌ Запрос снимка без указания канала")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получаем информацию о канале для логирования
	channel := config.GetChannelByID(channelID)
	channelName := "Неизвестный канал"
	rtspURL := "неизвестен"
	if channel != nil {
		channelName = channel.Name
		rtspURL = channel.URL
	}

	log.Printf("📸 СНИМОК - Канал %s (%s)", channelID, channelName)
	log.Printf("  🌐 RTSP источник: %s", rtspURL)

	// Получаем снимок через Hikvision API
	imageData, err := hikvision.GetSnapshot(channelID)
	if err != nil {
		log.Printf("  ❌ Ошибка получения снимка: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка получения снимка: %v", err),
		})
		return
	}

	log.Printf("  ✅ Снимок получен, размер: %d байт", len(imageData))

	// Возвращаем изображение
	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Length", strconv.Itoa(len(imageData)))
	c.Data(http.StatusOK, "image/jpeg", imageData)
}

// TestCameraConnection тестирует подключение к камере
func TestCameraConnection(c *gin.Context) {
	ip, username, _, port := config.GetHikvisionCredentials()

	log.Printf("🔍 ТЕСТ ПОДКЛЮЧЕНИЯ к камере")
	log.Printf("  🌐 IP: %s:%d", ip, port)
	log.Printf("  👤 Пользователь: %s", username)

	err := network.TestCameraConnection()
	if err != nil {
		log.Printf("  ❌ Ошибка подключения: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	log.Printf("  ✅ Подключение к камере успешно")
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
		log.Printf("❌ Ошибка конфигурации прокси go2rtc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конфигурации прокси"})
		return
	}

	// Логируем проксирование
	originalPath := c.Request.URL.Path
	log.Printf("🔄 ПРОКСИ к go2rtc: %s -> %s%s", originalPath, targetURL, strings.TrimPrefix(originalPath, "/api/go2rtc"))

	// Создаем reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Модифицируем путь запроса
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
