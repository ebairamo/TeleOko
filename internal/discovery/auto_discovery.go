// internal/discovery/auto_discovery.go
package discovery

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CameraInfo содержит информацию об обнаруженной камере
type CameraInfo struct {
	IP       string
	Brand    string
	Model    string
	IsOnline bool
}

// AutoDiscovery выполняет автоматическое обнаружение камер Hikvision в сети
func AutoDiscovery() (*CameraInfo, error) {
	log.Println("🔍 Автоматическое обнаружение камер Hikvision...")

	// Получаем локальную сеть
	localIP, err := getLocalIP()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить локальный IP: %v", err)
	}

	// Определяем подсеть (например, 192.168.8.0/24)
	subnet := getSubnet(localIP)
	log.Printf("🌐 Сканирование подсети: %s", subnet)

	// Сканируем подсеть
	cameras := scanSubnet(subnet)

	if len(cameras) == 0 {
		return nil, fmt.Errorf("камеры Hikvision не найдены в сети")
	}

	// Возвращаем первую найденную камеру
	camera := cameras[0]
	log.Printf("✅ Найдена камера Hikvision: %s", camera.IP)

	return &camera, nil
}

// getLocalIP получает локальный IP адрес
func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// getSubnet определяет подсеть для сканирования
func getSubnet(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) >= 3 {
		return fmt.Sprintf("%s.%s.%s", parts[0], parts[1], parts[2])
	}
	return "192.168.1" // fallback
}

// scanSubnet сканирует подсеть в поисках камер Hikvision
func scanSubnet(subnet string) []CameraInfo {
	var cameras []CameraInfo
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Канал для ограничения количества одновременных соединений
	semaphore := make(chan struct{}, 50)

	// Сканируем IP адреса от 1 до 254
	for i := 1; i <= 254; i++ {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Занимаем слот
			defer func() { <-semaphore }() // Освобождаем слот

			if isHikvisionCamera(ip) {
				camera := CameraInfo{
					IP:       ip,
					Brand:    "Hikvision",
					IsOnline: true,
				}

				// Пытаемся получить модель камеры
				if model := getDeviceModel(ip); model != "" {
					camera.Model = model
				}

				mu.Lock()
				cameras = append(cameras, camera)
				mu.Unlock()

				log.Printf("📹 Найдена камера: %s (%s)", ip, camera.Model)
			}
		}(fmt.Sprintf("%s.%d", subnet, i))
	}

	wg.Wait()
	return cameras
}

// isHikvisionCamera проверяет, является ли устройство камерой Hikvision
func isHikvisionCamera(ip string) bool {
	// Список портов для проверки
	ports := []int{80, 8080, 554}

	for _, port := range ports {
		if checkHikvisionPort(ip, port) {
			return true
		}
	}

	return false
}

// checkHikvisionPort проверяет порт на наличие Hikvision сервиса
func checkHikvisionPort(ip string, port int) bool {
	timeout := 2 * time.Second

	// Используем net.JoinHostPort для корректной работы с IPv6
	address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))

	// Проверяем TCP соединение
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()

	// Для HTTP портов пытаемся получить заголовки
	if port == 80 || port == 8080 {
		return checkHikvisionHTTP(ip, port)
	}

	return true
}

// checkHikvisionHTTP проверяет HTTP заголовки Hikvision
func checkHikvisionHTTP(ip string, port int) bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Используем net.JoinHostPort для правильного форматирования адреса
	url := fmt.Sprintf("http://%s/", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Проверяем заголовки на наличие Hikvision
	server := resp.Header.Get("Server")
	wwwAuth := resp.Header.Get("WWW-Authenticate")

	hikvisionKeywords := []string{
		"hikvision",
		"HIKVISION",
		"Hikvision",
		"DVR",
		"NVR",
		"IVMS",
	}

	headers := strings.ToLower(server + " " + wwwAuth)

	for _, keyword := range hikvisionKeywords {
		if strings.Contains(headers, strings.ToLower(keyword)) {
			return true
		}
	}

	// Проверяем статус 401 (требуется авторизация) - часто признак камеры
	return resp.StatusCode == 401
}

// getDeviceModel пытается получить модель устройства
func getDeviceModel(ip string) string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Пытаемся получить информацию о системе
	urls := []string{
		fmt.Sprintf("http://%s/ISAPI/System/deviceInfo", net.JoinHostPort(ip, "80")),
		fmt.Sprintf("http://%s/ISAPI/System/deviceInfo", net.JoinHostPort(ip, "8080")),
	}

	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		// Базовая авторизация с стандартными логинами
		credentials := [][]string{
			{"admin", "admin"},
			{"admin", "12345"},
			{"admin", ""},
		}

		for _, cred := range credentials {
			req.SetBasicAuth(cred[0], cred[1])

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				// Здесь можно парсить XML ответ для получения модели
				return "Hikvision Camera"
			}
		}
	}

	return "Unknown Model"
}

// FindHikvisionCameras ищет все камеры Hikvision в локальной сети
func FindHikvisionCameras() ([]CameraInfo, error) {
	localIP, err := getLocalIP()
	if err != nil {
		return nil, err
	}

	subnet := getSubnet(localIP)
	cameras := scanSubnet(subnet)

	return cameras, nil
}

// GetBestCamera возвращает лучшую найденную камеру (с наименьшим IP)
func GetBestCamera(cameras []CameraInfo) *CameraInfo {
	if len(cameras) == 0 {
		return nil
	}

	// Сортируем по IP и возвращаем первую
	best := &cameras[0]
	for i := 1; i < len(cameras); i++ {
		if cameras[i].IP < best.IP {
			best = &cameras[i]
		}
	}

	return best
}
