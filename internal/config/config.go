package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

// Load загружает конфигурацию из файла
func Load() (*Config, error) {
	configFile := "config.json"

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %v", err)
	}

	if err := json.Unmarshal(data, &GlobalConfig); err != nil {
		return nil, fmt.Errorf("ошибка разбора конфигурации: %v", err)
	}

	log.Printf("✅ Конфигурация загружена: %d каналов", len(GlobalConfig.Channels))
	return &GlobalConfig, nil
}

// Save сохраняет текущую конфигурацию в файл
func Save() error {
	data, err := json.MarshalIndent(GlobalConfig, "", "    ")
	if err != nil {
		return err
	}

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
