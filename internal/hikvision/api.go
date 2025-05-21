// internal/hikvision/api.go
package hikvision

import (
	"TeleOko/internal/network"
	"fmt"
)

// SearchRecordings ищет записи в архиве
func SearchRecordings(channel, startDate, endDate string) ([]Recording, error) {
	// Получаем информацию о камере
	camera := network.GetDefaultCamera()

	// Проверяем, получена ли информация о камере
	if camera == nil {
		return nil, fmt.Errorf("не удалось получить информацию о камере")
	}

	// TODO: Реализовать поиск записей
	// 1. Сформировать URL для API
	// 2. Создать XML-запрос
	// 3. Отправить запрос к API Hikvision
	// 4. Обработать результат
	// 5. Вернуть список записей

	// Временная заглушка для тестирования
	recordings := []Recording{
		{
			StartTime: startDate + "T08:00:00Z",
			EndTime:   startDate + "T08:15:00Z",
			Channel:   channel,
		},
		{
			StartTime: startDate + "T12:30:00Z",
			EndTime:   startDate + "T12:45:00Z",
			Channel:   channel,
		},
		{
			StartTime: startDate + "T18:15:00Z",
			EndTime:   startDate + "T18:30:00Z",
			Channel:   channel,
		},
	}

	return recordings, nil
}

// GetRTSPURL возвращает URL для RTSP-потока
func GetRTSPURL(channel string, isPlayback bool) string {
	// Получаем информацию о камере
	camera := network.GetDefaultCamera()

	// Если не удалось получить информацию о камере, используем значения по умолчанию
	if camera == nil {
		return fmt.Sprintf("rtsp://admin:oborotni2447@192.168.8.15:554/Streaming/Channels/%s",
			channel)
	}

	// Формируем URL в зависимости от типа потока
	if isPlayback {
		return fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/tracks/%s",
			camera.Username, camera.Password, camera.IP, camera.Port, channel)
	}

	return fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/Channels/%s",
		camera.Username, camera.Password, camera.IP, camera.Port, channel)
}

// GetPlaybackURL возвращает URL для воспроизведения архива
func GetPlaybackURL(channel, startTime, endTime string) string {
	// Получаем информацию о камере
	camera := network.GetDefaultCamera()

	// Если не удалось получить информацию о камере, используем значения по умолчанию
	if camera == nil {
		return fmt.Sprintf("rtsp://admin:oborotni2447@192.168.8.15:554/Streaming/tracks/%s?starttime=%s&endtime=%s",
			channel, startTime, endTime)
	}

	// Формируем URL для воспроизведения архива
	return fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/tracks/%s?starttime=%s&endtime=%s",
		camera.Username, camera.Password, camera.IP, camera.Port, channel, startTime, endTime)
}
