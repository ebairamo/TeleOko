package hls

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// StreamManager управляет HLS потоками
type StreamManager struct {
	streams    map[string]*Stream
	streamsMux sync.RWMutex
	outputDir  string
}

// Stream представляет один HLS поток
type Stream struct {
	ChannelID   string
	RTSPURL     string
	OutputPath  string
	PlaylistURL string
	cmd         *exec.Cmd
	Active      bool
	StartTime   time.Time
	mu          sync.Mutex
}

// NewStreamManager создает новый менеджер потоков
func NewStreamManager(outputDir string) *StreamManager {
	// Создаем директорию для потоков
	os.MkdirAll(outputDir, 0755)

	// Очищаем старые файлы при запуске
	files, _ := filepath.Glob(filepath.Join(outputDir, "*.m3u8"))
	for _, f := range files {
		os.Remove(f)
	}
	files, _ = filepath.Glob(filepath.Join(outputDir, "*.ts"))
	for _, f := range files {
		os.Remove(f)
	}

	return &StreamManager{
		streams:   make(map[string]*Stream),
		outputDir: outputDir,
	}
}

// StartStream запускает HLS поток для канала
func (sm *StreamManager) StartStream(channelID, rtspURL string) (*Stream, error) {
	sm.streamsMux.Lock()
	defer sm.streamsMux.Unlock()

	// Проверяем, не запущен ли уже поток
	if stream, exists := sm.streams[channelID]; exists && stream.Active {
		log.Printf("📺 Поток для канала %s уже активен", channelID)
		return stream, nil
	}

	// Создаем пути для файлов
	playlistName := fmt.Sprintf("channel_%s.m3u8", channelID)
	playlistPath := filepath.Join(sm.outputDir, playlistName)

	// Создаем новый поток
	stream := &Stream{
		ChannelID:   channelID,
		RTSPURL:     rtspURL,
		OutputPath:  playlistPath,
		PlaylistURL: fmt.Sprintf("/streams/%s", playlistName),
		Active:      false,
		StartTime:   time.Now(),
	}

	// Запускаем FFmpeg для конвертации RTSP в HLS
	if err := stream.start(); err != nil {
		return nil, err
	}

	// Сохраняем поток
	sm.streams[channelID] = stream

	// Запускаем очистку старых сегментов
	go sm.cleanupOldSegments(channelID)

	return stream, nil
}

// StopStream останавливает HLS поток
func (sm *StreamManager) StopStream(channelID string) error {
	sm.streamsMux.Lock()
	defer sm.streamsMux.Unlock()

	stream, exists := sm.streams[channelID]
	if !exists {
		return fmt.Errorf("поток для канала %s не найден", channelID)
	}

	// Останавливаем поток
	stream.stop()

	// Удаляем из карты
	delete(sm.streams, channelID)

	// Удаляем файлы
	go sm.cleanupStreamFiles(channelID)

	return nil
}

// GetStream возвращает информацию о потоке
func (sm *StreamManager) GetStream(channelID string) (*Stream, bool) {
	sm.streamsMux.RLock()
	defer sm.streamsMux.RUnlock()

	stream, exists := sm.streams[channelID]
	return stream, exists
}

// GetActiveStreams возвращает список активных потоков
func (sm *StreamManager) GetActiveStreams() []*Stream {
	sm.streamsMux.RLock()
	defer sm.streamsMux.RUnlock()

	streams := make([]*Stream, 0, len(sm.streams))
	for _, stream := range sm.streams {
		if stream.Active {
			streams = append(streams, stream)
		}
	}

	return streams
}

func (s *Stream) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем доступность ffmpeg
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg не найден: %v", err)
	}

	// Параметры FFmpeg для HLS (обновленные)
	args := []string{
		"-rtsp_transport", "tcp", // Используем TCP для RTSP
		"-i", s.RTSPURL, // Входной RTSP поток
		"-c:v", "copy", // Копируем видео без перекодирования
		"-c:a", "copy", // Копируем аудио без перекодирования
		"-f", "hls", // Формат HLS
		"-hls_time", "2", // Длительность сегмента 2 секунды
		"-hls_list_size", "10", // Хранить 10 сегментов в плейлисте
		"-hls_flags", "delete_segments+append_list", // Удалять старые сегменты и поддерживать список
		"-hls_delete_threshold", "1", // Удалять сегменты сразу
		"-hls_start_number_source", "datetime", // Нумерация по времени
		"-preset", "ultrafast", // Быстрый пресет
		"-tune", "zerolatency", // Минимальная задержка
		"-fflags", "nobuffer", // Без буферизации
		"-flags", "low_delay", // Низкая задержка
		"-strict", "-2", // Совместимость
		s.OutputPath, // Выходной файл
	}

	// Создаем команду
	s.cmd = exec.Command(ffmpegPath, args...)

	// Перенаправляем вывод в логи
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	// Запускаем процесс
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска ffmpeg: %v", err)
	}

	s.Active = true

	log.Printf("✅ HLS поток запущен для канала %s", s.ChannelID)
	log.Printf("📄 Playlist: %s", s.PlaylistURL)

	// Мониторим процесс
	go s.monitor()

	// Ждем создания плейлиста
	time.Sleep(3 * time.Second)

	return nil
}

// stop останавливает FFmpeg процесс
func (s *Stream) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		// Посылаем сигнал завершения
		s.cmd.Process.Signal(os.Interrupt)

		// Ждем завершения
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Процесс завершился
		case <-time.After(5 * time.Second):
			// Принудительное завершение
			s.cmd.Process.Kill()
		}
	}

	s.Active = false
	log.Printf("⏹️ HLS поток остановлен для канала %s", s.ChannelID)
}

// monitor следит за состоянием процесса
func (s *Stream) monitor() {
	if s.cmd == nil {
		return
	}

	// Ждем завершения процесса
	err := s.cmd.Wait()

	s.mu.Lock()
	s.Active = false
	s.mu.Unlock()

	if err != nil {
		log.Printf("⚠️ FFmpeg процесс для канала %s завершился с ошибкой: %v", s.ChannelID, err)
	} else {
		log.Printf("ℹ️ FFmpeg процесс для канала %s завершился", s.ChannelID)
	}
}

// cleanupOldSegments периодически удаляет старые сегменты
func (sm *StreamManager) cleanupOldSegments(channelID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Проверяем, активен ли еще поток
			sm.streamsMux.RLock()
			stream, exists := sm.streams[channelID]
			sm.streamsMux.RUnlock()

			if !exists || !stream.Active {
				return
			}

			// Удаляем старые .ts файлы
			pattern := filepath.Join(sm.outputDir, fmt.Sprintf("channel_%s*.ts", channelID))
			files, _ := filepath.Glob(pattern)

			cutoff := time.Now().Add(-1 * time.Minute)
			for _, file := range files {
				if info, err := os.Stat(file); err == nil {
					if info.ModTime().Before(cutoff) {
						os.Remove(file)
					}
				}
			}
		}
	}
}

// cleanupStreamFiles удаляет все файлы потока
func (sm *StreamManager) cleanupStreamFiles(channelID string) {
	// Удаляем плейлист
	playlistPath := filepath.Join(sm.outputDir, fmt.Sprintf("channel_%s.m3u8", channelID))
	os.Remove(playlistPath)

	// Удаляем сегменты
	pattern := filepath.Join(sm.outputDir, fmt.Sprintf("channel_%s*.ts", channelID))
	files, _ := filepath.Glob(pattern)
	for _, file := range files {
		os.Remove(file)
	}
}

// StopAll останавливает все потоки
func (sm *StreamManager) StopAll() {
	sm.streamsMux.Lock()
	defer sm.streamsMux.Unlock()

	for channelID, stream := range sm.streams {
		stream.stop()
		go sm.cleanupStreamFiles(channelID)
	}

	sm.streams = make(map[string]*Stream)
}
