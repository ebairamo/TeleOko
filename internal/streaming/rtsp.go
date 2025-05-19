package streaming

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/deepch/vdk/format/rtsp"
)

// RTSPClient представляет клиент для подключения к RTSP-потоку
type RTSPClient struct {
	URL      string
	Username string
	Password string
	Session  *rtsp.Client
	Active   bool
	Mutex    sync.Mutex
}

// StreamPacket представляет пакет данных из потока
type StreamPacket struct {
	Data       []byte
	IsKeyFrame bool
	Duration   time.Duration
	Time       time.Duration
}

// Кэш для хранения подключений к RTSP
var (
	rtspConnections = make(map[string]*RTSPClient)
	rtspMutex       = &sync.Mutex{}
)

// NewRTSPClient создает новый RTSP-клиент
func NewRTSPClient(url, username, password string) (*RTSPClient, error) {
	log.Printf("Создание нового RTSP-клиента для %s", url)

	// Проверяем наличие учетных данных в URL
	if username != "" && password != "" {
		// Проверяем, содержит ли URL уже учетные данные
		if !strings.Contains(url, "@") {
			// Разбиваем URL на составные части
			urlParts := strings.SplitN(url, "://", 2)
			if len(urlParts) != 2 {
				return nil, fmt.Errorf("неверный формат URL: %s", url)
			}

			// Добавляем учетные данные
			url = fmt.Sprintf("%s://%s:%s@%s", urlParts[0], username, password, urlParts[1])
		}
	}

	// Создаем клиент RTSP
	session, err := rtsp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к RTSP: %w", err)
	}

	// Создаем экземпляр клиента
	rtspClient := &RTSPClient{
		URL:      url,
		Username: username,
		Password: password,
		Session:  session,
		Active:   true,
	}

	return rtspClient, nil
}

// GetRTSPConnection получает или создает RTSP-соединение
func GetRTSPConnection(url, username, password string) (*RTSPClient, error) {
	rtspMutex.Lock()
	defer rtspMutex.Unlock()

	// Формируем ключ для кэша
	cacheKey := url

	// Проверяем наличие соединения в кэше
	if client, ok := rtspConnections[cacheKey]; ok {
		// Проверяем состояние соединения
		client.Mutex.Lock()
		defer client.Mutex.Unlock()

		if client.Active && client.Session != nil {
			log.Printf("Использование существующего RTSP-соединения для %s", url)
			return client, nil
		}

		// Если соединение неактивно или закрыто, пересоздаем его
		log.Printf("Переподключение к RTSP %s", url)
		if client.Session != nil {
			client.Session.Close()
		}
	}

	// Создаем новое соединение
	client, err := NewRTSPClient(url, username, password)
	if err != nil {
		return nil, err
	}

	// Добавляем в кэш
	rtspConnections[cacheKey] = client
	return client, nil
}

// CloseConnection закрывает соединение
func (c *RTSPClient) CloseConnection() {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.Session != nil {
		c.Session.Close()
		c.Session = nil
	}
	c.Active = false
}

// CloseAllConnections закрывает все соединения
func CloseAllConnections() {
	rtspMutex.Lock()
	defer rtspMutex.Unlock()

	for _, client := range rtspConnections {
		client.CloseConnection()
	}

	// Очистка кэша соединений
	rtspConnections = make(map[string]*RTSPClient)
}

// GetPacket получает пакет из RTSP-потока
func (c *RTSPClient) GetPacket() (*StreamPacket, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	if c.Session == nil || !c.Active {
		return nil, fmt.Errorf("клиент RTSP неактивен")
	}

	// Читаем пакет
	pkt, err := c.Session.ReadPacket()
	if err != nil {
		c.Active = false
		return nil, fmt.Errorf("ошибка чтения пакета: %w", err)
	}

	// Проверяем, что это видеопакет
	if pkt.IsKeyFrame || pkt.Idx == 0 { // обычно видео имеет индекс 0
		// Создаем пакет
		packet := &StreamPacket{
			Data:       pkt.Data,
			IsKeyFrame: pkt.IsKeyFrame,
			Duration:   pkt.Duration,
			Time:       pkt.Time,
		}

		return packet, nil
	}

	// Если это не видеопакет, возвращаем ошибку
	return nil, fmt.Errorf("не видеопакет")
}
