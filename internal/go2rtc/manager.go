package go2rtc

import (
	"TeleOko/internal/config"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
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
		return fmt.Errorf("go2rtc не найден по пути: %s", m.binaryPath)
	}

	// Проверяем конфигурацию
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("конфигурация go2rtc не найдена: %s", m.configPath)
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
		if runtime.GOOS == "windows" {
			// Для Windows используем taskkill
			cmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", m.process.Process.Pid))
			cmd.Run()
		} else {
			// Для Unix-систем
			m.process.Process.Signal(os.Interrupt)
			time.Sleep(2 * time.Second)
			if m.process.ProcessState == nil {
				m.process.Process.Kill()
			}
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

// startProcess запускает процесс go2rtc
func (m *Manager) startProcess() error {
	// Получаем абсолютный путь к go2rtc
	absPath, err := filepath.Abs(m.binaryPath)
	if err != nil {
		absPath = m.binaryPath
	}

	log.Printf("🚀 Запуск go2rtc: %s", absPath)

	// Создаем команду
	m.process = exec.Command(absPath, "-config", m.configPath)

	// Устанавливаем рабочую директорию
	m.process.Dir = "."

	// Перенаправляем вывод
	m.process.Stdout = os.Stdout
	m.process.Stderr = os.Stderr

	// Запускаем процесс
	if err := m.process.Start(); err != nil {
		return fmt.Errorf("не удалось запустить %s: %v", absPath, err)
	}

	m.isRunning = true

	// Ждем немного, чтобы процесс запустился
	time.Sleep(2 * time.Second)

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
	return "./go2rtc"
}

// CheckHealth проверяет здоровье go2rtc
func (m *Manager) CheckHealth() bool {
	if !m.isRunning {
		return false
	}

	// Проверяем доступность API
	url := fmt.Sprintf("http://localhost:%d/api", config.GetGo2RTCPort())
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
