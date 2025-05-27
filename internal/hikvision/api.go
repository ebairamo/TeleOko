// internal/hikvision/api.go
package hikvision

import (
	"TeleOko/internal/config"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SearchRecordings ищет записи в архиве Hikvision
func SearchRecordings(channelID, startDate, endDate string) ([]Recording, error) {
	ip, username, password, port := config.GetHikvisionCredentials()

	// Преобразуем дату из dd.mm.yyyy в формат ISO
	startTime, err := parseDate(startDate, "00:00:00")
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга даты начала: %v", err)
	}

	endTime, err := parseDate(endDate, "23:59:59")
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга даты окончания: %v", err)
	}

	// Создаем XML запрос для поиска записей
	searchReq := PlaybackSearchRequest{
		XMLName:              xml.Name{Local: "CMSearchDescription"},
		SearchID:             "1",
		SearchResultPosition: 0,
		MaxResults:           1000,
		SearchMode:           "byTime",
		StartTime:            startTime,
		EndTime:              endTime,
		Channels:             channelID,
	}

	// Сериализуем в XML
	xmlData, err := xml.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания XML запроса: %v", err)
	}

	// Создаем HTTP запрос
	url := fmt.Sprintf("http://%s:%d/ISAPI/ContentMgmt/search", ip, port)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/xml; charset=UTF-8")
	req.SetBasicAuth(username, password)

	// Выполняем запрос
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP: %d, ответ: %s", resp.StatusCode, string(body))
	}

	// Парсим XML ответ
	var searchResp SearchResponse
	if err := xml.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("ошибка парсинга XML ответа: %v", err)
	}

	// Преобразуем в наш формат
	recordings := make([]Recording, 0, len(searchResp.MatchList.Recordings))
	for _, rec := range searchResp.MatchList.Recordings {
		recording := Recording{
			StartTime: formatTimeForAPI(rec.StartTime),
			EndTime:   formatTimeForAPI(rec.EndTime),
			Channel:   channelID,
		}
		recordings = append(recordings, recording)
	}

	return recordings, nil
}

// GetPlaybackURL возвращает URL для воспроизведения архивной записи
func GetPlaybackURL(channelID, startTime, endTime string) (string, error) {
	ip, username, password, port := config.GetHikvisionCredentials()

	// Формируем URL для воспроизведения архива
	playbackURL := fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/tracks/%s?starttime=%s&endtime=%s",
		username, password, ip, port, channelID,
		formatTimeForRTSP(startTime), formatTimeForRTSP(endTime))

	return playbackURL, nil
}

// GetSnapshot получает снимок с камеры
func GetSnapshot(channelID string) ([]byte, error) {
	ip, username, password, port := config.GetHikvisionCredentials()

	// URL для получения снимка
	url := fmt.Sprintf("http://%s:%d/ISAPI/Streaming/channels/%s/picture", ip, port, channelID)

	// Создаем HTTP запрос
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}

	req.SetBasicAuth(username, password)

	// Выполняем запрос
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка HTTP: %d", resp.StatusCode)
	}

	// Читаем изображение
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения изображения: %v", err)
	}

	return imageData, nil
}

// TestConnection проверяет подключение к камере
func TestConnection() error {
	ip, username, password, port := config.GetHikvisionCredentials()

	// Пробуем получить информацию о системе
	url := fmt.Sprintf("http://%s:%d/ISAPI/System/deviceInfo", ip, port)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}

	req.SetBasicAuth(username, password)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("неверные учетные данные")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка HTTP: %d", resp.StatusCode)
	}

	return nil
}

// parseDate преобразует дату из формата dd.mm.yyyy в ISO формат
func parseDate(dateStr, timeStr string) (string, error) {
	// Разбираем дату формата dd.mm.yyyy
	parts := strings.Split(dateStr, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("неверный формат даты: %s", dateStr)
	}

	// Переставляем в формат yyyy-mm-dd
	isoDate := fmt.Sprintf("%s-%s-%s", parts[2], parts[1], parts[0])

	// Добавляем время
	return fmt.Sprintf("%sT%s", isoDate, timeStr), nil
}

// formatTimeForAPI форматирует время для API запросов
func formatTimeForAPI(timeStr string) string {
	// Входящий формат: 2006-01-02T15:04:05Z
	// Выходящий формат: 2006-01-02T15:04:05Z (без изменений)
	return timeStr
}

// formatTimeForRTSP форматирует время для RTSP URL
func formatTimeForRTSP(timeStr string) string {
	// Преобразуем ISO время в формат для RTSP
	// Из: 2006-01-02T15:04:05Z
	// В: 20060102T150405Z
	timeStr = strings.ReplaceAll(timeStr, "-", "")
	timeStr = strings.ReplaceAll(timeStr, ":", "")
	return timeStr
}
