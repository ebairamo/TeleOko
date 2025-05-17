package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Получение локального IP-адреса
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "127.0.0.1", nil
}

func main() {
	r := gin.Default()

	// Получение IP-адреса
	ip, err := getLocalIP()
	if err != nil {
		log.Printf("Ошибка определения IP: %v", err)
		ip = "127.0.0.1"
	}
	log.Printf("IP-адрес сервера: %s", ip)

	// Статические файлы и шаблоны
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Главная страница
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ip": ip,
		})
	})

	// Заглушки API для демонстрации интерфейса
	api := r.Group("/api")
	{
		// Информация о системе
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"ip":      ip,
				"version": "1.0.0",
				"status":  "online",
			})
		})

		// Проверка соединения
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Заглушка для списка записей
		api.GET("/recordings", func(c *gin.Context) {
			// Генерируем тестовые данные
			recordings := []map[string]string{
				{
					"StartTime": "2025-05-17T08:00:00Z",
					"EndTime":   "2025-05-17T08:15:00Z",
					"Channel":   c.Query("channel"),
				},
				{
					"StartTime": "2025-05-17T12:30:00Z",
					"EndTime":   "2025-05-17T12:45:00Z",
					"Channel":   c.Query("channel"),
				},
				{
					"StartTime": "2025-05-17T18:15:00Z",
					"EndTime":   "2025-05-17T18:30:00Z",
					"Channel":   c.Query("channel"),
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"recordings": recordings,
			})
		})

		// Заглушка для URL воспроизведения
		api.GET("/playback-url", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"url": "rtsp://example/test",
			})
		})

		// Заглушки WebRTC
		api.POST("/webrtc/offer", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"type": "answer",
				"sdp":  "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=msid-semantic: WMS\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:dummy\r\na=ice-pwd:dummy\r\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:toffset\r\na=recvonly\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtcp-fb:96 goog-remb\r\n",
			})
		})

		api.POST("/webrtc/offer/playback", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"type": "answer",
				"sdp":  "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=msid-semantic: WMS\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:dummy\r\na=ice-pwd:dummy\r\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:toffset\r\na=recvonly\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtcp-fb:96 goog-remb\r\n",
			})
		})
	}

	// Запуск сервера
	log.Printf("Запуск сервера на порту 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
