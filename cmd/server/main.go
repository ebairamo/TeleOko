// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"TeleOko/internal/config"
	"TeleOko/internal/go2rtc"
	"TeleOko/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
<<<<<<< HEAD
	log.Println("🚀 Запуск TeleOko v2.0 - Система видеонаблюдения с автообнаружением")
	log.Println("=====================================================================")

	// Загрузка конфигурации (с автообнаружением камер)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
=======
	log.Println("🚀 Запуск TeleOko - Система видеонаблюдения")

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
	}
	log.Println("✅ Конфигурация загружена")

	// Получение IP-адреса сервера
	ip, err := getLocalIP()
	if err != nil {
		log.Printf("⚠️ Ошибка определения IP сервера: %v", err)
		ip = "127.0.0.1"
	}
	log.Printf("🌐 IP-адрес сервера: %s", ip)

	// Запуск go2rtc если включен
	var go2rtcManager *go2rtc.Manager
	if config.IsGo2RTCEnabled() {
<<<<<<< HEAD
		log.Println("🎥 Инициализация go2rtc...")
		go2rtcManager = go2rtc.NewManager()

		if err := go2rtcManager.Start(); err != nil {
			log.Printf("⚠️ Ошибка запуска go2rtc: %v", err)
			log.Println("📝 go2rtc будет отключен, система будет работать только с RTSP")
		} else {
			log.Println("✅ go2rtc успешно запущен")

			// Ждем полного запуска go2rtc
			time.Sleep(3 * time.Second)

			// Обновляем потоки (особенно если IP камеры изменился)
			if config.ShouldUpdateGo2RTC() {
				log.Println("🔄 Обновление конфигурации go2rtc с новым IP камеры...")
				if err := go2rtcManager.UpdateStreams(); err != nil {
					log.Printf("⚠️ Ошибка обновления потоков: %v", err)
				} else {
					log.Println("✅ Конфигурация go2rtc обновлена")
				}
			}
		}
	} else {
		log.Println("📺 go2rtc отключен - система будет работать только с RTSP URL")
=======
		log.Println("🎥 Запуск go2rtc...")
		go2rtcManager = go2rtc.NewManager()
		if err := go2rtcManager.Start(); err != nil {
			log.Fatalf("❌ Ошибка запуска go2rtc: %v", err)
		}
		log.Println("✅ go2rtc успешно запущен")

		// Добавляем потоки
		time.Sleep(3 * time.Second) // Ждем полного запуска go2rtc
		if err := go2rtcManager.UpdateStreams(); err != nil {
			log.Printf("⚠️ Ошибка обновления потоков: %v", err)
		}
>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
	}

	// Настройка Gin
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Настройка CORS для WebRTC
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Статические файлы и шаблоны
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Главная страница
	r.GET("/", func(c *gin.Context) {
<<<<<<< HEAD
		// Получаем актуальную информацию о камере
		cameraIP, _, _, _ := config.GetHikvisionCredentials()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"ip":        ip,
			"camera_ip": cameraIP,
			"channels":  config.GetChannels(),
=======
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ip":       ip,
			"channels": config.GetChannels(),
>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
		})
	})

	// API группа
	api := r.Group("/api")
	{
		// Информация о системе
		api.GET("/info", handlers.GetSystemInfo)

		// Проверка соединения
		api.GET("/ping", func(c *gin.Context) {
<<<<<<< HEAD
			cameraIP, _, _, _ := config.GetHikvisionCredentials()
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"camera_ip": cameraIP,
			})
=======
			c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
		})

		// Работа с каналами
		api.GET("/channels", handlers.GetChannels)

		// Прямой эфир
		api.GET("/stream/:channel", handlers.GetLiveStream)
		api.POST("/webrtc/offer", handlers.HandleWebRTCOffer)

		// Архивные записи
		api.GET("/recordings", handlers.GetRecordings)
		api.GET("/playback-url", handlers.GetPlaybackURL)
		api.POST("/webrtc/offer/playback", handlers.HandlePlaybackWebRTC)

<<<<<<< HEAD
		// Снимки
=======
		// Снимки (если понадобятся)
>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
		api.GET("/snapshot/:channel", handlers.GetSnapshot)

		// Тестирование подключения к камере
		api.GET("/test-connection", handlers.TestCameraConnection)

<<<<<<< HEAD
		// Ручное обновление IP камеры
		api.POST("/update-camera-ip", func(c *gin.Context) {
			var request struct {
				IP string `json:"ip"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
				return
			}

			// Проверяем формат IP
			if net.ParseIP(request.IP) == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат IP адреса"})
				return
			}

			// Обновляем IP
			config.SetHikvisionIP(request.IP)

			// Сохраняем конфигурацию
			if err := config.Save(); err != nil {
				log.Printf("⚠️ Ошибка сохранения конфигурации: %v", err)
			}

			// Обновляем go2rtc если нужно
			if go2rtcManager != nil {
				go func() {
					time.Sleep(1 * time.Second)
					if err := go2rtcManager.UpdateStreams(); err != nil {
						log.Printf("⚠️ Ошибка обновления go2rtc потоков: %v", err)
					}
				}()
			}

			c.JSON(http.StatusOK, gin.H{
				"status":  "success",
				"message": fmt.Sprintf("IP камеры обновлен на %s", request.IP),
				"new_ip":  request.IP,
			})
		})

		// Повторное автообнаружение камер
		api.POST("/rediscover-cameras", func(c *gin.Context) {
			log.Println("🔍 Запущено повторное обнаружение камер...")

			// Временно включаем автообнаружение
			oldAutoDetect := cfg.Hikvision.AutoDetect
			cfg.Hikvision.AutoDetect = true

			// Перезагружаем конфигурацию с автообнаружением
			if _, err := config.Load(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Ошибка повторного обнаружения: " + err.Error(),
				})
				return
			}

			// Восстанавливаем настройку автообнаружения
			cfg.Hikvision.AutoDetect = oldAutoDetect

			// Получаем обновленный IP
			newIP, _, _, _ := config.GetHikvisionCredentials()

			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"message":   "Повторное обнаружение выполнено",
				"camera_ip": newIP,
			})
		})

=======
>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
		// Проксирование запросов к go2rtc
		if go2rtcManager != nil {
			api.Any("/go2rtc/*path", handlers.ProxyToGo2RTC)
		}
	}

	// Обработка сигналов завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\n🛑 Получен сигнал завершения...")

		// Остановка go2rtc
		if go2rtcManager != nil {
			log.Println("⏹️ Остановка go2rtc...")
			if err := go2rtcManager.Stop(); err != nil {
				log.Printf("⚠️ Ошибка остановки go2rtc: %v", err)
			}
		}

		log.Println("👋 TeleOko завершен")
		os.Exit(0)
	}()

<<<<<<< HEAD
	// Отображение информации о запуске
	cameraIP, username, _, _ := config.GetHikvisionCredentials()
	channels := config.GetChannels()

	log.Println()
	log.Println("🎉 TeleOko v2.0 готов к работе!")
	log.Println("================================")
	log.Printf("🌍 Веб-интерфейс:    http://localhost:%d", cfg.Server.Port)
	log.Printf("🌐 По сети:          http://%s:%d", ip, cfg.Server.Port)
	log.Printf("📹 IP камеры:        %s (пользователь: %s)", cameraIP, username)
	log.Printf("📺 Каналов:          %d", len(channels))
	if config.IsGo2RTCEnabled() {
		log.Printf("🎥 go2rtc:           http://localhost:%d", config.GetGo2RTCPort())
	}
	log.Println()
	log.Println("💡 Функции:")
	log.Println("   ✅ Автоматическое обнаружение камер")
	log.Println("   ✅ WebRTC стрим в браузере")
	log.Println("   ✅ Архивные записи")
	log.Println("   ✅ Снимки с камер")
	log.Println("   ✅ Мобильная версия")
	log.Println()

	// Запуск веб-сервера
=======
	// Запуск веб-сервера
	log.Printf("🌍 Запуск веб-сервера на порту %d", cfg.Server.Port)
	log.Printf("🔗 Откройте браузер: http://localhost:%d", cfg.Server.Port)
	log.Printf("🔗 Или по сети: http://%s:%d", ip, cfg.Server.Port)

>>>>>>> 17ac5825c609ebcafbd76fbf2fa91fe09048c1ca
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

// getLocalIP получает локальный IP-адрес
func getLocalIP() (string, error) {
	// Создаем UDP соединение для определения локального IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Если не удалось, пробуем через интерфейсы
		return getLocalIPFromInterfaces()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// getLocalIPFromInterfaces получает IP через сетевые интерфейсы
func getLocalIPFromInterfaces() (string, error) {
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
