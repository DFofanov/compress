package entities

import "time"

// Config представляет конфигурацию приложения
type Config struct {
	Scanner     ScannerConfig        `yaml:"scanner"`
	Compression AppCompressionConfig `yaml:"compression"`
	Processing  ProcessingConfig     `yaml:"processing"`
	Output      OutputConfig         `yaml:"output"`
}

// ScannerConfig настройки сканирования директорий
type ScannerConfig struct {
	SourceDirectory string `yaml:"source_directory"`
	TargetDirectory string `yaml:"target_directory"`
	ReplaceOriginal bool   `yaml:"replace_original"`
}

// AppCompressionConfig настройки сжатия приложения
type AppCompressionConfig struct {
	Level            int    `yaml:"level"`
	Algorithm        string `yaml:"algorithm"`
	AutoStart        bool   `yaml:"auto_start"`
	UniPDFLicenseKey string `yaml:"unipdf_license_key"`
	// Настройки сжатия изображений
	EnableJPEG  bool `yaml:"enable_jpeg"`
	EnablePNG   bool `yaml:"enable_png"`
	JPEGQuality int  `yaml:"jpeg_quality"` // Качество JPEG в процентах (10-50)
	PNGQuality  int  `yaml:"png_quality"`  // Качество PNG в процентах (10-50)
}

// ProcessingConfig настройки обработки
type ProcessingConfig struct {
	ParallelWorkers int `yaml:"parallel_workers"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
	RetryAttempts   int `yaml:"retry_attempts"`
}

// OutputConfig настройки вывода
type OutputConfig struct {
	LogLevel     string `yaml:"log_level"`
	ProgressBar  bool   `yaml:"progress_bar"`
	LogToFile    bool   `yaml:"log_to_file"`
	LogFileName  string `yaml:"log_file_name"`
	LogMaxSizeMB int    `yaml:"log_max_size_mb"`
}

// ProcessingStatus статус обработки
type ProcessingStatus struct {
	// Текущая фаза обработки
	Phase ProcessingPhase

	// Информация о текущем файле
	CurrentFile     string
	CurrentFileSize int64

	// Общая статистика
	TotalFiles      int
	ProcessedFiles  int
	SuccessfulFiles int
	FailedFiles     int
	SkippedFiles    int

	// Прогресс
	Progress float64

	// Статистика сжатия
	TotalOriginalSize   int64
	TotalCompressedSize int64
	TotalSavedSpace     int64
	AverageCompression  float64

	// Текущий результат
	LastResult *CompressionResult

	// Время выполнения
	StartTime     time.Time
	ElapsedTime   time.Duration
	EstimatedTime time.Duration

	// Состояние
	IsComplete bool
	Error      error

	// Сообщение для UI
	Message string
}

// ProcessingPhase фаза обработки
type ProcessingPhase int

const (
	PhaseInitializing ProcessingPhase = iota
	PhaseScanning
	PhaseCompressing
	PhaseReplacing
	PhaseCompleted
	PhaseFailed
)

// UIScreen типы экранов UI
type UIScreen int

const (
	UIScreenMenu UIScreen = iota
	UIScreenConfig
	UIScreenProcessing
	// UIScreenResults
)

// Validate проверяет корректность конфигурации приложения
func (c *AppCompressionConfig) Validate() error {
	// Проверка уровня сжатия
	if c.Level < 10 || c.Level > 90 {
		return ErrInvalidCompressionLevel
	}

	// Проверка качества JPEG
	if c.EnableJPEG {
		if c.JPEGQuality < 10 || c.JPEGQuality > 50 || c.JPEGQuality%5 != 0 {
			return ErrInvalidJPEGQuality
		}
	}

	// Проверка качества PNG
	if c.EnablePNG {
		if c.PNGQuality < 10 || c.PNGQuality > 50 || c.PNGQuality%5 != 0 {
			return ErrInvalidPNGQuality
		}
	}

	return nil
}

// GetSupportedImageFormats возвращает список поддерживаемых форматов изображений
func (c *AppCompressionConfig) GetSupportedImageFormats() []string {
	var formats []string
	if c.EnableJPEG {
		formats = append(formats, "JPEG")
	}
	if c.EnablePNG {
		formats = append(formats, "PNG")
	}
	return formats
}

// NewProcessingStatus создает новый статус обработки
func NewProcessingStatus(totalFiles int) *ProcessingStatus {
	return &ProcessingStatus{
		Phase:      PhaseInitializing,
		TotalFiles: totalFiles,
		StartTime:  time.Now(),
	}
}

// UpdateProgress обновляет прогресс обработки
func (ps *ProcessingStatus) UpdateProgress() {
	if ps.TotalFiles > 0 {
		ps.Progress = float64(ps.ProcessedFiles) / float64(ps.TotalFiles) * 100
	}

	ps.ElapsedTime = time.Since(ps.StartTime)

	// Оценка оставшегося времени
	if ps.ProcessedFiles > 0 && ps.ProcessedFiles < ps.TotalFiles {
		avgTimePerFile := ps.ElapsedTime / time.Duration(ps.ProcessedFiles)
		remainingFiles := ps.TotalFiles - ps.ProcessedFiles
		ps.EstimatedTime = avgTimePerFile * time.Duration(remainingFiles)
	}
}

// AddResult добавляет результат обработки файла
func (ps *ProcessingStatus) AddResult(result *CompressionResult) {
	ps.ProcessedFiles++
	ps.LastResult = result

	if result.Success && result.Error == nil {
		ps.SuccessfulFiles++
		ps.TotalOriginalSize += result.OriginalSize
		ps.TotalCompressedSize += result.CompressedSize
		ps.TotalSavedSpace += result.SavedSpace

		// Пересчитываем среднее сжатие
		if ps.TotalOriginalSize > 0 {
			ps.AverageCompression = ((float64(ps.TotalOriginalSize) - float64(ps.TotalCompressedSize)) / float64(ps.TotalOriginalSize)) * 100
		}
	} else {
		ps.FailedFiles++
	}

	ps.UpdateProgress()
}

// SetPhase устанавливает фазу обработки
func (ps *ProcessingStatus) SetPhase(phase ProcessingPhase, message string) {
	ps.Phase = phase
	ps.Message = message
}

// SetCurrentFile устанавлиет текущий обрабатываемый файл
func (ps *ProcessingStatus) SetCurrentFile(filePath string, size int64) {
	ps.CurrentFile = filePath
	ps.CurrentFileSize = size
}

// Complete завершает обработку
func (ps *ProcessingStatus) Complete() {
	ps.IsComplete = true
	ps.Phase = PhaseCompleted
	ps.Progress = 100
	ps.ElapsedTime = time.Since(ps.StartTime)
	ps.EstimatedTime = 0
}

// Fail отмечает обработку как неудачную
func (ps *ProcessingStatus) Fail(err error) {
	ps.IsComplete = true
	ps.Phase = PhaseFailed
	ps.Error = err
	ps.ElapsedTime = time.Since(ps.StartTime)
}

// GetPhaseName возвращает название фазы
func (phase ProcessingPhase) String() string {
	switch phase {
	case PhaseInitializing:
		return "Инициализация"
	case PhaseScanning:
		return "Сканирование файлов"
	case PhaseCompressing:
		return "Сжатие файлов"
	case PhaseReplacing:
		return "Замена оригиналов"
	case PhaseCompleted:
		return "Завершено"
	case PhaseFailed:
		return "Ошибка"
	default:
		return "Неизвестно"
	}
}

// FormatElapsedTime форматирует время выполнения
func (ps *ProcessingStatus) FormatElapsedTime() string {
	duration := ps.ElapsedTime
	if duration < time.Second {
		return "< 1 сек"
	}
	if duration < time.Minute {
		return duration.Round(time.Second).String()
	}
	return duration.Round(time.Second).String()
}

// FormatEstimatedTime форматирует оставшееся время
func (ps *ProcessingStatus) FormatEstimatedTime() string {
	if ps.EstimatedTime == 0 {
		return "N/A"
	}
	duration := ps.EstimatedTime
	if duration < time.Second {
		return "< 1 сек"
	}
	if duration < time.Minute {
		return duration.Round(time.Second).String()
	}
	return duration.Round(time.Second).String()
}
