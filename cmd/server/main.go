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
	"TeleOko/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 Запуск TeleOko v2.0 HLS - Система видеонаблюдения")
	log.Println("==================================================")

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}
	log.Println("✅ Конфигурация загружена")

	// Инициализация HLS менеджера
	handlers.InitStreamManager("web/static/streams")

	// Получение IP-адреса сервера
	ip, err := getLocalIP()
	if err != nil {
		log.Printf("⚠️ Ошибка определения IP сервера: %v", err)
		ip = "127.0.0.1"
	}
	log.Printf("🌐 IP-адрес сервера: %s", ip)

	// Настройка Gin
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Настройка CORS
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

		// Управление HLS потоками
		api.POST("/stream/:channel/stop", handlers.StopStream)
		api.GET("/streams/active", handlers.GetActiveStreams)

		// Архивные записи
		api.GET("/recordings", handlers.GetRecordings)
		api.GET("/playback-url", handlers.GetPlaybackURL)

		// Снимки
		api.GET("/snapshot/:channel", handlers.GetSnapshot)

		// Тестирование подключения к камере
		api.GET("/test-connection", handlers.TestCameraConnection)
	}

	// Обслуживание HLS файлов
	r.GET("/streams/:filename", handlers.ServeHLSPlaylist)
	r.OPTIONS("/streams/:filename", handlers.HandleOptions)

	// Обработка сигналов завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\n🛑 Получен сигнал завершения...")

		// Остановка HLS потоков
		handlers.Cleanup()

		log.Println("👋 TeleOko завершен")
		os.Exit(0)
	}()

	// Отображение информации о запуске
	log.Println()
	log.Println("🎉 TeleOko v2.0 HLS готов к работе!")
	log.Println("====================================")
	log.Printf("🌍 Веб-интерфейс:    http://localhost:%d", cfg.Server.Port)
	log.Printf("🌐 По сети:          http://%s:%d", ip, cfg.Server.Port)
	log.Printf("📺 Каналов:          %d", len(config.GetChannels()))
	log.Printf("🎬 Режим стриминга:  HLS (работает на всех устройствах)")
	log.Println()
	log.Println("⚠️  Убедитесь, что FFmpeg установлен!")
	log.Println()

	// Запуск веб-сервера
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

// getLocalIP получает локальный IP-адрес
func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
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
