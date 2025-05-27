// internal/go2rtc/manager.go
package go2rtc

import (
	"TeleOko/internal/config"
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	go2rtcVersion = "1.9.9"
	go2rtcRepo    = "AlexxIT/go2rtc"
)

// Manager управляет процессом go2rtc
type Manager struct {
	process    *exec.Cmd
	configPath string
	binaryPath string
	isRunning  bool
}

var manager *Manager

// NewManager создает новый менеджер go2rtc
func NewManager() *Manager {
	if manager == nil {
		manager = &Manager{
			configPath: "go2rtc.yaml",
			binaryPath: getGo2RTCBinaryPath(),
		}
	}
	return manager
}

// Start запускает go2rtc
func (m *Manager) Start() error {
	// Проверяем, установлен ли go2rtc
	if !m.isBinaryExists() {
		log.Println("go2rtc не найден")
		log.Println("📥 Варианты решения:")
		log.Println("1. Запустите: download-go2rtc.bat (ручная загрузка)")
		log.Println("2. Скачайте go2rtc.exe с https://github.com/AlexxIT/go2rtc/releases/latest")
		log.Println("3. Или отключите go2rtc в config.json: \"enabled\": false")

		// Пробуем автоматическую загрузку один раз
		log.Println("🔄 Попытка автоматической загрузки...")
		if err := m.downloadGo2RTC(); err != nil {
			return fmt.Errorf("автоматическая загрузка go2rtc не удалась: %v\n\n"+
				"🛠️ Решение:\n"+
				"1. Запустите: download-go2rtc.bat\n"+
				"2. Или скачайте go2rtc.exe вручную с GitHub\n"+
				"3. Или отключите go2rtc в config.json", err)
		}
	}

	// Создаем конфигурацию
	if err := m.createConfig(); err != nil {
		return fmt.Errorf("ошибка создания конфигурации go2rtc: %v", err)
	}

	// Запускаем процесс
	if err := m.startProcess(); err != nil {
		return fmt.Errorf("ошибка запуска go2rtc: %v", err)
	}

	log.Println("✅ go2rtc успешно запущен")
	return nil
}

// Stop останавливает go2rtc
func (m *Manager) Stop() error {
	if m.process != nil && m.isRunning {
		if err := m.process.Process.Kill(); err != nil {
			return fmt.Errorf("ошибка остановки go2rtc: %v", err)
		}
		m.isRunning = false
		log.Println("go2rtc остановлен")
	}
	return nil
}

// IsRunning проверяет, запущен ли go2rtc
func (m *Manager) IsRunning() bool {
	return m.isRunning
}

// GetAPIURL возвращает URL для API go2rtc
func (m *Manager) GetAPIURL() string {
	return fmt.Sprintf("http://localhost:%d", config.GetGo2RTCPort())
}

// isBinaryExists проверяет, существует ли бинарник go2rtc
func (m *Manager) isBinaryExists() bool {
	_, err := os.Stat(m.binaryPath)
	return err == nil
}

// downloadGo2RTC загружает go2rtc с GitHub
func (m *Manager) downloadGo2RTC() error {
	// Определяем архитектуру и ОС для правильного URL
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Правильные имена файлов для разных платформ
	var fileName string
	switch osName {
	case "windows":
		if archName == "amd64" {
			fileName = "go2rtc_windows_amd64.zip"
		} else if archName == "386" {
			fileName = "go2rtc_windows_386.zip"
		} else {
			return fmt.Errorf("неподдерживаемая архитектура Windows: %s", archName)
		}
	case "darwin":
		if archName == "amd64" {
			fileName = "go2rtc_darwin_amd64.zip"
		} else if archName == "arm64" {
			fileName = "go2rtc_darwin_arm64.zip"
		} else {
			return fmt.Errorf("неподдерживаемая архитектура macOS: %s", archName)
		}
	case "linux":
		if archName == "amd64" {
			fileName = "go2rtc_linux_amd64"
		} else if archName == "arm64" {
			fileName = "go2rtc_linux_arm64"
		} else if archName == "arm" {
			fileName = "go2rtc_linux_armv6"
		} else {
			return fmt.Errorf("неподдерживаемая архитектура Linux: %s", archName)
		}
	default:
		return fmt.Errorf("неподдерживаемая ОС: %s", osName)
	}

	// Пробуем несколько версий go2rtc для надежности
	versions := []string{go2rtcVersion, "1.9.8", "1.9.7"}

	for _, version := range versions {
		downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
			go2rtcRepo, version, fileName)

		log.Printf("Попытка загрузки go2rtc v%s с %s", version, downloadURL)

		// Загружаем файл
		resp, err := http.Get(downloadURL)
		if err != nil {
			log.Printf("Ошибка запроса: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("HTTP %d для версии %s", resp.StatusCode, version)
			resp.Body.Close()
			continue
		}

		// Успешная загрузка
		defer resp.Body.Close()

		// Для Windows и macOS - это ZIP файлы, для Linux - бинарник
		if strings.HasSuffix(fileName, ".zip") {
			return m.extractZip(resp.Body)
		} else {
			// Для Linux - прямой бинарник
			return m.saveBinary(resp.Body)
		}
	}

	return fmt.Errorf("не удалось загрузить go2rtc ни для одной из версий: %v", versions)
}

// extractZip извлекает ZIP архив (Windows/macOS)
func (m *Manager) extractZip(reader io.Reader) error {
	// Создаем временный файл для ZIP
	tempFile, err := os.CreateTemp("", "go2rtc_*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Копируем содержимое
	_, err = io.Copy(tempFile, reader)
	if err != nil {
		return err
	}
	tempFile.Close()

	// Открываем ZIP архив
	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		return err
	}
	defer zipReader.Close()

	// Ищем исполняемый файл go2rtc
	var binaryName string
	if runtime.GOOS == "windows" {
		binaryName = "go2rtc.exe"
	} else {
		binaryName = "go2rtc"
	}

	for _, file := range zipReader.File {
		if strings.Contains(file.Name, binaryName) {
			// Извлекаем файл
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			// Создаем выходной файл
			outFile, err := os.Create(m.binaryPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Копируем содержимое
			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}

			// Устанавливаем права на выполнение (не Windows)
			if runtime.GOOS != "windows" {
				return os.Chmod(m.binaryPath, 0755)
			}
			return nil
		}
	}

	return fmt.Errorf("исполняемый файл go2rtc не найден в архиве")
}

// saveBinary сохраняет бинарник напрямую (Linux)
func (m *Manager) saveBinary(reader io.Reader) error {
	outFile, err := os.Create(m.binaryPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, reader)
	if err != nil {
		return err
	}

	// Устанавливаем права на выполнение
	return os.Chmod(m.binaryPath, 0755)
}

// createConfig создает конфигурационный файл для go2rtc
func (m *Manager) createConfig() error {
	channels := config.GetChannels()

	configContent := "streams:\n"

	// Добавляем потоки для каждого канала
	for _, channel := range channels {
		configContent += fmt.Sprintf("  %s: %s\n", channel.ID, channel.URL)
	}

	configContent += "\nwebrtc:\n"
	configContent += fmt.Sprintf("  listen: :%d\n", config.GetGo2RTCPort())
	configContent += "  candidates:\n"
	configContent += "    - stun:stun.l.google.com:19302\n"

	configContent += "\napi:\n"
	configContent += fmt.Sprintf("  listen: :%d\n", config.GetGo2RTCPort())

	return os.WriteFile(m.configPath, []byte(configContent), 0644)
}

// startProcess запускает процесс go2rtc
func (m *Manager) startProcess() error {
	// Получаем абсолютный путь к go2rtc
	absPath, err := filepath.Abs(m.binaryPath)
	if err != nil {
		// Если не получается абсолютный путь, используем относительный с ./
		if runtime.GOOS == "windows" {
			absPath = ".\\go2rtc.exe"
		} else {
			absPath = "./go2rtc"
		}
	}

	log.Printf("🚀 Запуск go2rtc: %s", absPath)

	// Создаем команду с абсолютным путем
	m.process = exec.Command(absPath, "-config", m.configPath)

	// Перенаправляем логи
	m.process.Stdout = os.Stdout
	m.process.Stderr = os.Stderr

	if err := m.process.Start(); err != nil {
		return fmt.Errorf("не удалось запустить %s: %v", absPath, err)
	}

	m.isRunning = true

	// Ждем немного, чтобы процесс запустился
	time.Sleep(3 * time.Second)

	// Проверяем, что процесс еще работает
	if m.process.ProcessState != nil && m.process.ProcessState.Exited() {
		m.isRunning = false
		return fmt.Errorf("go2rtc завершился сразу после запуска")
	}

	log.Printf("✅ go2rtc запущен (PID: %d)", m.process.Process.Pid)
	return nil
}

// getGo2RTCBinaryPath возвращает путь к бинарнику go2rtc
func getGo2RTCBinaryPath() string {
	if runtime.GOOS == "windows" {
		return "go2rtc.exe"
	}
	return "go2rtc"
}

// UpdateStreams обновляет потоки в go2rtc
func (m *Manager) UpdateStreams() error {
	if !m.isRunning {
		return fmt.Errorf("go2rtc не запущен")
	}

	// Пока просто логируем - в реальной реализации здесь API вызовы
	channels := config.GetChannels()
	log.Printf("Обновление %d потоков в go2rtc", len(channels))

	return nil
}
