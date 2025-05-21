// internal/handlers/live.go
package handlers

import (
	"net/http"

	"TeleOko/internal/hikvision"
	"TeleOko/internal/network"

	"github.com/gin-gonic/gin"
)

// GetLiveStream обрабатывает запрос на получение прямого эфира
func GetLiveStream(c *gin.Context) {
	// Получаем параметры запроса (канал)
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получение URL для RTSP с автоматическим определением IP камеры
	rtspURL := hikvision.GetRTSPURL(channel, false)

	// Получение информации о текущей камере для логирования
	camera := network.GetDefaultCamera()
	cameraIP := "неизвестно"
	if camera != nil {
		cameraIP = camera.IP
	}

	// TODO: Настроить проксирование в WebRTC или HLS

	c.JSON(http.StatusOK, gin.H{
		"channel":   channel,
		"rtsp_url":  rtspURL,
		"camera_ip": cameraIP,
	})
}

// HandleWebRTCOffer обрабатывает WebRTC-оффер
func HandleWebRTCOffer(c *gin.Context) {
	// Получаем параметры запроса (канал и оффер)
	channel := c.Query("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Канал не указан"})
		return
	}

	// Получение информации о выбранной камере
	preferredCameraIP := c.Query("camera_ip")
	camera := network.GetBestCamera(preferredCameraIP)

	// Информация о камере для включения в ответ
	cameraInfo := map[string]string{
		"ip":     "неизвестно",
		"status": "неизвестно",
	}

	if camera != nil {
		cameraInfo["ip"] = camera.IP
		cameraInfo["status"] = camera.Status
	}

	// TODO: Реализовать остальную логику
	// 1. Получить оффер из тела запроса
	// 2. Создать WebRTC-соединение
	// 3. Настроить проксирование RTSP в WebRTC
	// 4. Сформировать и вернуть ответ

	// Временная заглушка ответа
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"channel": channel,
		"camera":  cameraInfo,
		"type":    "answer",
		"sdp":     "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=msid-semantic: WMS\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:dummy\r\na=ice-pwd:dummy\r\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:toffset\r\na=recvonly\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtcp-fb:96 goog-remb\r\n",
	})
}
