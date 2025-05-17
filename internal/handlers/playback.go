package handlers

import (
	"net/http"
	"time"

	"github.com/ebairamo/TeleOco/internal/hikvision"
	"github.com/gin-gonic/gin"
)

// GetRecordings обрабатывает запрос на получение списка записей
func GetRecordings(c *gin.Context) {
	// TODO: Реализовать обработчик для получения списка записей
	// 1. Получить параметры запроса (канал, даты)
	// 2. Вызвать API для поиска записей
	// 3. Обработать результат
	// 4. Вернуть список записей

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

	// TODO: Получить IP и учетные данные из конфигурации
	ip := "192.168.8.15"
	username := "admin"
	password := "oborotni2447"

	// Поиск записей
	recordings, err := hikvision.SearchRecordings(ip, username, password, channel, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings": recordings,
	})
}

// GetPlaybackURL обрабатывает запрос на получение URL для воспроизведения архива
func GetPlaybackURL(c *gin.Context) {
	// TODO: Реализовать обработчик для получения URL воспроизведения архива
	// 1. Получить параметры запроса (канал, время начала и конца)
	// 2. Сформировать URL для воспроизведения архива
	// 3. Вернуть URL

	channel := c.Query("channel")
	startTime := c.Query("start")
	endTime := c.Query("end")

	if channel == "" || startTime == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не все параметры указаны"})
		return
	}

	// Если конечное время не указано, используем текущее
	if endTime == "" {
		endTime = time.Now().Format("2006-01-02T15:04:05Z")
	}

	// TODO: Получить IP и учетные данные из конфигурации
	ip := "192.168.8.15"
	username := "admin"
	password := "oborotni2447"

	// Получение URL для воспроизведения архива
	url := hikvision.GetPlaybackURL(ip, username, password, channel, startTime, endTime)

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}
