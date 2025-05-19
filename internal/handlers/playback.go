package handlers

import (
	"net/http"
	"time"

	"TeleOko/internal/config"
	"TeleOko/internal/hikvision"

	"github.com/gin-gonic/gin"
)

// GetRecordings обрабатывает запрос на получение списка записей
func GetRecordings(c *gin.Context) {
	// Получаем параметры запроса
	channel := c.Query("channel")
	startDate := c.Query("start")
	endDate := c.Query("end")

	if channel == "" || startDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не все параметры указаны"})
		return
	}

	// Если конечная дата не указана, используем текущую
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Получаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки конфигурации"})
		return
	}

	// Поиск записей
	recordings, err := hikvision.SearchRecordings(
		cfg.Hikvision.IP,
		cfg.Hikvision.Username,
		cfg.Hikvision.Password,
		channel,
		startDate,
		endDate,
	)

	if err != nil {
		// Если произошла ошибка при поиске записей, возвращаем тестовые данные
		// для возможности демонстрации интерфейса
		c.JSON(http.StatusOK, gin.H{
			"recordings": []map[string]string{
				{
					"StartTime": startDate + "T08:00:00Z",
					"EndTime":   startDate + "T08:15:00Z",
					"Channel":   channel,
				},
				{
					"StartTime": startDate + "T12:30:00Z",
					"EndTime":   startDate + "T12:45:00Z",
					"Channel":   channel,
				},
				{
					"StartTime": startDate + "T18:15:00Z",
					"EndTime":   startDate + "T18:30:00Z",
					"Channel":   channel,
				},
			},
			"warning": "Используются тестовые данные: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings": recordings,
	})
}

// GetPlaybackURL обрабатывает запрос на получение URL для воспроизведения архива
func GetPlaybackURL(c *gin.Context) {
	// Получаем параметры запроса
	channel := c.Query("channel")
	startTime := c.Query("start")
	endTime := c.Query("end")

	if channel == "" || startTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не все параметры указаны"})
		return
	}

	// Если конечное время не указано, используем текущее или добавляем 10 минут к начальному
	if endTime == "" {
		startTimeObj, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			endTime = time.Now().Format(time.RFC3339)
		} else {
			// Добавляем 10 минут к начальному времени
			endTime = startTimeObj.Add(10 * time.Minute).Format(time.RFC3339)
		}
	}

	// Получаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки конфигурации"})
		return
	}

	// Получение URL для воспроизведения архива
	url := hikvision.GetPlaybackURL(
		cfg.Hikvision.IP,
		cfg.Hikvision.Username,
		cfg.Hikvision.Password,
		channel,
		startTime,
		endTime,
	)

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}
