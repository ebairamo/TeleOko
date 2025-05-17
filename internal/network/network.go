package network

import (
	"time"
)

// GetLocalIP возвращает локальный IP-адрес
func GetLocalIP() (string, error) {
	// TODO: Реализовать определение локального IP-адреса
	// 1. Получить список сетевых интерфейсов
	// 2. Найти подходящий IP-адрес (не локальный)
	// 3. Вернуть найденный IP-адрес

	// Временная заглушка для примера
	return "192.168.1.100", nil
}

// IPChangeCallback - функция обратного вызова при изменении IP
type IPChangeCallback func(oldIP, newIP string)

// MonitorIP периодически проверяет IP-адрес и вызывает callback при изменении
func MonitorIP(callback IPChangeCallback) {
	// TODO: Реализовать периодическую проверку IP-адреса
	// 1. Запомнить текущий IP-адрес
	// 2. Периодически проверять новый IP-адрес
	// 3. Если IP изменился, вызвать callback

	currentIP, _ := GetLocalIP()

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// TODO: Получить новый IP и сравнить с текущим
		newIP, err := GetLocalIP()
		if err != nil {
			continue
		}

		if newIP != currentIP {
			callback(currentIP, newIP)
			currentIP = newIP
		}
	}
}
