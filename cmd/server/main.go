package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"TeleOko/internal/auth"
	"TeleOko/internal/config"
	"TeleOko/internal/handlers"
	"TeleOko/internal/network"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Запуск TeleOko - системы видеонаблюдения")

	// Добавляем флаг для возможности переопределения порта
	portFlag := flag.Int("port", 0, "Порт для запуска сервера (переопределяет настройки конфигурации)")
	flag.Parse()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Предупреждение при загрузке конфигурации: %v. Используются настройки по умолчанию.", err)
	}

	// Проверяем порт из переменной среды
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil && port > 0 {
			log.Printf("Использование порта из переменной среды PORT: %d", port)
			cfg.Server.Port = port
		}
	}

	// Проверяем порт из флага командной строки (приоритетнее чем переменная среды)
	if *portFlag > 0 {
		log.Printf("Использование порта из аргумента командной строки: %d", *portFlag)
		cfg.Server.Port = *portFlag
	}

	// Получение локального IP-адреса
	ip, err := network.GetLocalIP()
	if err != nil {
		log.Printf("Ошибка определения IP: %v", err)
		ip = "127.0.0.1"
	}
	log.Printf("IP-адрес сервера: %s", ip)
	log.Printf("Порт сервера: %d", cfg.Server.Port)

	// Настройка маршрутизатора
	r := setupRouter(cfg, ip)

	// Запуск сервера в отдельной горутине
	go func() {
		address := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Printf("Запуск сервера на порту %d", cfg.Server.Port)
		log.Printf("Веб-интерфейс доступен по адресу: http://localhost:%d", cfg.Server.Port)
		if err := r.Run(address); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидание сигнала для корректного завершения
	waitForShutdown()
}

// setupRouter настраивает маршрутизатор для API и веб-интерфейса
func setupRouter(cfg *config.Config, ip string) *gin.Engine {
	r := gin.Default()

	// Применяем middleware аутентификации
	r.Use(auth.BasicAuth(
		cfg.Auth.Username,
		cfg.Auth.Password,
		cfg.Auth.Enabled,
	))

	// Статические файлы и шаблоны
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Главная страница
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"ip": ip,
		})
	})

	// API группа
	api := r.Group("/api")
	{
		// Информация о системе
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"ip":      ip,
				"version": "1.0.0",
				"status":  "online",
				"port":    cfg.Server.Port,
			})
		})

		// Проверка соединения
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Видеопотоки
		api.GET("/stream/:channel", handlers.GetLiveStream)

		// Поиск записей
		api.GET("/recordings", handlers.GetRecordings)

		// URL для воспроизведения
		api.GET("/playback-url", handlers.GetPlaybackURL)

		// WebRTC API
		api.POST("/webrtc/offer", handlers.HandleWebRTCOffer)
		api.POST("/webrtc/offer/playback", handlers.HandlePlaybackOffer)
	}

	return r
}

// waitForShutdown ожидает сигналы для корректного завершения
func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Получен сигнал завершения. Закрытие соединений...")

	// Очистка ресурсов перед завершением
	handlers.CleanupConnections()

	log.Println("Сервер остановлен")
}
