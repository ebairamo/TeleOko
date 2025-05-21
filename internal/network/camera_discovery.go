// internal/network/camera_discovery.go
package network

import (
	"TeleOko/internal/config"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// CameraInfo хранит информацию о найденной камере
type CameraInfo struct {
	IP       string
	Username string
	Password string
	Port     int
	Status   string
	LastSeen time.Time
}

var (
	discoveredCameras = make(map[string]*CameraInfo)
	camerasMutex      sync.RWMutex
)

// FindCameras выполняет сканирование сети для поиска камер
func FindCameras(subnet string, port int, timeout time.Duration) []*CameraInfo {
	if subnet == "" {
		// Определение локальной подсети на основе IP-адреса сервера
		localIP, err := GetLocalIP()
		if err != nil {
			log.Printf("Ошибка определения локального IP: %v", err)
			return nil
		}

		// Преобразование IP в формат подсети (напр. 192.168.1.0/24)
		ip := net.ParseIP(localIP).To4()
		if ip == nil {
			log.Printf("Некорректный IPv4 адрес: %s", localIP)
			return nil
		}

		// Используем первые 3 октета IP-адреса и добавляем маску /24
		subnet = fmt.Sprintf("%d.%d.%d.0/24", ip[0], ip[1], ip[2])
	}

	log.Printf("Сканирование подсети %s для поиска камер", subnet)

	// Разбор CIDR нотации подсети
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Printf("Ошибка разбора подсети: %v", err)
		return nil
	}

	// Получаем список IP-адресов в подсети
	var ips []string
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		// Пропускаем адрес сети и широковещательный адрес
		if !isNetworkOrBroadcast(ip, ipNet) {
			ips = append(ips, ip.String())
		}
	}

	// Создаем каналы для результатов
	results := make(chan *CameraInfo, len(ips))
	var wg sync.WaitGroup

	// Ограничиваем количество одновременных горутин
	semaphore := make(chan struct{}, 20)

	// Проверяем каждый IP-адрес
	for _, ip := range ips {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// Проверяем, является ли этот IP камерой RTSP
			camera := checkRTSPCamera(ip, port, timeout)
			if camera != nil {
				results <- camera
			}
		}(ip)
	}

	// Закрываем канал результатов после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты
	var cameras []*CameraInfo
	for camera := range results {
		cameras = append(cameras, camera)

		// Обновляем глобальный кэш камер
		camerasMutex.Lock()
		discoveredCameras[camera.IP] = camera
		camerasMutex.Unlock()
	}

	log.Printf("Найдено %d камер", len(cameras))
	return cameras
}

// GetCachedCameras возвращает список ранее найденных камер
func GetCachedCameras() []*CameraInfo {
	camerasMutex.RLock()
	defer camerasMutex.RUnlock()

	cameras := make([]*CameraInfo, 0, len(discoveredCameras))
	for _, camera := range discoveredCameras {
		cameras = append(cameras, camera)
	}

	return cameras
}

// GetBestCamera возвращает наиболее подходящую камеру для использования
// (например, последнюю найденную или с заданным IP)
func GetBestCamera(preferredIP string) *CameraInfo {
	camerasMutex.RLock()
	defer camerasMutex.RUnlock()

	// Если указан предпочтительный IP и камера с таким IP существует,
	// возвращаем ее
	if preferredIP != "" {
		if camera, exists := discoveredCameras[preferredIP]; exists {
			return camera
		}
	}

	// Проверяем предпочтительные IP из конфигурации
	preferredIPs := config.GetPreferredCameraIPs()
	for _, ip := range preferredIPs {
		if camera, exists := discoveredCameras[ip]; exists {
			return camera
		}
	}

	// Если нет предпочтительного IP или он не найден,
	// ищем последнюю активную камеру
	var bestCamera *CameraInfo
	var newestTime time.Time

	for _, camera := range discoveredCameras {
		if camera.LastSeen.After(newestTime) {
			newestTime = camera.LastSeen
			bestCamera = camera
		}
	}

	return bestCamera
}

// StartCameraDiscovery запускает постоянный мониторинг камер в сети
func StartCameraDiscovery(interval time.Duration) {
	// Если автообнаружение отключено в конфигурации, не запускаем мониторинг
	if !config.IsAutoDiscoveryEnabled() {
		log.Println("Автоматическое обнаружение камер отключено в конфигурации")
		return
	}

	// Получаем интервал сканирования из конфигурации (если он был изменен)
	configInterval := time.Duration(config.GetScanInterval()) * time.Minute
	if configInterval > 0 {
		interval = configInterval
	}

	log.Printf("Запуск мониторинга камер с интервалом %v", interval)

	go func() {
		for {
			FindCameras("", 554, 500*time.Millisecond)
			time.Sleep(interval)
		}
	}()
}

// GetDefaultCamera возвращает информацию о камере по умолчанию
// Если камера не найдена в кэше, возвращает настройки по умолчанию
func GetDefaultCamera() *CameraInfo {
	// Сначала пытаемся получить камеру из кэша
	camera := GetBestCamera("")

	// Если камера найдена, возвращаем ее
	if camera != nil {
		return camera
	}

	// Если камера не найдена, возвращаем настройки по умолчанию из конфигурации
	username, password := config.GetCameraCredentials()
	defaultIP := config.GlobalConfig.Hikvision.IP // IP из конфигурации

	return &CameraInfo{
		IP:       defaultIP,
		Username: username,
		Password: password,
		Port:     554,
		Status:   "unknown",
		LastSeen: time.Now(),
	}
}

// Вспомогательные функции

// checkRTSPCamera проверяет, отвечает ли указанный IP-адрес на RTSP-запросы
func checkRTSPCamera(ip string, port int, timeout time.Duration) *CameraInfo {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil
	}
	defer conn.Close()

	// Если соединение успешно установлено, проверяем, что это RTSP
	// В реальной реализации здесь можно отправить RTSP OPTIONS запрос
	// и проверить ответ, но для простоты просто считаем успешное
	// TCP-соединение признаком наличия RTSP-сервера

	// Получаем учетные данные для камеры из конфигурации
	username, password := config.GetCameraCredentials()

	return &CameraInfo{
		IP:       ip,
		Username: username,
		Password: password,
		Port:     port,
		Status:   "online",
		LastSeen: time.Now(),
	}
}

// inc увеличивает IP-адрес на 1
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// isNetworkOrBroadcast проверяет, является ли IP-адрес адресом сети или широковещательным
func isNetworkOrBroadcast(ip net.IP, network *net.IPNet) bool {
	// Копируем IP для проверки
	ipCopy := make(net.IP, len(ip))
	copy(ipCopy, ip)

	// Проверяем, является ли это адресом сети
	if ip.Equal(ip.Mask(network.Mask)) {
		return true
	}

	// Проверяем, является ли это широковещательным адресом
	// Устанавливаем все биты хостовой части в 1
	for i := 0; i < len(ipCopy); i++ {
		ipCopy[i] = ipCopy[i] | ^network.Mask[i]
	}

	return ip.Equal(ipCopy)
}
