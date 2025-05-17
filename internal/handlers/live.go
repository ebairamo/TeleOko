package handlers

import (
	"net/http"

	"TeleOko/internal/hikvision"

	"github.com/gin-gonic/gin"
)

// GetLiveStream обрабатывает запрос на получение прямого эфира
func GetLiveStream(c *gin.Context) {
	// TODO: Реализовать обработчик для получения прямого эфира
	// 1. Получить параметры запроса (канал)
	// 2. Сформировать URL для RTSP
	// 3. Настроить проксирование RTSP в WebRTC или HLS
	// 4. Вернуть результат

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// TODO: Получить IP и учетные данные из конфигурации
	ip := "192.168.8.15"
	username := "admin"
	password := "oborotni2447"

	// Получение URL для RTSP
	rtspURL := hikvision.GetRTSPURL(ip, username, password, channel, false)

	// TODO: Настроить проксирование в WebRTC или HLS

	c.JSON(http.StatusOK, gin.H{
		"channel":  channel,
		"rtsp_url": rtspURL,
	})
}

// HandleWebRTCOffer обрабатывает WebRTC-оффер
func HandleWebRTCOffer(c *gin.Context) {
	// TODO: Реализовать обработчик для WebRTC-оффера
	// 1. Получить параметры запроса (канал и оффер)
	// 2. Создать WebRTC-соединение
	// 3. Настроить проксирование RTSP в WebRTC
	// 4. Сформировать и вернуть ответ

	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// TODO: Реализовать остальную логику

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"channel": channel,
	})
}
