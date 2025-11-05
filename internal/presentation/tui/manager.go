package tui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"compressor/internal/domain/entities"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v3"
)

// ConfigData структура для отображения конфигурации в UI
type ConfigData struct {
	Scanner struct {
		SourceDirectory string `yaml:"source_directory"`
		TargetDirectory string `yaml:"target_directory"`
		ReplaceOriginal bool   `yaml:"replace_original"`
	} `yaml:"scanner"`
	Compression struct {
		Level            int    `yaml:"level"`
		Algorithm        string `yaml:"algorithm"`
		AutoStart        bool   `yaml:"auto_start"`
		UniPDFLicenseKey string `yaml:"unipdf_license_key"`
		EnableJPEG       bool   `yaml:"enable_jpeg"`
		EnablePNG        bool   `yaml:"enable_png"`
		JPEGQuality      int    `yaml:"jpeg_quality"`
		PNGQuality       int    `yaml:"png_quality"`
	} `yaml:"compression"`
	Processing struct {
		ParallelWorkers int `yaml:"parallel_workers"`
		TimeoutSeconds  int `yaml:"timeout_seconds"`
		RetryAttempts   int `yaml:"retry_attempts"`
	} `yaml:"processing"`
	Output struct {
		LogLevel     string `yaml:"log_level"`
		ProgressBar  bool   `yaml:"progress_bar"`
		LogToFile    bool   `yaml:"log_to_file"`
		LogFileName  string `yaml:"log_file_name"`
		LogMaxSizeMB int    `yaml:"log_max_size_mb"`
	} `yaml:"output"`
}

// UI Configuration constants
const (
	MaxLogBufferSize     = 1000
	LogFlushInterval     = 50 * time.Millisecond
	ProgressBarWidth     = 40
	MaxFileNameLength    = 60
	MaxFileNameDisplay   = 57
	ProgressViewHeight   = 9
	FormItemLicenseIndex = 5
)

// Manager управляет TUI интерфейсом
type Manager struct {
	app           *tview.Application
	pages         *tview.Pages
	currentScreen entities.UIScreen

	// UI компоненты
	mainMenu     *tview.List
	configForm   *tview.Form
	progressView *tview.TextView
	logView      *tview.TextView
	statusBar    *tview.TextView

	// Callbacks
	onStartProcessing func()

	// Состояние
	configData   ConfigData
	logBuffer    []string
	statusMutex  sync.RWMutex
	isProcessing bool

	// Оптимизированный батчинг логов через канал
	logChan  chan string
	logDone  chan struct{}
	logMutex sync.Mutex
}

// NewManager создает новый менеджер TUI
func NewManager() *Manager {
	m := &Manager{
		app:       tview.NewApplication(),
		pages:     tview.NewPages(),
		logBuffer: make([]string, 0, MaxLogBufferSize),
		logChan:   make(chan string, 100), // Buffered channel для батчинга
		logDone:   make(chan struct{}),
	}
	// Запускаем горутину обработки логов
	go m.logProcessor()
	return m
}

// Initialize инициализирует TUI
func (m *Manager) Initialize() {
	m.loadConfig()
	m.createUI()
	m.setupKeyBindings()
}

// Run запускает TUI
func (m *Manager) Run() error {
	return m.app.SetRoot(m.pages, true).EnableMouse(true).Run()
}

// SetOnStartProcessing устанавливает callback для начала обработки
func (m *Manager) SetOnStartProcessing(callback func()) {
	m.onStartProcessing = callback
}

// SendStatusUpdate отправляет обновление статуса
func (m *Manager) SendStatusUpdate(status entities.ProcessingStatus) {
	m.updateProgress(status)
}

// loadConfig загружает конфигурацию
func (m *Manager) loadConfig() {
	configPath := "config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Создаем конфигурацию по умолчанию
		m.configData = ConfigData{
			Scanner: struct {
				SourceDirectory string `yaml:"source_directory"`
				TargetDirectory string `yaml:"target_directory"`
				ReplaceOriginal bool   `yaml:"replace_original"`
			}{
				SourceDirectory: "./pdfs",
				TargetDirectory: "./compressed",
				ReplaceOriginal: false,
			},
			Compression: struct {
				Level            int    `yaml:"level"`
				Algorithm        string `yaml:"algorithm"`
				AutoStart        bool   `yaml:"auto_start"`
				UniPDFLicenseKey string `yaml:"unipdf_license_key"`
				EnableJPEG       bool   `yaml:"enable_jpeg"`
				EnablePNG        bool   `yaml:"enable_png"`
				JPEGQuality      int    `yaml:"jpeg_quality"`
				PNGQuality       int    `yaml:"png_quality"`
			}{
				Level:            50,
				Algorithm:        "pdfcpu",
				AutoStart:        false,
				UniPDFLicenseKey: "",
				EnableJPEG:       false,
				EnablePNG:        false,
				JPEGQuality:      30,
				PNGQuality:       25,
			},
			Processing: struct {
				ParallelWorkers int `yaml:"parallel_workers"`
				TimeoutSeconds  int `yaml:"timeout_seconds"`
				RetryAttempts   int `yaml:"retry_attempts"`
			}{
				ParallelWorkers: 2,
				TimeoutSeconds:  30,
				RetryAttempts:   3,
			},
			Output: struct {
				LogLevel     string `yaml:"log_level"`
				ProgressBar  bool   `yaml:"progress_bar"`
				LogToFile    bool   `yaml:"log_to_file"`
				LogFileName  string `yaml:"log_file_name"`
				LogMaxSizeMB int    `yaml:"log_max_size_mb"`
			}{
				LogLevel:     "info",
				ProgressBar:  true,
				LogToFile:    true,
				LogFileName:  "compressor.log",
				LogMaxSizeMB: 10,
			},
		}
		m.saveConfig()
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	yaml.Unmarshal(data, &m.configData)
}

// saveConfig сохраняет конфигурацию
func (m *Manager) saveConfig() {
	data, err := yaml.Marshal(&m.configData)
	if err != nil {
		return
	}
	os.WriteFile("config.yaml", data, 0644)
}

// createUI создает пользовательский интерфейс
func (m *Manager) createUI() {
	m.createMainMenu()
	m.createConfigScreen()
	m.createProcessingScreen()
	// m.createResultsScreen()

	m.pages.AddPage("menu", m.mainMenu, true, true)
	m.pages.AddPage("config", m.configForm, true, false)
	m.pages.AddPage("processing", m.createProcessingLayout(), true, false)

	m.currentScreen = entities.UIScreenMenu
}

// createMainMenu создает главное меню
func (m *Manager) createMainMenu() {
	m.mainMenu = tview.NewList().
		AddItem("🚀 Запуск алгоритма сжатия", "Начать автоматическое сжатие PDF файлов", '1', func() {
			m.startProcessing()
		}).
		AddItem("⚙️ Конфигурация", "Настроить параметры сжатия и обработки", '2', func() {
			m.switchToScreen(entities.UIScreenConfig)
		}).
		AddItem("❌ Выход", "Закрыть приложение", 'q', func() {
			m.Cleanup()
			m.app.Stop()
		})

	m.mainMenu.SetBorder(true).
		SetTitle("🔥 Universal File Compressor - Главное меню").
		SetTitleAlign(tview.AlignCenter)

	// Настраиваем стиль
	m.mainMenu.SetSelectedBackgroundColor(tcell.ColorDarkBlue).
		SetSelectedTextColor(tcell.ColorWhite).
		SetMainTextColor(tcell.ColorWhite).
		SetSecondaryTextColor(tcell.ColorGray)
}

// createConfigScreen создает экран конфигурации
func (m *Manager) createConfigScreen() {
	m.configForm = tview.NewForm().
		AddInputField("Исходная директория", m.configData.Scanner.SourceDirectory, 60, nil, func(text string) {
			m.configData.Scanner.SourceDirectory = text
		}).
		AddInputField("Целевая директория", m.configData.Scanner.TargetDirectory, 60, nil, func(text string) {
			m.configData.Scanner.TargetDirectory = text
		}).
		AddCheckbox("Заменить оригинал", m.configData.Scanner.ReplaceOriginal, func(checked bool) {
			m.configData.Scanner.ReplaceOriginal = checked
		}).
		AddInputField("Уровень сжатия (10-90)", strconv.Itoa(m.configData.Compression.Level), 10, nil, func(text string) {
			if level, err := strconv.Atoi(text); err == nil && level >= 10 && level <= 90 {
				m.configData.Compression.Level = level
			}
		}).
		AddDropDown("Алгоритм", []string{"pdfcpu", "unipdf"}, func() int {
			if m.configData.Compression.Algorithm == "unipdf" {
				return 1
			}
			return 0
		}(), func(option string, optionIndex int) {
			m.configData.Compression.Algorithm = option
			m.updateLicenseFieldVisibility()
		}).
		AddInputField("Лицензия UniPDF (UNIDOC_LICENSE_API_KEY)", m.configData.Compression.UniPDFLicenseKey, 60, nil, func(text string) {
			m.configData.Compression.UniPDFLicenseKey = text
		}).
		AddCheckbox("Автостарт", m.configData.Compression.AutoStart, func(checked bool) {
			m.configData.Compression.AutoStart = checked
		}).
		AddCheckbox("Сжимать JPEG", m.configData.Compression.EnableJPEG, func(checked bool) {
			m.configData.Compression.EnableJPEG = checked
		}).
		AddDropDown("Качество JPEG (%)", []string{"10", "15", "20", "25", "30", "35", "40", "45", "50"}, func() int {
			return (m.configData.Compression.JPEGQuality - 10) / 5
		}(), func(option string, optionIndex int) {
			if quality, err := strconv.Atoi(option); err == nil {
				m.configData.Compression.JPEGQuality = quality
			}
		}).
		AddCheckbox("Сжимать PNG", m.configData.Compression.EnablePNG, func(checked bool) {
			m.configData.Compression.EnablePNG = checked
		}).
		AddDropDown("Качество PNG (%)", []string{"10", "15", "20", "25", "30", "35", "40", "45", "50"}, func() int {
			return (m.configData.Compression.PNGQuality - 10) / 5
		}(), func(option string, optionIndex int) {
			if quality, err := strconv.Atoi(option); err == nil {
				m.configData.Compression.PNGQuality = quality
			}
		}).
		AddButton("Сохранить", func() {
			m.saveConfig()
			m.switchToScreen(entities.UIScreenMenu)
			// Позиционируемся на пункте "Конфигурация" (индекс 1)
			m.mainMenu.SetCurrentItem(1)
		})

	m.updateLicenseFieldVisibility()

	m.configForm.SetBorder(true).
		SetTitle("🔥 Universal File Compressor - Конфигурация (ESC - выйти без сохранения)").
		SetTitleAlign(tview.AlignCenter)

	// Обработка ESC для выхода без сохранения
	m.configForm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			// Перезагружаем конфигурацию из файла (отменяем изменения)
			m.loadConfig()
			m.switchToScreen(entities.UIScreenMenu)
			return nil
		}
		return event
	})
}

// createProcessingScreen создает экран обработки
func (m *Manager) createProcessingScreen() {
	m.progressView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetScrollable(true)

	m.progressView.SetBorder(true).
		SetTitle("📊 Прогресс обработки").
		SetTitleAlign(tview.AlignCenter)

	m.logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(MaxLogBufferSize)

	m.logView.SetBorder(true).
		SetTitle("📋 Журнал событий").
		SetTitleAlign(tview.AlignCenter)
}

// createProcessingLayout создает layout для экрана обработки
func (m *Manager) createProcessingLayout() *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(m.logView, 0, 1, false).
		AddItem(m.progressView, ProgressViewHeight, 0, false)
}

// setupKeyBindings настраивает горячие клавиши
func (m *Manager) setupKeyBindings() {
	m.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1:
			m.switchToScreen(entities.UIScreenMenu)
			return nil
		case tcell.KeyF2:
			m.switchToScreen(entities.UIScreenConfig)
			return nil
		case tcell.KeyF3:
			if m.isProcessing {
				m.switchToScreen(entities.UIScreenProcessing)
			}
			return nil
		case tcell.KeyEscape:
			// ESC работает по-разному в зависимости от экрана
			if m.currentScreen == entities.UIScreenConfig {
				// В конфигурации ESC обрабатывается локально формой
				return event
			} else if m.currentScreen != entities.UIScreenMenu {
				m.switchToScreen(entities.UIScreenMenu)
				return nil
			}
		}

		// Обработка числовых клавиш для меню
		if m.currentScreen == entities.UIScreenMenu {
			switch event.Rune() {
			case '1':
				m.startProcessing()
				return nil
			case '2':
				m.switchToScreen(entities.UIScreenConfig)
				return nil
			case 'q', 'Q':
				m.Cleanup()
				m.app.Stop()
				return nil
			}
		}

		return event
	})
}

// switchToScreen переключает на указанный экран
func (m *Manager) switchToScreen(screen entities.UIScreen) {
	m.statusMutex.Lock()
	defer m.statusMutex.Unlock()

	m.currentScreen = screen

	switch screen {
	case entities.UIScreenMenu:
		m.pages.SwitchToPage("menu")
	case entities.UIScreenConfig:
		// При входе в конфигурацию обновляем данные из файла и синхронизируем форму
		m.loadConfig()
		m.refreshConfigForm()
		m.pages.SwitchToPage("config")
	case entities.UIScreenProcessing:
		m.pages.SwitchToPage("processing")
	}
}

// startProcessing начинает обработку
func (m *Manager) startProcessing() {
	m.saveConfig()
	m.isProcessing = true
	m.switchToScreen(entities.UIScreenProcessing)

	if m.onStartProcessing != nil {
		go m.onStartProcessing()
	}
}

// updateProgress обновляет прогресс
func (m *Manager) updateProgress(status entities.ProcessingStatus) {
	if m.progressView == nil {
		return
	}

	// Обновляем прогресс-бар
	progressBar := m.createProgressBar(status.Progress, ProgressBarWidth)

	// Корректное усечение имени файла с учетом UTF-8
	displayFile := m.truncateFileName(status.CurrentFile, MaxFileNameLength, MaxFileNameDisplay)

	// Формируем текст статуса
	var progressText string

	// Фаза обработки
	phaseText := status.Phase.String()
	if status.Message != "" {
		phaseText = status.Message
	}

	progressText = fmt.Sprintf(
		"[yellow]⚙️  Фаза:[white] %s\n\n"+
			"[yellow]📁 Текущий файл:[white] %s\n",
		phaseText,
		filepath.Base(displayFile),
	)

	// Размер текущего файла
	if status.CurrentFileSize > 0 {
		progressText += fmt.Sprintf("[dim]   Размер: %.2f MB[white]\n", float64(status.CurrentFileSize)/1024/1024)
	}

	// Прогресс-бар
	progressText += fmt.Sprintf(
		"\n[cyan]📊 Прогресс:[white] %s [cyan]%.1f%%[white]\n\n",
		progressBar,
		status.Progress,
	)

	// Статистика файлов
	progressText += fmt.Sprintf(
		"[green]📈 Статистика файлов:[white]\n"+
			"  • Всего: [cyan]%d[white]\n"+
			"  • Обработано: [cyan]%d[white]\n"+
			"  • Успешно: [green]%d[white]",
		status.TotalFiles,
		status.ProcessedFiles,
		status.SuccessfulFiles,
	)

	if status.FailedFiles > 0 {
		progressText += fmt.Sprintf("\n  • Ошибок: [red]%d[white]", status.FailedFiles)
	}

	if status.SkippedFiles > 0 {
		progressText += fmt.Sprintf("\n  • Пропущено: [yellow]%d[white]", status.SkippedFiles)
	}

	// Статистика сжатия
	if status.TotalOriginalSize > 0 {
		progressText += fmt.Sprintf(
			"\n\n[green]💾 Статистика сжатия:[white]\n"+
				"  • Исходный размер: [cyan]%.2f MB[white]\n"+
				"  • Сжатый размер: [cyan]%.2f MB[white]\n"+
				"  • Среднее сжатие: [green]%.1f%%[white]\n"+
				"  • Сэкономлено: [green]%.2f MB[white]",
			float64(status.TotalOriginalSize)/1024/1024,
			float64(status.TotalCompressedSize)/1024/1024,
			status.AverageCompression,
			float64(status.TotalSavedSpace)/1024/1024,
		)
	}

	// Время выполнения
	progressText += fmt.Sprintf(
		"\n\n[yellow]⏱️  Время:[white]\n"+
			"  • Прошло: [cyan]%s[white]",
		status.FormatElapsedTime(),
	)

	if !status.IsComplete && status.EstimatedTime > 0 {
		progressText += fmt.Sprintf("\n  • Осталось: [cyan]~%s[white]", status.FormatEstimatedTime())
	}

	progressText += "\n\n"

	if status.IsComplete {
		if status.Error != nil {
			progressText += "[red]❌ Обработка завершена с ошибкой![white]\n"
			progressText += fmt.Sprintf("[red]Ошибка: %v[white]\n", status.Error)
		} else {
			progressText += "[green]✅ Обработка успешно завершена![white]\n"
		}
		progressText += "\n[yellow]F1[white] - Главное меню\n"
		progressText += "[yellow]ESC[white] - Главное меню\n"
		m.isProcessing = false
	} else {
		progressText += "[yellow]F1[white] - Главное меню\n"
		progressText += "[yellow]ESC[white] - Главное меню\n"
	}

	if status.Error != nil {
		progressText += fmt.Sprintf("\n[red]❌ Ошибка: %v[white]\n", status.Error)
	}

	// Обновляем UI потокобезопасно через QueueUpdateDraw
	m.app.QueueUpdateDraw(func() {
		m.progressView.SetText(progressText)
	})
}

// truncateFileName корректно усекает имя файла с учетом UTF-8
func (m *Manager) truncateFileName(fileName string, maxLength, truncateAt int) string {
	runes := []rune(fileName)
	if len(runes) <= maxLength {
		return fileName
	}
	return string(runes[:truncateAt]) + "..."
}

// createProgressBar создает красивый цветной прогресс-бар
func (m *Manager) createProgressBar(progress float64, width int) string {
	// Нормализуем значения
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}

	filled := int(math.Round(progress * float64(width) / 100))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	// Разные символы для заполненной и пустой части
	const filledChar = "█"
	const emptyChar = "░"

	// Цвет зависит от прогресса
	var color string
	switch {
	case progress < 25:
		color = "red"
	case progress < 50:
		color = "yellow"
	case progress < 75:
		color = "blue"
	default:
		color = "green"
	}

	filledPart := strings.Repeat(filledChar, filled)
	emptyPart := strings.Repeat(emptyChar, width-filled)

	return fmt.Sprintf("[%s]%s[gray]%s", color, filledPart, emptyPart)
}

// AddLog добавляет запись в лог через канал (неблокирующе)
func (m *Manager) AddLog(level, message string) {
	var color string
	switch strings.ToLower(level) {
	case "error":
		color = "red"
	case "warning":
		color = "yellow"
	case "success":
		color = "green"
	case "debug":
		color = "gray"
	default:
		color = "white"
	}

	logLine := fmt.Sprintf("[%s]%s:[white] %s", color, strings.ToUpper(level), message)

	// Неблокирующая отправка в канал
	select {
	case m.logChan <- logLine:
	default:
		// Если канал переполнен, пропускаем лог (лучше чем блокировка)
	}
}

// logProcessor обрабатывает логи в отдельной горутине с батчингом
func (m *Manager) logProcessor() {
	ticker := time.NewTicker(LogFlushInterval)
	defer ticker.Stop()

	batch := make([]string, 0, 50)

	for {
		select {
		case logLine := <-m.logChan:
			batch = append(batch, logLine)

			// Если накопился достаточный батч, сбрасываем
			if len(batch) >= 20 {
				m.flushLogBatch(batch)
				batch = make([]string, 0, 50)
			}

		case <-ticker.C:
			// Периодический сброс
			if len(batch) > 0 {
				m.flushLogBatch(batch)
				batch = make([]string, 0, 50)
			}

		case <-m.logDone:
			// Финальный сброс при завершении
			if len(batch) > 0 {
				m.flushLogBatch(batch)
			}
			return
		}
	}
}

// flushLogBatch сбрасывает батч логов в UI
func (m *Manager) flushLogBatch(batch []string) {
	m.statusMutex.Lock()
	m.logBuffer = append(m.logBuffer, batch...)

	// Ограничиваем размер буфера
	if len(m.logBuffer) > MaxLogBufferSize {
		m.logBuffer = m.logBuffer[len(m.logBuffer)-MaxLogBufferSize:]
	}

	// Создаем копию буфера для UI
	logText := strings.Join(m.logBuffer, "\n")
	m.statusMutex.Unlock()

	// Обновляем UI потокобезопасно
	if m.logView != nil {
		m.app.QueueUpdateDraw(func() {
			if m.logView != nil { // Двойная проверка
				m.logView.SetText(logText)
				m.logView.ScrollToEnd()
			}
		})
	}
}

// Cleanup освобождает ресурсы менеджера (идемпотентный)
func (m *Manager) Cleanup() {
	m.logMutex.Lock()
	defer m.logMutex.Unlock()

	// Проверяем, что канал еще открыт
	select {
	case <-m.logDone:
		// Канал уже закрыт
		return
	default:
		// Закрываем канал
		close(m.logDone)
	}
} // updateLicenseFieldVisibility обновляет видимость поля лицензии в зависимости от выбранного алгоритма
func (m *Manager) updateLicenseFieldVisibility() {
	if m.configForm == nil {
		return
	}

	// Получаем количество элементов формы
	formItemCount := m.configForm.GetFormItemCount()

	if formItemCount > FormItemLicenseIndex {
		// Получаем поле лицензии
		licenseField := m.configForm.GetFormItem(FormItemLicenseIndex)

		if m.configData.Compression.Algorithm == "unipdf" {
			// Показываем поле лицензии для UniPDF
			licenseField.(*tview.InputField).SetTitle("🔑 Лицензия UniPDF (UNIDOC_LICENSE_API_KEY) - ОБЯЗАТЕЛЬНО")
			licenseField.(*tview.InputField).SetFieldBackgroundColor(tcell.ColorDarkBlue)
		} else {
			// Скрываем поле лицензии для PDFCPU
			licenseField.(*tview.InputField).SetTitle("Лицензия UniPDF (не требуется для PDFCPU)")
			licenseField.(*tview.InputField).SetFieldBackgroundColor(tcell.ColorDarkGray)
		}
	}
}

// refreshConfigForm синхронизирует значения формы с текущими данными конфигурации
func (m *Manager) refreshConfigForm() {
	if m.configForm == nil {
		return
	}

	// 0: Исходная директория (Input)
	if item := m.configForm.GetFormItem(0); item != nil {
		item.(*tview.InputField).SetText(m.configData.Scanner.SourceDirectory)
	}
	// 1: Целевая директория (Input)
	if item := m.configForm.GetFormItem(1); item != nil {
		item.(*tview.InputField).SetText(m.configData.Scanner.TargetDirectory)
	}
	// 2: Заменить оригинал (Checkbox)
	if item := m.configForm.GetFormItem(2); item != nil {
		item.(*tview.Checkbox).SetChecked(m.configData.Scanner.ReplaceOriginal)
	}
	// 3: Уровень сжатия (Input)
	if item := m.configForm.GetFormItem(3); item != nil {
		item.(*tview.InputField).SetText(strconv.Itoa(m.configData.Compression.Level))
	}
	// 4: Алгоритм (DropDown)
	if item := m.configForm.GetFormItem(4); item != nil {
		dd := item.(*tview.DropDown)
		if m.configData.Compression.Algorithm == "unipdf" {
			dd.SetCurrentOption(1)
		} else {
			dd.SetCurrentOption(0)
		}
	}
	// 5: Лицензия UniPDF (Input)
	if item := m.configForm.GetFormItem(5); item != nil {
		item.(*tview.InputField).SetText(m.configData.Compression.UniPDFLicenseKey)
	}
	// 6: Автостарт (Checkbox)
	if item := m.configForm.GetFormItem(6); item != nil {
		item.(*tview.Checkbox).SetChecked(m.configData.Compression.AutoStart)
	}

	m.updateLicenseFieldVisibility()
}

// GetConfig возвращает текущую конфигурацию в формате entities.Config
func (m *Manager) GetConfig() *entities.Config {
	return &entities.Config{
		Scanner: entities.ScannerConfig{
			SourceDirectory: m.configData.Scanner.SourceDirectory,
			TargetDirectory: m.configData.Scanner.TargetDirectory,
			ReplaceOriginal: m.configData.Scanner.ReplaceOriginal,
		},
		Compression: entities.AppCompressionConfig{
			Level:            m.configData.Compression.Level,
			Algorithm:        m.configData.Compression.Algorithm,
			AutoStart:        m.configData.Compression.AutoStart,
			UniPDFLicenseKey: m.configData.Compression.UniPDFLicenseKey,
			EnableJPEG:       m.configData.Compression.EnableJPEG,
			EnablePNG:        m.configData.Compression.EnablePNG,
			JPEGQuality:      m.configData.Compression.JPEGQuality,
			PNGQuality:       m.configData.Compression.PNGQuality,
		},
		Processing: entities.ProcessingConfig{
			ParallelWorkers: m.configData.Processing.ParallelWorkers,
			TimeoutSeconds:  m.configData.Processing.TimeoutSeconds,
			RetryAttempts:   m.configData.Processing.RetryAttempts,
		},
		Output: entities.OutputConfig{
			LogLevel:     m.configData.Output.LogLevel,
			ProgressBar:  m.configData.Output.ProgressBar,
			LogToFile:    m.configData.Output.LogToFile,
			LogFileName:  m.configData.Output.LogFileName,
			LogMaxSizeMB: m.configData.Output.LogMaxSizeMB,
		},
	}
}
