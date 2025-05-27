// internal/network/discovery.go
package network

import (
	"TeleOko/internal/config"
	"fmt"
	"log"
	"net"
	"time"
)

// GetLocalIP возвращает локальный IP-адрес
func GetLocalIP() (string, error) {
	// Создаем UDP соединение для определения локального IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Если не удалось, пробуем через интерфейсы
		return getLocalIPFromInterfaces()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// getLocalIPFromInterfaces получает IP через сетевые интерфейсы
func getLocalIPFromInterfaces() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "127.0.0.1", nil
}

// TestCameraConnection проверяет доступность камеры
func TestCameraConnection() error {
	ip, username, password, port := config.GetHikvisionCredentials()

	// Проверяем TCP подключение к RTSP порту
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к %s: %v", address, err)
	}
	defer conn.Close()

	log.Printf("✅ Камера %s:%d доступна (пользователь: %s)", ip, port, username)

	// Для полной проверки можно добавить RTSP OPTIONS запрос
	_ = password // используем переменную чтобы избежать предупреждения

	return nil
}

// GetCameraRTSPURL формирует RTSP URL для канала
func GetCameraRTSPURL(channelID string) string {
	ip, username, password, port := config.GetHikvisionCredentials()
	return fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/Channels/%s",
		username, password, ip, port, channelID)
}
