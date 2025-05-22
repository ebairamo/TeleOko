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
	log.Println("🚀 Запуск TeleOko - Система видеонаблюдения")

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
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
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ip":       ip,
			"channels": config.GetChannels(),
		})
	})

	// API группа
	api := r.Group("/api")
	{
		// Информация о системе
		api.GET("/info", handlers.GetSystemInfo)

		// Проверка соединения
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
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

		// Снимки (если понадобятся)
		api.GET("/snapshot/:channel", handlers.GetSnapshot)

		// Тестирование подключения к камере
		api.GET("/test-connection", handlers.TestCameraConnection)

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

	// Запуск веб-сервера
	log.Printf("🌍 Запуск веб-сервера на порту %d", cfg.Server.Port)
	log.Printf("🔗 Откройте браузер: http://localhost:%d", cfg.Server.Port)
	log.Printf("🔗 Или по сети: http://%s:%d", ip, cfg.Server.Port)

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
