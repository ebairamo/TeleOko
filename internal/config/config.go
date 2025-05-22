// internal/config/config.go
package config

import (
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
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
		Port     int    `json:"port"`
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
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
		Port     int    `json:"port"`
	}{
		IP:       "192.168.8.5",
		Username: "admin",
		Password: "oborotni2447",
		Port:     554,
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
		{ID: "1", Name: "Общий план", URL: ""},
		{ID: "201", Name: "Камера 1 (HD)", URL: ""},
		{ID: "202", Name: "Камера 1 (SD)", URL: ""},
		{ID: "301", Name: "Камера 2 (HD)", URL: ""},
		{ID: "302", Name: "Камера 2 (SD)", URL: ""},
		{ID: "401", Name: "Камера 3 (HD)", URL: ""},
		{ID: "402", Name: "Камера 3 (SD)", URL: ""},
		{ID: "501", Name: "Камера 4 (HD)", URL: ""},
		{ID: "502", Name: "Камера 4 (SD)", URL: ""},
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
		log.Printf("Загрузка конфигурации из файла: %s", configFile)
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Printf("Ошибка чтения файла конфигурации: %v", err)
			GlobalConfig = defaultConfig
			generateChannelURLs()
			return &GlobalConfig, nil
		}

		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			log.Printf("Ошибка разбора файла конфигурации: %v", err)
			GlobalConfig = defaultConfig
			generateChannelURLs()
			return &GlobalConfig, nil
		}

		GlobalConfig = config
		generateChannelURLs()
		return &config, nil
	}

	// Если файл не найден, создаем его с настройками по умолчанию
	log.Println("Файл конфигурации не найден, создание файла с настройками по умолчанию")

	GlobalConfig = defaultConfig
	generateChannelURLs()

	// Сохраняем конфигурацию
	if err := Save(); err != nil {
		log.Printf("Ошибка сохранения конфигурации: %v", err)
	}

	return &GlobalConfig, nil
}

// generateChannelURLs генерирует RTSP URL для каналов
func generateChannelURLs() {
	baseURL := fmt.Sprintf("rtsp://%s:%s@%s:%d/Streaming/Channels/",
		GlobalConfig.Hikvision.Username,
		GlobalConfig.Hikvision.Password,
		GlobalConfig.Hikvision.IP,
		GlobalConfig.Hikvision.Port)

	for i := range GlobalConfig.Channels {
		if GlobalConfig.Channels[i].URL == "" {
			GlobalConfig.Channels[i].URL = baseURL + GlobalConfig.Channels[i].ID
		}
	}
}

// Save сохраняет текущую конфигурацию в файл
func Save() error {
	// Создаем директорию config, если ее нет
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	// Сериализуем конфигурацию в JSON
	data, err := json.MarshalIndent(GlobalConfig, "", "    ")
	if err != nil {
		return err
	}

	// Записываем в файл
	return ioutil.WriteFile("config.json", data, 0644)
}

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
