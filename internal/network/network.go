package network

import (
	"net"
	"time"
)

// GetLocalIP возвращает локальный IP-адрес
func GetLocalIP() (string, error) {
	// Получаем список сетевых интерфейсов
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1", err
	}

	// Находим подходящий IP-адрес
	for _, iface := range interfaces {
		// Пропускаем неактивные интерфейсы и loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Получаем адреса интерфейса
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Ищем IPv4 адрес
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}

			// Пропускаем локальные и специальные адреса
			if ip4[0] == 127 || ip4[0] == 169 {
				continue
			}

			return ip4.String(), nil
		}
	}

	// Если не нашли, возвращаем localhost
	return "127.0.0.1", nil
}

// IPChangeCallback - функция обратного вызова при изменении IP
type IPChangeCallback func(oldIP, newIP string)

// MonitorIP периодически проверяет IP-адрес и вызывает callback при изменении
func MonitorIP(callback IPChangeCallback) {
	currentIP, _ := GetLocalIP()

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
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
