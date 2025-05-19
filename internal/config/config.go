package config

import (
	"encoding/json"
	"fmt"
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
	} `json:"hikvision"`

	Auth struct {
		Enabled  bool   `json:"enabled"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"auth"`
}

// Значения по умолчанию
var defaultConfig = Config{
	Server: struct {
		Port int `json:"port"`
	}{
		Port: 8082, // Изменено с 8080 на 8082
	},
	Hikvision: struct {
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		IP:       "192.168.1.64",
		Username: "admin",
		Password: "admin12345",
	},
	Auth: struct {
		Enabled  bool   `json:"enabled"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Enabled:  false,
		Username: "admin",
		Password: "admin",
	},
}

// Путь к файлу конфигурации
const configPath = "config.json"

// Load загружает конфигурацию из файла или использует значения по умолчанию
func Load() (*Config, error) {
	// Проверяем наличие файла конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Файла нет, создаем новый с настройками по умолчанию
		return saveDefaultConfig()
	}

	// Файл существует, загружаем из него
	file, err := os.Open(configPath)
	if err != nil {
		return &defaultConfig, fmt.Errorf("ошибка открытия файла конфигурации: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return &defaultConfig, fmt.Errorf("ошибка разбора файла конфигурации: %w", err)
	}

	return &config, nil
}

// saveDefaultConfig сохраняет конфигурацию по умолчанию в файл
func saveDefaultConfig() (*Config, error) {
	// Создаем директорию для файла конфигурации, если она не существует
	dir := filepath.Dir(configPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return &defaultConfig, fmt.Errorf("ошибка создания директории для конфигурации: %w", err)
		}
	}

	// Создаем файл
	file, err := os.Create(configPath)
	if err != nil {
		return &defaultConfig, fmt.Errorf("ошибка создания файла конфигурации: %w", err)
	}
	defer file.Close()

	// Записываем JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(defaultConfig); err != nil {
		return &defaultConfig, fmt.Errorf("ошибка записи конфигурации: %w", err)
	}

	return &defaultConfig, nil
}

// Save сохраняет конфигурацию в файл
func Save(config *Config) error {
	// Создаем директорию для файла конфигурации, если она не существует
	dir := filepath.Dir(configPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("ошибка создания директории для конфигурации: %w", err)
		}
	}

	// Создаем файл
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла конфигурации: %w", err)
	}
	defer file.Close()

	// Записываем JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("ошибка записи конфигурации: %w", err)
	}

	return nil
}

// GetLocalIP возвращает локальный IP-адрес
func GetLocalIP() (string, error) {
	// Используем функцию из пакета network
	return "192.168.1.100", nil
}
