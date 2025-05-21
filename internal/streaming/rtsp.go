package streaming

import (
	"log"
	"sync"
	"time"

	"github.com/deepch/vdk/format/rtsp"
)

// RTSPClient представляет клиент для подключения к RTSP-потоку
type RTSPClient struct {
	URL        string
	Username   string
	Password   string
	Connection *rtsp.Client
	Connected  bool
	LastError  error
	LastAccess time.Time
}

// Кэш для хранения подключений к RTSP
var (
	rtspConnections = make(map[string]*RTSPClient)
	rtspMutex       = &sync.Mutex{}
)

// NewRTSPClient создает новый RTSP-клиент
func NewRTSPClient(url, username, password string) (*RTSPClient, error) {
	// Создаем клиент RTSP
	client, err := rtsp.Dial(url)
	if err != nil {
		return &RTSPClient{
			URL:        url,
			Username:   username,
			Password:   password,
			Connected:  false,
			LastError:  err,
			LastAccess: time.Now(),
		}, err
	}

	// Если для подключения требуется аутентификация, не можем явно установить учетные данные
	// т.к. метод SetCredentials отсутствует в библиотеке vdk/rtsp
	// Вместо этого учетные данные должны быть частью URL

	// Создаем и возвращаем клиент
	rtspClient := &RTSPClient{
		URL:        url,
		Username:   username,
		Password:   password,
		Connection: client,
		Connected:  true,
		LastAccess: time.Now(),
	}

	return rtspClient, nil
}

// GetRTSPConnection получает или создает RTSP-соединение
func GetRTSPConnection(url, username, password string) (*RTSPClient, error) {
	rtspMutex.Lock()
	defer rtspMutex.Unlock()

	// Проверяем, есть ли соединение в кэше
	if client, ok := rtspConnections[url]; ok {
		// Обновляем время последнего доступа
		client.LastAccess = time.Now()

		// Проверяем, не закрыто ли соединение
		if client.Connected && client.Connection != nil {
			return client, nil
		}

		// Если соединение закрыто, пытаемся пересоздать
		newClient, err := NewRTSPClient(url, username, password)
		if err != nil {
			return client, err // Возвращаем старый клиент и ошибку
		}

		// Заменяем клиент в кэше и возвращаем новый
		rtspConnections[url] = newClient
		return newClient, nil
	}

	// Если соединения нет в кэше, создаем новое
	client, err := NewRTSPClient(url, username, password)
	if err != nil {
		// Все равно сохраняем клиент в кэше, даже если соединение не удалось
		// В следующий раз мы попробуем пересоздать
		rtspConnections[url] = client
		return client, err
	}

	// Добавляем клиент в кэш
	rtspConnections[url] = client
	return client, nil
}

// CloseRTSPConnection закрывает RTSP-соединение
func CloseRTSPConnection(url string) {
	rtspMutex.Lock()
	defer rtspMutex.Unlock()

	if client, ok := rtspConnections[url]; ok {
		if client.Connection != nil {
			client.Connection.Close()
		}
		client.Connected = false
		delete(rtspConnections, url)
	}
}

// CloseAllConnections закрывает все соединения
func CloseAllConnections() {
	rtspMutex.Lock()
	defer rtspMutex.Unlock()

	for url, client := range rtspConnections {
		if client.Connection != nil {
			client.Connection.Close()
		}
		delete(rtspConnections, url)
	}

	log.Println("Все RTSP-соединения закрыты")
}

// CleanupOldConnections закрывает устаревшие соединения
func CleanupOldConnections(maxAge time.Duration) {
	rtspMutex.Lock()
	defer rtspMutex.Unlock()

	now := time.Now()
	for url, client := range rtspConnections {
		// Если соединение не использовалось дольше maxAge, закрываем его
		if now.Sub(client.LastAccess) > maxAge {
			if client.Connection != nil {
				client.Connection.Close()
			}
			delete(rtspConnections, url)
			log.Printf("Закрыто устаревшее RTSP-соединение: %s", url)
		}
	}
}

// StartRTSPCleanupRoutine запускает периодическую очистку устаревших соединений
func StartRTSPCleanupRoutine(interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			CleanupOldConnections(maxAge)
		}
	}()

	log.Printf("Запущена периодическая очистка RTSP-соединений (интервал: %v, максимальный возраст: %v)", interval, maxAge)
}
