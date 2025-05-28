package handlers

import (
	"TeleOko/internal/config"
	"TeleOko/internal/hikvision"
	"TeleOko/internal/network"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

	log.Printf("📺 Запрос списка каналов - всего доступно: %d каналов", len(channels))

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

	channel := config.GetChannelByID(channelID)
	if channel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Канал не найден"})
		return
	}

	log.Printf("🔴 ПРЯМОЙ ЭФИР - Запрос канала %s (%s)", channelID, channel.Name)

	if config.IsGo2RTCEnabled() {
		// Возвращаем информацию для WebRTC
		c.JSON(http.StatusOK, gin.H{
			"channel":      channelID,
			"channel_name": channel.Name,
			"type":         "webrtc",
			"rtsp_url":     channel.URL,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"channel":      channelID,
			"channel_name": channel.Name,
			"rtsp_url":     channel.URL,
			"type":         "rtsp",
		})
	}
}

// HandleWebRTCOffer обрабатывает WebRTC offer и проксирует его к go2rtc
// HandleWebRTCOffer обрабатывает WebRTC offer и проксирует его к go2rtc
func HandleWebRTCOffer(c *gin.Context) {
	channelID := c.Query("channel")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	if !config.IsGo2RTCEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "WebRTC сервис недоступен"})
		return
	}

	// Читаем offer от клиента
	var offer map[string]interface{}
	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	// Добавляем подробные логи
	log.Printf("🎯 WebRTC OFFER для канала %s", channelID)

	// Создаем URL для go2rtc API
	go2rtcURL := fmt.Sprintf("http://localhost:%d/api/webrtc?src=%s", config.GetGo2RTCPort(), channelID)
	log.Printf("🔄 Отправка запроса к go2rtc: %s", go2rtcURL)

	// Отправляем offer к go2rtc
	jsonData, err := json.Marshal(offer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
		return
	}

	// Создаем HTTP-клиент с настройками для работы через ngrok
	client := &http.Client{
		Transport: &http.Transport{
			// Устанавливаем таймауты для работы через ngrok
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 30 * time.Second,
			// Отключаем проверку SSL для случаев, когда ngrok использует самоподписанные сертификаты
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 60 * time.Second, // Увеличиваем общий таймаут
	}

	req, err := http.NewRequest("POST", go2rtcURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Ошибка создания запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания запроса к go2rtc"})
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Добавляем заголовки для работы через прокси
	if origin := c.GetHeader("Origin"); origin != "" {
		req.Header.Set("Origin", origin)
	}

	log.Printf("📤 Отправка запроса к go2rtc...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Ошибка отправки к go2rtc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подключения к go2rtc"})
		return
	}
	defer resp.Body.Close()

	log.Printf("📥 Получен ответ от go2rtc: %s", resp.Status)

	// Читаем ответ от go2rtc
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Ошибка чтения ответа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения ответа"})
		return
	}

	// Проверяем, не является ли ответ ошибкой
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ go2rtc вернул ошибку: %s", string(body))
		c.JSON(resp.StatusCode, gin.H{"error": "Ошибка go2rtc", "details": string(body)})
		return
	}

	var answer map[string]interface{}
	if err := json.Unmarshal(body, &answer); err != nil {
		log.Printf("❌ Ошибка парсинга ответа: %v", err)
		log.Printf("📄 Тело ответа: %s", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка парсинга ответа"})
		return
	}

	// Проверяем наличие ошибки в ответе
	if errMsg, exists := answer["error"]; exists {
		log.Printf("❌ go2rtc вернул ошибку: %v", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("go2rtc error: %v", errMsg)})
		return
	}

	log.Printf("✅ WebRTC answer получен для канала %s", channelID)

	// Возвращаем answer клиенту
	c.JSON(http.StatusOK, answer)
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

	if endDate == "" {
		endDate = startDate
	}

	log.Printf("📼 ПОИСК АРХИВА - Канал %s, период: %s - %s", channelID, startDate, endDate)

	recordings, err := hikvision.SearchRecordings(channelID, startDate, endDate)
	if err != nil {
		log.Printf("❌ Ошибка поиска записей: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка поиска записей: %v", err),
		})
		return
	}

	log.Printf("✅ Найдено записей: %d", len(recordings))

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

	if endTime == "" {
		if t, err := time.Parse("2006-01-02T15:04:05Z", startTime); err == nil {
			endTime = t.Add(time.Hour).Format("2006-01-02T15:04:05Z")
		} else {
			endTime = startTime
		}
	}

	log.Printf("📺 АРХИВНОЕ ВОСПРОИЗВЕДЕНИЕ - Канал %s, время: %s - %s", channelID, startTime, endTime)

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

	log.Printf("🎯 WebRTC PLAYBACK запрос")

	if config.IsGo2RTCEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"error": "Воспроизведение архива через WebRTC пока не поддерживается",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebRTC сервис недоступен",
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

	log.Printf("📸 СНИМОК - Канал %s", channelID)

	imageData, err := hikvision.GetSnapshot(channelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Ошибка получения снимка: %v", err),
		})
		return
	}

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
	// Получаем путь запроса, убирая префикс "/api/go2rtc"
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/go2rtc")

	// Создаем URL для go2rtc
	targetURL := fmt.Sprintf("http://localhost:%d%s", config.GetGo2RTCPort(), path)
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Создаем новый запрос
	var body io.Reader
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания запроса"})
		return
	}

	// Копируем заголовки
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Устанавливаем правильный Content-Type для WebRTC
	if strings.Contains(path, "/webrtc") {
		req.Header.Set("Content-Type", "application/json")
	}

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка выполнения запроса"})
		return
	}
	defer resp.Body.Close()

	// Копируем ответ
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
