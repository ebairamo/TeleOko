// internal/config/config.go
package config

import (
	"TeleOko/internal/discovery"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Config содержит конфигурацию приложения
type Config struct {
	Server struct {
		Port int `json:"port"`
	} `json:"server"`

	Hikvision struct {
		IP         string `json:"ip"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Port       int    `json:"port"`
		AutoDetect bool   `json:"auto_detect"` // Новое поле для автообнаружения
	} `json:"hikvision"`

	Go2RTC struct {
		Port    int  `json:"port"`
		Enabled bool `json:"enabled"`
	} `json:"go2rtc"`

	Auth struct {
		Enabled  bool   `json:"enabled"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`

	Channels []Channel `json:"channels"`
}

// Channel представляет канал камеры
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Глобальная переменная для хранения конфигурации
var GlobalConfig Config

// Значения по умолчанию
var defaultConfig = Config{
	Server: struct {
		Port int `json:"port"`
	}{
		Port: 8082,
	},
	Hikvision: struct {
		IP         string `json:"ip"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Port       int    `json:"port"`
		AutoDetect bool   `json:"auto_detect"`
	}{
		IP:         "192.168.8.4", // Fallback IP
		Username:   "admin",
		Password:   "oborotni2447",
		Port:       554,
		AutoDetect: true, // Включаем автообнаружение по умолчанию
	},
	Go2RTC: struct {
		Port    int  `json:"port"`
		Enabled bool `json:"enabled"`
	}{
		Port:    1984,
		Enabled: true,
	},
	Auth: struct {
		Enabled  bool   `json:"enabled"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Enabled:  false,
		Username: "admin",
		Password: "password",
	},
	Channels: []Channel{
		{ID: "1", Name: "🎥 Общий план", URL: ""},
		{ID: "101", Name: "📹 Камера 1 (HD)", URL: ""},
		{ID: "102", Name: "📹 Камера 1 (SD)", URL: ""},
		{ID: "201", Name: "📹 Камера 2 (HD)", URL: ""},
		{ID: "202", Name: "📹 Камера 2 (SD)", URL: ""},
		{ID: "301", Name: "📹 Камера 3 (HD)", URL: ""},
		{ID: "302", Name: "📹 Камера 3 (SD)", URL: ""},
		{ID: "401", Name: "📹 Камера 4 (HD)", URL: ""},
		{ID: "402", Name: "📹 Камера 4 (SD)", URL: ""},
		{ID: "501", Name: "📹 Камера 5 (HD)", URL: ""},
		{ID: "502", Name: "📹 Камера 5 (SD)", URL: ""},
		{ID: "601", Name: "📹 Камера 6 (HD)", URL: ""},
		{ID: "602", Name: "📹 Камера 6 (SD)", URL: ""},
		{ID: "701", Name: "📹 Камера 7 (HD)", URL: ""},
		{ID: "702", Name: "📹 Камера 7 (SD)", URL: ""},
		{ID: "801", Name: "📹 Камера 8 (HD)", URL: ""},
		{ID: "802", Name: "📹 Камера 8 (SD)", URL: ""},
		{ID: "901", Name: "📹 Камера 9 (HD)", URL: ""},
		{ID: "902", Name: "📹 Камера 9 (SD)", URL: ""},
		{ID: "1001", Name: "📹 Камера 10 (HD)", URL: ""},
		{ID: "1002", Name: "📹 Камера 10 (SD)", URL: ""},
		{ID: "1101", Name: "📹 Камера 11 (HD)", URL: ""},
		{ID: "1102", Name: "📹 Камера 11 (SD)", URL: ""},
		{ID: "1201", Name: "📹 Камера 12 (HD)", URL: ""},
		{ID: "1202", Name: "📹 Камера 12 (SD)", URL: ""},
		{ID: "1301", Name: "📹 Камера 13 (HD)", URL: ""},
		{ID: "1302", Name: "📹 Камера 13 (SD)", URL: ""},
		{ID: "1401", Name: "📹 Камера 14 (HD)", URL: ""},
		{ID: "1402", Name: "📹 Камера 14 (SD)", URL: ""},
		{ID: "1501", Name: "📹 Камера 15 (HD)", URL: ""},
		{ID: "1502", Name: "📹 Камера 15 (SD)", URL: ""},
		{ID: "1601", Name: "📹 Камера 16 (HD)", URL: ""},
		{ID: "1602", Name: "📹 Камера 16 (SD)", URL: ""},
	},
}

// Load загружает конфигурацию из файла или использует значения по умолчанию
func Load() (*Config, error) {
	// Пути к файлу конфигурации
	configPaths := []string{
		"config.json",
		filepath.Join("config", "config.json"),
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	// Если файл конфигурации найден, загружаем его
	if configFile != "" {
		log.Printf("📂 Загрузка конфигурации из файла: %s", configFile)
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("⚠️ Ошибка чтения файла конфигурации: %v", err)
			GlobalConfig = defaultConfig
		} else {
			var config Config
			if err := json.Unmarshal(data, &config); err != nil {
				log.Printf("⚠️ Ошибка разбора файла конфигурации: %v", err)
				GlobalConfig = defaultConfig
			} else {
				GlobalConfig = config
			}
		}
	} else {
		// Если файл не найден, создаем его с настройками по умолчанию
		log.Println("📝 Файл конфигурации не найден, создание файла с настройками по умолчанию")
		GlobalConfig = defaultConfig
	}

	// Автообнаружение камер (если включено)
	if GlobalConfig.Hikvision.AutoDetect {
		if err := autoDetectCamera(); err != nil {
			log.Printf("⚠️ Автообнаружение не удалось: %v", err)
			log.Printf("📍 Используется IP из конфигурации: %s", GlobalConfig.Hikvision.IP)
		}
	}

	// Генерируем URL каналов
	generateChannelURLs()

	// Сохраняем обновленную конфигурацию
	if err := Save(); err != nil {
		log.Printf("⚠️ Ошибка сохранения конфигурации: %v", err)
	}

	return &GlobalConfig, nil
}

// autoDetectCamera выполняет автоматическое обнаружение камеры
func autoDetectCamera() error {
	log.Println("🔍 Автообнаружение камер Hikvision...")

	// Ищем камеры в сети
	cameras, err := discovery.FindHikvisionCameras()
	if err != nil {
		return fmt.Errorf("ошибка поиска камер: %v", err)
	}

	if len(cameras) == 0 {
		return fmt.Errorf("камеры Hikvision не найдены в локальной сети")
	}

	// Выбираем лучшую камеру
	bestCamera := discovery.GetBestCamera(cameras)
	if bestCamera == nil {
		return fmt.Errorf("не удалось выбрать камеру")
	}

	// Обновляем конфигурацию, если IP изменился
	if bestCamera.IP != GlobalConfig.Hikvision.IP {
		oldIP := GlobalConfig.Hikvision.IP
		GlobalConfig.Hikvision.IP = bestCamera.IP
		log.Printf("🔄 IP камеры обновлен: %s → %s", oldIP, bestCamera.IP)
		log.Printf("📹 Обнаружена камера: %s (%s)", bestCamera.IP, bestCamera.Model)

		// Отмечаем, что нужно обновить go2rtc конфигурацию
		updateGo2RTCConfig = true
	} else {
		log.Printf("✅ IP камеры актуален: %s", GlobalConfig.Hikvision.IP)
	}

	// Показываем все найденные камеры
	if len(cameras) > 1 {
		log.Printf("📋 Найдено камер: %d", len(cameras))
		for i, camera := range cameras {
			status := ""
			if camera.IP == GlobalConfig.Hikvision.IP {
				status = " (выбрана)"
			}
			log.Printf("  %d. %s - %s%s", i+1, camera.IP, camera.Model, status)
		}
	}

	return nil
}

// updateGo2RTCConfig флаг для обновления go2rtc конфигурации
var updateGo2RTCConfig = false

// ShouldUpdateGo2RTC возвращает true, если нужно обновить go2rtc конфигурацию
func ShouldUpdateGo2RTC() bool {
	return updateGo2RTCConfig
}

// generateChannelURLs генерирует RTSP URL для каналов
func generateChannelURLs() {
	baseURL := fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/Channels/",
		GlobalConfig.Hikvision.Username,
		GlobalConfig.Hikvision.Password,
		GlobalConfig.Hikvision.IP,
		GlobalConfig.Hikvision.Port)

	for i := range GlobalConfig.Channels {
		// Всегда перегенерируем URL на основе текущих настроек
		GlobalConfig.Channels[i].URL = baseURL + GlobalConfig.Channels[i].ID
	}

	log.Printf("✅ Обновлены RTSP URL для %d каналов с IP: %s",
		len(GlobalConfig.Channels), GlobalConfig.Hikvision.IP)
}

// Save сохраняет текущую конфигурацию в файл
func Save() error {
	// Сериализуем конфигурацию в JSON
	data, err := json.MarshalIndent(GlobalConfig, "", "    ")
	if err != nil {
		return err
	}

	// Записываем в файл
	return ioutil.WriteFile("config.json", data, 0644)
}

// Остальные функции остаются без изменений...

// GetChannels возвращает список каналов
func GetChannels() []Channel {
	return GlobalConfig.Channels
}

// GetChannelByID возвращает канал по ID
func GetChannelByID(id string) *Channel {
	for i := range GlobalConfig.Channels {
		if GlobalConfig.Channels[i].ID == id {
			return &GlobalConfig.Channels[i]
		}
	}
	return nil
}

// GetHikvisionCredentials возвращает учетные данные для Hikvision
func GetHikvisionCredentials() (string, string, string, int) {
	return GlobalConfig.Hikvision.IP,
		GlobalConfig.Hikvision.Username,
		GlobalConfig.Hikvision.Password,
		GlobalConfig.Hikvision.Port
}

// GetGo2RTCPort возвращает порт go2rtc
func GetGo2RTCPort() int {
	return GlobalConfig.Go2RTC.Port
}

// IsGo2RTCEnabled проверяет, включен ли go2rtc
func IsGo2RTCEnabled() bool {
	return GlobalConfig.Go2RTC.Enabled
}

// SetHikvisionIP обновляет IP камеры и перегенерирует все URL
func SetHikvisionIP(newIP string) {
	oldIP := GlobalConfig.Hikvision.IP
	GlobalConfig.Hikvision.IP = newIP
	generateChannelURLs()

	log.Printf("🔄 IP камеры изменен: %s -> %s", oldIP, newIP)
	log.Printf("✅ Обновлены URL для всех %d каналов", len(GlobalConfig.Channels))
}

// GetChannelURL возвращает актуальный RTSP URL для канала (динамически)
func GetChannelURL(channelID string) string {
	return fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/Channels/%s",
		GlobalConfig.Hikvision.Username,
		GlobalConfig.Hikvision.Password,
		GlobalConfig.Hikvision.IP,
		GlobalConfig.Hikvision.Port,
		channelID)
}
