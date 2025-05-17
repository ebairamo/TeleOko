package hikvision

import (
	"fmt"
)

// SearchRecordings ищет записи в архиве
func SearchRecordings(ip, username, password, channel, startDate, endDate string) ([]Recording, error) {
	// TODO: Реализовать поиск записей
	// 1. Сформировать URL для API
	// 2. Создать XML-запрос
	// 3. Отправить запрос к API Hikvision
	// 4. Обработать результат
	// 5. Вернуть список записей

	return []Recording{}, nil
}

// GetRTSPURL возвращает URL для RTSP-потока
func GetRTSPURL(ip, username, password, channel string, isPlayback bool) string {
	// TODO: Сформировать URL для RTSP-потока
	// 1. Для прямого эфира: rtsp://user:pass@ip:554/Streaming/Channels/{channel}
	// 2. Для архива: rtsp://user:pass@ip:554/Streaming/tracks/{channel}?starttime=...

	if isPlayback {
		return fmt.Sprintf("rtsp://%s:%s@%s:554/Streaming/tracks/%s",
			username, password, ip, channel)
	}

	return fmt.Sprintf("rtsp://%s:%s@%s:554/Streaming/Channels/%s",
		username, password, ip, channel)
}

// GetPlaybackURL возвращает URL для воспроизведения архива
func GetPlaybackURL(ip, username, password, channel, startTime, endTime string) string {
	// TODO: Сформировать URL для воспроизведения архива
	// rtsp://user:pass@ip:554/Streaming/tracks/{channel}?starttime={startTime}&endtime={endTime}

	return fmt.Sprintf("rtsp://%s:%s@%s:554/Streaming/tracks/%s?starttime=%s&endtime=%s",
		username, password, ip, channel, startTime, endTime)
}
