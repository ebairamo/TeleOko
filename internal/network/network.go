// internal/network/network.go
package network

import (
	"net"
	"strings"
	"time"
)

// GetLocalIP возвращает локальный IP-адрес
func GetLocalIP() (string, error) {
	// Список интерфейсов, которые следует игнорировать
	ignoredInterfaces := []string{"lo", "docker", "veth", "br-", "vEthernet", "VMware", "VirtualBox"}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var candidateIPs []string

	for _, iface := range interfaces {
		// Пропускаем интерфейсы, которые не активны
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Пропускаем интерфейсы из списка игнорируемых
		skip := false
		for _, ignored := range ignoredInterfaces {
			if strings.Contains(strings.ToLower(iface.Name), strings.ToLower(ignored)) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Получаем адреса интерфейса
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Пропускаем не CIDR адреса
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Пропускаем loopback и IPv6 адреса
			ip := ipNet.IP.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Приоритет для IP адресов домашних сетей
			ipStr := ip.String()
			if strings.HasPrefix(ipStr, "192.168.") || strings.HasPrefix(ipStr, "10.") {
				candidateIPs = append(candidateIPs, ipStr)
			}
		}
	}

	// Если нашли подходящие IP, возвращаем первый
	if len(candidateIPs) > 0 {
		return candidateIPs[0], nil
	}

	// Если не нашли подходящие, используем другой метод
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1", nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// IPChangeCallback - функция обратного вызова при изменении IP
type IPChangeCallback func(oldIP, newIP string)

// MonitorIP периодически проверяет IP-адрес и вызывает callback при изменении
func MonitorIP(callback IPChangeCallback) {
	// Запоминаем текущий IP-адрес
	currentIP, err := GetLocalIP()
	if err != nil {
		currentIP = "127.0.0.1"
	}

	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Получаем новый IP и сравниваем с текущим
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
