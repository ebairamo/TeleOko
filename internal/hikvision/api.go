package hikvision

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SearchRecordings ищет записи в архиве
func SearchRecordings(ip, username, password, channel, startDate, endDate string) ([]Recording, error) {
	// Формируем URL для API
	url := fmt.Sprintf("http://%s/ISAPI/ContentMgmt/search", ip)

	// Если дата конца не указана, используем текущую
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02T15:04:05Z")
	}

	// Если дата начала указана только в формате YYYY-MM-DD, добавляем время
	if len(startDate) == 10 {
		startDate = startDate + "T00:00:00Z"
	}

	// Если дата конца указана только в формате YYYY-MM-DD, добавляем время
	if len(endDate) == 10 {
		endDate = endDate + "T23:59:59Z"
	}

	// Создаем запрос на поиск
	searchReq := PlaybackSearchRequest{
		XMLName:              xml.Name{Local: "CMSearchDescription"},
		SearchID:             "0",
		SearchResultPosition: 0,
		MaxResults:           50,
		SearchMode:           "CMSearchMode",
		StartTime:            startDate,
		EndTime:              endDate,
		Channels:             channel,
	}

	// Сериализуем в XML
	reqBody, err := xml.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания XML: %w", err)
	}

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(username, password)

	// Выполняем запрос
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер вернул статус: %s", resp.Status)
	}

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Разбираем XML-ответ
	var searchResponse SearchResponse
	if err := xml.Unmarshal(respBody, &searchResponse); err != nil {
		return nil, fmt.Errorf("ошибка разбора XML: %w", err)
	}

	// Добавляем информацию о канале к каждой записи
	recordings := searchResponse.MatchList.Recordings
	for i := range recordings {
		recordings[i].Channel = channel
	}

	return recordings, nil
}

// GetRTSPURL возвращает URL для RTSP-потока
func GetRTSPURL(ip, username, password, channel string, isPlayback bool) string {
	// Формируем URL для RTSP-потока
	if isPlayback {
		return fmt.Sprintf("rtsp://%s:%s@%s:554/Streaming/tracks/%s",
			username, password, ip, channel)
	}

	return fmt.Sprintf("rtsp://%s:%s@%s:554/Streaming/Channels/%s",
		username, password, ip, channel)
}

// GetPlaybackURL возвращает URL для воспроизведения архива
func GetPlaybackURL(ip, username, password, channel, startTime, endTime string) string {
	// Формируем URL для воспроизведения архива
	return fmt.Sprintf("rtsp://%s:%s@%s:554/Streaming/tracks/%s?starttime=%s&endtime=%s",
		username, password, ip, channel, startTime, endTime)
}
