package config

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
		Port: 8080,
	},
	Hikvision: struct {
		IP       string `json:"ip"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		IP:       "192.168.8.15",
		Username: "admin",
		Password: "oborotni2447",
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
	// TODO: Реализовать загрузку конфигурации из файла
	// 1. Проверить наличие файла конфигурации
	// 2. Если файл есть, загрузить из него
	// 3. Если файла нет, использовать значения по умолчанию

	return &defaultConfig, nil
}
