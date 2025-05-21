// internal/config/config.go
package config

import (
	"encoding/json"
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
		IP            string   `json:"ip"`
		Username      string   `json:"username"`
		Password      string   `json:"password"`
		PreferredIPs  []string `json:"preferred_ips"`  // Предпочтительные IP-адреса камер
		AutoDiscovery bool     `json:"auto_discovery"` // Включить автоматическое обнаружение камер
		ScanInterval  int      `json:"scan_interval"`  // Интервал сканирования в минутах
	} `json:"hikvision"`

	Auth struct {
		Enabled  bool   `json:"enabled"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// Глобальная переменная для хранения конфигурации
var GlobalConfig Config

// Значения по умолчанию
var defaultConfig = Config{
	Server: struct {
		Port int `json:"port"`
	}{
		Port: 8080,
	},
	Hikvision: struct {
		IP            string   `json:"ip"`
		Username      string   `json:"username"`
		Password      string   `json:"password"`
		PreferredIPs  []string `json:"preferred_ips"`
		AutoDiscovery bool     `json:"auto_discovery"`
		ScanInterval  int      `json:"scan_interval"`
	}{
		IP:            "192.168.8.15",
		Username:      "admin",
		Password:      "oborotni2447",
		PreferredIPs:  []string{},
		AutoDiscovery: true,
		ScanInterval:  5,
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
}

// Load загружает конфигурацию из файла или использует значения по умолчанию
func Load() (*Config, error) {
	// Пути к файлу конфигурации
	configPaths := []string{
		"config.json",
		filepath.Join("config", "config.json"),
		filepath.Join("..", "config", "config.json"),
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
			return &defaultConfig, nil
		}

		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			log.Printf("Ошибка разбора файла конфигурации: %v", err)
			return &defaultConfig, nil
		}

		GlobalConfig = config
		return &config, nil
	}

	// Если файл не найден, создаем его с настройками по умолчанию
	log.Println("Файл конфигурации не найден, создание файла с настройками по умолчанию")

	// Создаем директорию config, если ее нет
	if err := os.MkdirAll("config", 0755); err != nil {
		log.Printf("Ошибка создания директории config: %v", err)
	}

	// Сериализуем конфигурацию по умолчанию в JSON
	data, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		log.Printf("Ошибка сериализации конфигурации: %v", err)
	} else {
		// Записываем в файл
		if err := ioutil.WriteFile(filepath.Join("config", "config.json"), data, 0644); err != nil {
			log.Printf("Ошибка записи файла конфигурации: %v", err)
		}
	}

	GlobalConfig = defaultConfig
	return &defaultConfig, nil
}

// GetCameraCredentials возвращает учетные данные для камеры
func GetCameraCredentials() (string, string) {
	return GlobalConfig.Hikvision.Username, GlobalConfig.Hikvision.Password
}

// Save сохраняет текущую конфигурацию в файл
func Save() error {
	// Сериализуем конфигурацию в JSON
	data, err := json.MarshalIndent(GlobalConfig, "", "    ")
	if err != nil {
		return err
	}

	// Создаем директорию config, если ее нет
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	// Записываем в файл
	return ioutil.WriteFile(filepath.Join("config", "config.json"), data, 0644)
}

// GetPreferredCameraIPs возвращает список предпочтительных IP-адресов камер
func GetPreferredCameraIPs() []string {
	return GlobalConfig.Hikvision.PreferredIPs
}

// AddPreferredCameraIP добавляет IP-адрес в список предпочтительных
func AddPreferredCameraIP(ip string) error {
	// Проверяем, есть ли уже такой IP в списке
	for _, existingIP := range GlobalConfig.Hikvision.PreferredIPs {
		if existingIP == ip {
			return nil // IP уже есть в списке
		}
	}

	// Добавляем IP в список
	GlobalConfig.Hikvision.PreferredIPs = append(GlobalConfig.Hikvision.PreferredIPs, ip)

	// Сохраняем конфигурацию
	return Save()
}

// IsAutoDiscoveryEnabled возвращает, включено ли автоматическое обнаружение камер
func IsAutoDiscoveryEnabled() bool {
	return GlobalConfig.Hikvision.AutoDiscovery
}

// GetScanInterval возвращает интервал сканирования в минутах
func GetScanInterval() int {
	interval := GlobalConfig.Hikvision.ScanInterval
	if interval < 1 {
		return 5 // Минимальный интервал 1 минута, по умолчанию 5 минут
	}
	return interval
}
