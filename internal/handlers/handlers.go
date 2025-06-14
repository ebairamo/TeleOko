package handlers

import (
	"TeleOko/internal/config"
	"TeleOko/internal/hikvision"
	"TeleOko/internal/hls"
	"TeleOko/internal/network"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Глобальный менеджер потоков
var streamManager *hls.StreamManager

// InitStreamManager инициализирует менеджер HLS потоков
func InitStreamManager(outputDir string) {
	streamManager = hls.NewStreamManager(outputDir)
	log.Printf("✅ HLS менеджер инициализирован, директория: %s", outputDir)
}

// GetSystemInfo возвращает информацию о системе
func GetSystemInfo(c *gin.Context) {
	channels := config.GetChannels()
	localIP, _ := network.GetLocalIP()

	activeStreams := 0
	if streamManager != nil {
		activeStreams = len(streamManager.GetActiveStreams())
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "online",
		"version":        "2.0.0-HLS",
		"channels_count": len(channels),
		"streaming_type": "HLS",
		"active_streams": activeStreams,
		"local_ip":       localIP,
		"timestamp":      time.Now().Unix(),
	})
}

// GetChannels возвращает список доступных каналов
func GetChannels(c *gin.Context) {
	channels := config.GetChannels()

	log.Printf("📺 Запрос списка каналов - всего доступно: %d каналов", len(channels))

	// Добавляем информацию об активных потоках
	channelList := make([]gin.H, 0, len(channels))
	for _, channel := range channels {
		channelInfo := gin.H{
			"id":   channel.ID,
			"name": channel.Name,
			"url":  channel.URL,
		}

		// Проверяем, активен ли поток
		if streamManager != nil {
			if stream, exists := streamManager.GetStream(channel.ID); exists && stream.Active {
				channelInfo["streaming"] = true
				channelInfo["stream_url"] = stream.PlaylistURL
			} else {
				channelInfo["streaming"] = false
			}
		}

		channelList = append(channelList, channelInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channelList,
		"count":    len(channels),
	})
}

// GetLiveStream запускает HLS поток для канала
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

	if streamManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Менеджер потоков не инициализирован"})
		return
	}

	// Проверяем, запущен ли уже поток
	stream, exists := streamManager.GetStream(channelID)
	if exists && stream.Active {
		log.Printf("📺 Используем существующий поток для канала %s", channelID)
		c.JSON(http.StatusOK, gin.H{
			"channel":      channelID,
			"channel_name": channel.Name,
			"type":         "hls",
			"stream_url":   stream.PlaylistURL,
			"status":       "active",
		})
		return
	}

	// Запускаем новый поток
	stream, err := streamManager.StartStream(channelID, channel.URL)
	if err != nil {
		log.Printf("❌ Ошибка запуска потока: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Не удалось запустить поток: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel":      channelID,
		"channel_name": channel.Name,
		"type":         "hls",
		"stream_url":   stream.PlaylistURL,
		"status":       "starting",
		"message":      "Поток запускается, подождите несколько секунд",
	})
}

// StopStream останавливает HLS поток
func StopStream(c *gin.Context) {
	channelID := c.Param("channel")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	if streamManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Менеджер потоков не инициализирован"})
		return
	}

	err := streamManager.StopStream(channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	log.Printf("⏹️ Поток остановлен для канала %s", channelID)
	c.JSON(http.StatusOK, gin.H{
		"status":  "stopped",
		"channel": channelID,
	})
}

// GetActiveStreams возвращает список активных потоков
func GetActiveStreams(c *gin.Context) {
	if streamManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Менеджер потоков не инициализирован"})
		return
	}

	streams := streamManager.GetActiveStreams()
	streamList := make([]gin.H, 0, len(streams))

	for _, stream := range streams {
		streamList = append(streamList, gin.H{
			"channel_id": stream.ChannelID,
			"stream_url": stream.PlaylistURL,
			"active":     stream.Active,
			"start_time": stream.StartTime.Format(time.RFC3339),
			"duration":   time.Since(stream.StartTime).Seconds(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"streams": streamList,
		"count":   len(streamList),
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
		"note":       "Используйте VLC для просмотра RTSP потока",
	})
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

// ServeHLSPlaylist обслуживает HLS плейлисты
func ServeHLSPlaylist(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не указан"})
		return
	}

	// Путь к файлу
	filePath := filepath.Join("web/static/streams", filename)

	// Проверяем существование файла
	if _, err := filepath.Abs(filePath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
		return
	}

	// Устанавливаем правильные заголовки для HLS
	if filepath.Ext(filename) == ".m3u8" {
		c.Header("Content-Type", "application/vnd.apple.mpegurl")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	} else if filepath.Ext(filename) == ".ts" {
		c.Header("Content-Type", "video/mp2t")
		c.Header("Cache-Control", "max-age=3600")
	}

	// CORS для HLS
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type")

	// Отдаем файл
	c.File(filePath)
}

// HandleOptions обрабатывает preflight запросы
func HandleOptions(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	c.Header("Access-Control-Max-Age", "86400")
	c.Status(http.StatusNoContent)
}

// Cleanup очищает ресурсы при завершении
func Cleanup() {
	if streamManager != nil {
		log.Println("🧹 Остановка всех HLS потоков...")
		streamManager.StopAll()
	}
}
