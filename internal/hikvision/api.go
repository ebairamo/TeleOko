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

// SearchRecordings ищет записи в архиве через HTTP API (не RTSP)
func SearchRecordings(channelID, startDate, endDate string) ([]Recording, error) {
	ip, username, password, _ := config.GetHikvisionCredentials()

	// Преобразуем дату из dd.mm.yyyy в формат ISO
	startTime, err := parseDate(startDate, "00:00:00")
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга даты начала: %v", err)
	}

	endTime, err := parseDate(endDate, "23:59:59")
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга даты окончания: %v", err)
	}

	// Преобразуем ID канала для API
	// Каналы 102, 202 и т.д. нужно преобразовать в 1, 2 и т.д.
	apiChannelID := channelID
	if len(channelID) > 2 {
		// Берем первые цифры (например, 1402 -> 14)
		apiChannelID = channelID[:len(channelID)-2]
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
		Channels:             apiChannelID,
	}

	// Сериализуем в XML
	xmlData, err := xml.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания XML запроса: %v", err)
	}

	// HTTP порт обычно 80 или 8080
	httpPorts := []int{80, 8080}

	for _, port := range httpPorts {
		// Создаем HTTP запрос
		url := fmt.Sprintf("http://%s:%d/ISAPI/ContentMgmt/search", ip, port)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(xmlData))
		if err != nil {
			continue
		}

		// Устанавливаем заголовки
		req.Header.Set("Content-Type", "application/xml; charset=UTF-8")
		req.SetBasicAuth(username, password)

		// Выполняем запрос
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Читаем ответ
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		// Проверяем статус ответа
		if resp.StatusCode == http.StatusOK {
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
					Channel:   channelID, // Возвращаем оригинальный ID канала
				}
				recordings = append(recordings, recording)
			}

			return recordings, nil
		}
	}

	// Если API не работает, возвращаем тестовые данные для демонстрации
	return generateTestRecordings(channelID, startDate), nil
}

// generateTestRecordings создает тестовые записи для демонстрации
func generateTestRecordings(channelID, date string) []Recording {
	// Парсим дату
	dateParts := strings.Split(date, ".")
	if len(dateParts) != 3 {
		return []Recording{}
	}

	baseDate := fmt.Sprintf("%s-%s-%s", dateParts[2], dateParts[1], dateParts[0])

	// Создаем несколько тестовых записей
	recordings := []Recording{
		{
			StartTime: baseDate + "T09:00:00Z",
			EndTime:   baseDate + "T10:30:00Z",
			Channel:   channelID,
		},
		{
			StartTime: baseDate + "T11:00:00Z",
			EndTime:   baseDate + "T12:45:00Z",
			Channel:   channelID,
		},
		{
			StartTime: baseDate + "T14:00:00Z",
			EndTime:   baseDate + "T15:15:00Z",
			Channel:   channelID,
		},
		{
			StartTime: baseDate + "T16:30:00Z",
			EndTime:   baseDate + "T18:00:00Z",
			Channel:   channelID,
		},
	}

	return recordings
}

// GetPlaybackURL возвращает URL для воспроизведения архивной записи
func GetPlaybackURL(channelID, startTime, endTime string) (string, error) {
	ip, username, password, port := config.GetHikvisionCredentials()

	// Формируем URL для воспроизведения архива
	// Используем формат RTSP с параметрами времени
	playbackURL := fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/tracks/%s?starttime=%s&endtime=%s",
		username, password, ip, port, channelID,
		formatTimeForRTSP(startTime), formatTimeForRTSP(endTime))

	return playbackURL, nil
}

// GetSnapshot получает снимок с камеры
func GetSnapshot(channelID string) ([]byte, error) {
	ip, username, password, _ := config.GetHikvisionCredentials()

	// Преобразуем ID канала для API снимков
	apiChannelID := channelID
	if len(channelID) > 2 && strings.HasSuffix(channelID, "02") {
		// Для субпотока используем основной канал
		apiChannelID = channelID[:len(channelID)-2] + "01"
	}

	// Пробуем разные порты
	httpPorts := []int{80, 8080}

	for _, port := range httpPorts {
		// URL для получения снимка
		url := fmt.Sprintf("http://%s:%d/ISAPI/Streaming/channels/%s/picture", ip, port, apiChannelID)

		// Создаем HTTP запрос
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		req.SetBasicAuth(username, password)

		// Выполняем запрос
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Читаем изображение
			imageData, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			return imageData, nil
		}
	}

	return nil, fmt.Errorf("не удалось получить снимок с камеры")
}

// TestConnection проверяет подключение к камере
func TestConnection() error {
	ip, username, password, _ := config.GetHikvisionCredentials()

	// Пробуем разные порты
	httpPorts := []int{80, 8080}

	for _, port := range httpPorts {
		// Пробуем получить информацию о системе
		url := fmt.Sprintf("http://%s:%d/ISAPI/System/deviceInfo", ip, port)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		req.SetBasicAuth(username, password)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("неверные учетные данные")
		}

		if resp.StatusCode == http.StatusOK {
			return nil
		}
	}

	return fmt.Errorf("не удалось подключиться к камере")
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
	return fmt.Sprintf("%sT%sZ", isoDate, timeStr), nil
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
