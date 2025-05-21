// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"TeleOko/internal/config"
	"TeleOko/internal/handlers"
	"TeleOko/internal/network"

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

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Ошибка загрузки конфигурации: %v", err)
	} else {
		log.Println("Конфигурация успешно загружена")
	}

	// Получение IP-адреса сервера
	ip, err := getLocalIP()
	if err != nil {
		log.Printf("Ошибка определения IP сервера: %v", err)
		ip = "127.0.0.1"
	}
	log.Printf("IP-адрес сервера: %s", ip)

	// Запуск автоматического обнаружения камер (если включено в конфигурации)
	if config.IsAutoDiscoveryEnabled() {
		log.Println("Запуск обнаружения камер...")
		// Быстрое начальное сканирование
		cameras := network.FindCameras("", 554, 500*time.Millisecond)
		if len(cameras) > 0 {
			log.Printf("Найдено %d камер в сети", len(cameras))
			for i, camera := range cameras {
				log.Printf("Камера %d: %s", i+1, camera.IP)

				// Добавляем найденную камеру в список предпочтительных
				if err := config.AddPreferredCameraIP(camera.IP); err != nil {
					log.Printf("Ошибка добавления IP в конфигурацию: %v", err)
				}
			}
		} else {
			log.Println("Камеры не найдены. Будет использоваться конфигурация по умолчанию.")
			log.Printf("IP-адрес камеры из конфигурации: %s", cfg.Hikvision.IP)
		}

		// Запуск периодического сканирования в фоновом режиме
		// Интервал из конфигурации или по умолчанию 5 минут
		scanInterval := time.Duration(config.GetScanInterval()) * time.Minute
		network.StartCameraDiscovery(scanInterval)
	} else {
		log.Println("Автоматическое обнаружение камер отключено в конфигурации")
		log.Printf("Используется IP-адрес камеры из конфигурации: %s", cfg.Hikvision.IP)
	}

	// Статические файлы и шаблоны
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Главная страница
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ip": ip,
		})
	})

	// API
	api := r.Group("/api")
	{
		// Информация о системе
		api.GET("/info", func(c *gin.Context) {
			// Получаем список обнаруженных камер
			cameras := network.GetCachedCameras()

			// Получаем текущую активную камеру
			currentCamera := network.GetDefaultCamera()
			currentCameraIP := "не определен"
			if currentCamera != nil {
				currentCameraIP = currentCamera.IP
			}

			c.JSON(http.StatusOK, gin.H{
				"ip":             ip,
				"version":        "1.0.0",
				"status":         "online",
				"cameras_count":  len(cameras),
				"current_camera": currentCameraIP,
			})
		})

		// Проверка соединения
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Список обнаруженных камер
		api.GET("/cameras", func(c *gin.Context) {
			cameras := network.GetCachedCameras()

			// Получаем текущую активную камеру
			currentCamera := network.GetDefaultCamera()
			currentCameraIP := ""
			if currentCamera != nil {
				currentCameraIP = currentCamera.IP
			}

			c.JSON(http.StatusOK, gin.H{
				"cameras":        cameras,
				"count":          len(cameras),
				"current_camera": currentCameraIP,
			})
		})

		// Запуск ручного поиска камер
		api.POST("/scan_cameras", func(c *gin.Context) {
			go network.FindCameras("", 554, 1*time.Second)
			c.JSON(http.StatusOK, gin.H{
				"status":  "scanning",
				"message": "Запущено сканирование камер в фоновом режиме",
			})
		})

		// API для работы с архивом и записями
		api.GET("/recordings", handlers.GetRecordings)
		api.GET("/playback-url", handlers.GetPlaybackURL)

		// API для работы с прямым эфиром
		api.GET("/stream/:channel", handlers.GetLiveStream)
		api.POST("/webrtc/offer", handlers.HandleWebRTCOffer)

		// Заглушка для воспроизведения архива через WebRTC
		api.POST("/webrtc/offer/playback", func(c *gin.Context) {
			// Получаем информацию о текущей камере для включения в ответ
			camera := network.GetDefaultCamera()
			cameraIP := "неизвестно"
			if camera != nil {
				cameraIP = camera.IP
			}

			c.JSON(http.StatusOK, gin.H{
				"type":      "answer",
				"camera_ip": cameraIP,
				"sdp":       "v=0\r\no=- 0 0 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=msid-semantic: WMS\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:dummy\r\na=ice-pwd:dummy\r\na=fingerprint:sha-256 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:toffset\r\na=recvonly\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtcp-fb:96 goog-remb\r\n",
			})
		})

		// Выбор предпочтительной камеры
		api.POST("/set_preferred_camera", func(c *gin.Context) {
			cameraIP := c.PostForm("camera_ip")
			if cameraIP == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP-адрес камеры не указан"})
				return
			}

			// Добавляем IP в список предпочтительных в конфигурации
			err := config.AddPreferredCameraIP(cameraIP)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"message":   "Предпочтительная камера успешно установлена",
				"camera_ip": cameraIP,
			})
		})
	}

	// Запуск сервера
	log.Printf("Запуск сервера на порту %d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
