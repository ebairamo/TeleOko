// internal/handlers/playback.go
package handlers

import (
	"net/http"
	"time"

	"TeleOko/internal/hikvision"
	"TeleOko/internal/network"

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

	// Получение информации о текущей камере для логирования
	camera := network.GetDefaultCamera()
	cameraIP := "неизвестно"
	if camera != nil {
		cameraIP = camera.IP
	}

	// Поиск записей с автоматическим определением IP камеры
	recordings, err := hikvision.SearchRecordings(channel, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recordings": recordings,
		"camera_ip":  cameraIP,
		"start_date": startDate,
		"end_date":   endDate,
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

	// Если конечное время не указано, используем текущее
	if endTime == "" {
		endTime = time.Now().Format("2006-01-02T15:04:05Z")
	}

	// Получение URL для воспроизведения архива с автоматическим определением IP камеры
	url := hikvision.GetPlaybackURL(channel, startTime, endTime)

	// Получение информации о текущей камере для логирования
	camera := network.GetDefaultCamera()
	cameraIP := "неизвестно"
	if camera != nil {
		cameraIP = camera.IP
	}

	c.JSON(http.StatusOK, gin.H{
		"url":        url,
		"camera_ip":  cameraIP,
		"channel":    channel,
		"start_time": startTime,
		"end_time":   endTime,
	})
}
