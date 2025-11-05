package usecases

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"compressor/internal/domain/entities"
	"compressor/internal/domain/repositories"
)

// ProcessPDFsUseCase сценарий автоматической обработки PDF файлов
type ProcessPDFsUseCase struct {
	compressor       repositories.PDFCompressor
	fileRepo         repositories.FileRepository
	configRepo       repositories.ConfigRepository
	logger           repositories.Logger
	progressReporter func(entities.ProcessingStatus)
}

// NewProcessPDFsUseCase создает новый сценарий обработки PDF
func NewProcessPDFsUseCase(
	compressor repositories.PDFCompressor,
	fileRepo repositories.FileRepository,
	configRepo repositories.ConfigRepository,
	logger repositories.Logger,
) *ProcessPDFsUseCase {
	return &ProcessPDFsUseCase{
		compressor: compressor,
		fileRepo:   fileRepo,
		configRepo: configRepo,
		logger:     logger,
	}
}

// SetProgressReporter устанавливает функцию для отчета о прогрессе
func (uc *ProcessPDFsUseCase) SetProgressReporter(reporter func(entities.ProcessingStatus)) {
	uc.progressReporter = reporter
}

// reportProgress отправляет обновление прогресса
func (uc *ProcessPDFsUseCase) reportProgress(status *entities.ProcessingStatus) {
	if uc.progressReporter != nil {
		uc.progressReporter(*status)
	}
}

// Execute выполняет автоматическую обработку PDF файлов согласно конфигурации
func (uc *ProcessPDFsUseCase) Execute(config *entities.Config) error {
	// Фаза 1: Инициализация
	status := entities.NewProcessingStatus(0)
	status.SetPhase(entities.PhaseInitializing, "Инициализация обработки...")
	uc.reportProgress(status)

	uc.logInfo("╔════════════════════════════════════════════════════════════")
	uc.logInfo("║ Начало обработки PDF файлов")
	uc.logInfo("╠════════════════════════════════════════════════════════════")
	uc.logInfo("║ Исходная директория: %s", config.Scanner.SourceDirectory)

	if config.Scanner.ReplaceOriginal {
		uc.logInfo("║ Режим: Замена оригинальных файлов")
	} else {
		uc.logInfo("║ Целевая директория: %s", config.Scanner.TargetDirectory)
	}

	uc.logInfo("║ Алгоритм: %s", config.Compression.Algorithm)
	uc.logInfo("║ Уровень сжатия: %d%%", config.Compression.Level)
	uc.logInfo("║ Параллельных воркеров: %d", config.Processing.ParallelWorkers)
	uc.logInfo("╚════════════════════════════════════════════════════════════")

	// Проверяем существование исходной директории
	if !uc.fileRepo.FileExists(config.Scanner.SourceDirectory) {
		err := fmt.Errorf("исходная директория не существует: %s", config.Scanner.SourceDirectory)
		status.Fail(err)
		uc.reportProgress(status)
		return err
	}

	// Создаем целевую директорию, если нужно
	if !config.Scanner.ReplaceOriginal {
		if err := uc.fileRepo.CreateDirectory(config.Scanner.TargetDirectory); err != nil {
			err = fmt.Errorf("ошибка создания целевой директории: %w", err)
			status.Fail(err)
			uc.reportProgress(status)
			return err
		}
	}

	// Фаза 2: Сканирование файлов
	status.SetPhase(entities.PhaseScanning, "Сканирование PDF файлов...")
	uc.reportProgress(status)
	uc.logInfo("🔍 Сканирование директории...")

	files, err := uc.fileRepo.ListPDFFiles(config.Scanner.SourceDirectory)
	if err != nil {
		err = fmt.Errorf("ошибка получения списка файлов: %w", err)
		status.Fail(err)
		uc.reportProgress(status)
		return err
	}

	if len(files) == 0 {
		uc.logWarning("⚠️  PDF файлы не найдены в директории: %s", config.Scanner.SourceDirectory)
		status.Complete()
		uc.reportProgress(status)
		return nil
	}

	status.TotalFiles = len(files)
	uc.logSuccess("✓ Найдено файлов для обработки: %d", len(files))

	// Создаем конфигурацию сжатия
	compressionConfig := entities.NewCompressionConfigWithLicense(config.Compression.Level, config.Compression.UniPDFLicenseKey)

	if err := uc.configRepo.ValidateConfig(compressionConfig); err != nil {
		err = fmt.Errorf("ошибка валидации конфигурации сжатия: %w", err)
		status.Fail(err)
		uc.reportProgress(status)
		return err
	}

	// Фаза 3: Сжатие файлов
	status.SetPhase(entities.PhaseCompressing, "Сжатие PDF файлов...")
	uc.reportProgress(status)
	uc.logInfo("")
	uc.logInfo("🔄 Начало сжатия файлов...")
	uc.logInfo("─────────────────────────────────────────────────────────────")

	// Создаем воркеры для параллельной обработки
	workers := config.Processing.ParallelWorkers
	if workers <= 0 {
		workers = 1
	}

	// Каналы для координации работы
	jobs := make(chan string, len(files))
	results := make(chan *entities.CompressionResult, len(files))

	var wg sync.WaitGroup

	// Запускаем воркеров
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go uc.worker(w, jobs, results, &wg, config, compressionConfig, status)
	}

	// Отправляем задачи воркерам
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Горутина для сбора результатов
	go func() {
		wg.Wait()
		close(results)
	}()

	// Обрабатываем результаты
	fileCounter := 0
	for result := range results {
		fileCounter++
		status.AddResult(result)

		// Обновляем текущий файл
		status.SetCurrentFile(result.CurrentFile, result.OriginalSize)

		// Отправляем обновление прогресса
		uc.reportProgress(status)

		// Логируем результат обработки файла
		fileName := filepath.Base(result.CurrentFile)
		if result.Success && result.Error == nil {
			uc.logSuccess("[%d/%d] ✓ %s", fileCounter, status.TotalFiles, fileName)
			uc.logInfo("    └─ Размер: %.2f MB → %.2f MB",
				float64(result.OriginalSize)/1024/1024,
				float64(result.CompressedSize)/1024/1024)
			uc.logInfo("    └─ Сжатие: %.1f%% | Сэкономлено: %.2f MB",
				result.CompressionRatio,
				float64(result.SavedSpace)/1024/1024)
		} else {
			uc.logError("[%d/%d] ✗ %s", fileCounter, status.TotalFiles, fileName)
			uc.logError("    └─ Ошибка: %v", result.Error)
		}
	}

	// Финальная фаза
	status.Complete()
	uc.reportProgress(status)

	// Логируем итоговую статистику
	uc.logInfo("")
	uc.logInfo("╔════════════════════════════════════════════════════════════")
	uc.logInfo("║ Обработка завершена")
	uc.logInfo("╠════════════════════════════════════════════════════════════")
	uc.logInfo("║ Время выполнения: %s", status.FormatElapsedTime())
	uc.logInfo("╠════════════════════════════════════════════════════════════")
	uc.logInfo("║ Статистика файлов:")
	uc.logInfo("║   • Всего: %d", status.TotalFiles)
	uc.logSuccess("║   • Успешно: %d", status.SuccessfulFiles)

	if status.FailedFiles > 0 {
		uc.logError("║   • Ошибок: %d", status.FailedFiles)
	}

	if status.SkippedFiles > 0 {
		uc.logWarning("║   • Пропущено: %d", status.SkippedFiles)
	}

	if status.TotalOriginalSize > 0 {
		uc.logInfo("╠════════════════════════════════════════════════════════════")
		uc.logInfo("║ Статистика сжатия:")
		uc.logInfo("║   • Исходный размер: %.2f MB", float64(status.TotalOriginalSize)/1024/1024)
		uc.logInfo("║   • Сжатый размер: %.2f MB", float64(status.TotalCompressedSize)/1024/1024)
		uc.logSuccess("║   • Среднее сжатие: %.1f%%", status.AverageCompression)
		uc.logSuccess("║   • Сэкономлено: %.2f MB", float64(status.TotalSavedSpace)/1024/1024)
	}

	uc.logInfo("╚════════════════════════════════════════════════════════════")

	return nil
}

// worker обрабатывает файлы в отдельной горутине
func (uc *ProcessPDFsUseCase) worker(
	id int,
	jobs <-chan string,
	results chan<- *entities.CompressionResult,
	wg *sync.WaitGroup,
	config *entities.Config,
	compressionConfig *entities.CompressionConfig,
	status *entities.ProcessingStatus,
) {
	defer wg.Done()

	for inputFile := range jobs {
		fileName := filepath.Base(inputFile)

		// Определяем путь выходного файла
		var outputFile string
		if config.Scanner.ReplaceOriginal {
			outputFile = inputFile + ".tmp"
		} else {
			// Получаем относительный путь от исходной директории
			relPath, err := filepath.Rel(config.Scanner.SourceDirectory, inputFile)
			if err != nil {
				// Если не удалось получить относительный путь, используем просто имя файла
				outputFile = filepath.Join(config.Scanner.TargetDirectory, fileName)
			} else {
				// Сохраняем структуру директорий
				outputFile = filepath.Join(config.Scanner.TargetDirectory, relPath)
				// Создаем директорию для выходного файла
				outputDir := filepath.Dir(outputFile)
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					results <- &entities.CompressionResult{
						CurrentFile: inputFile,
						Success:     false,
						Error:       fmt.Errorf("не удалось создать директорию %s: %w", outputDir, err),
					}
					continue
				}
			}
		}

		// Получаем информацию о файле
		fileInfo, err := uc.fileRepo.GetFileInfo(inputFile)
		if err != nil {
			results <- &entities.CompressionResult{
				CurrentFile: inputFile,
				Success:     false,
				Error:       fmt.Errorf("ошибка получения информации о файле: %w", err),
			}
			continue
		}

		// Выполняем сжатие с повторными попытками
		var result *entities.CompressionResult
		for attempt := 0; attempt < config.Processing.RetryAttempts; attempt++ {
			result, err = uc.compressor.Compress(inputFile, outputFile, compressionConfig)
			if err == nil {
				break
			}

			if attempt < config.Processing.RetryAttempts-1 {
				if uc.logger != nil {
					uc.logger.Warning("Попытка %d/%d для файла %s не удалась: %v",
						attempt+1, config.Processing.RetryAttempts, fileName, err)
				}
				time.Sleep(time.Second * 2) // Пауза перед повторной попыткой
			}
		}

		if err != nil {
			results <- &entities.CompressionResult{
				CurrentFile:  inputFile,
				OriginalSize: fileInfo.Size,
				Success:      false,
				Error:        err,
			}
			continue
		}

		// Устанавливаем исходный размер и пересчитываем статистику
		result.CurrentFile = inputFile
		result.OriginalSize = fileInfo.Size
		result.CalculateCompressionRatio()

		// Если заменяем оригинал, переименовываем временный файл
		if config.Scanner.ReplaceOriginal {
			if err := uc.replaceOriginalFile(inputFile, outputFile); err != nil {
				result.Success = false
				result.Error = fmt.Errorf("ошибка замены оригинального файла: %w", err)
				// Удаляем временный файл при ошибке
				_ = os.Remove(outputFile)
				if uc.logger != nil {
					uc.logger.Error("Не удалось заменить оригинальный файл %s: %v", inputFile, err)
				}
			} else {
				// Успешно заменили - обновляем путь к файлу в результате
				result.CurrentFile = inputFile
				if uc.logger != nil {
					uc.logger.Info("Файл %s успешно заменен сжатой версией", inputFile)
				}
			}
		}

		results <- result
	}
}

// replaceOriginalFile заменяет оригинальный файл сжатым
func (uc *ProcessPDFsUseCase) replaceOriginalFile(originalFile, tempFile string) error {
	// Проверяем существование временного файла
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		return fmt.Errorf("временный файл не существует: %s", tempFile)
	}

	if uc.logger != nil {
		uc.logger.Info("Замена оригинального файла: %s", originalFile)
	}

	backupFile := originalFile + ".backup"

	// Создаем резервную копию оригинала
	if err := os.Rename(originalFile, backupFile); err != nil {
		if uc.logger != nil {
			uc.logger.Error("Ошибка создания резервной копии %s: %v", originalFile, err)
		}
		return fmt.Errorf("ошибка создания резервной копии: %w", err)
	}

	// Переименовываем временный файл в оригинальный
	if err := os.Rename(tempFile, originalFile); err != nil {
		if uc.logger != nil {
			uc.logger.Error("Ошибка замены файла %s: %v", originalFile, err)
		}
		// Восстанавливаем оригинальный файл из резервной копии
		_ = os.Rename(backupFile, originalFile)
		return fmt.Errorf("ошибка замены файла: %w", err)
	}

	// Удаляем резервную копию
	if err := os.Remove(backupFile); err != nil {
		if uc.logger != nil {
			uc.logger.Warning("Не удалось удалить резервную копию %s: %v", backupFile, err)
		}
	}

	if uc.logger != nil {
		uc.logger.Info("Оригинальный файл успешно заменен: %s", originalFile)
	}

	return nil
}

// Методы для логирования
func (uc *ProcessPDFsUseCase) logInfo(format string, args ...interface{}) {
	if uc.logger != nil {
		uc.logger.Info(format, args...)
	}
}

func (uc *ProcessPDFsUseCase) logSuccess(format string, args ...interface{}) {
	if uc.logger != nil {
		uc.logger.Success(format, args...)
	}
}

func (uc *ProcessPDFsUseCase) logWarning(format string, args ...interface{}) {
	if uc.logger != nil {
		uc.logger.Warning(format, args...)
	}
}

func (uc *ProcessPDFsUseCase) logError(format string, args ...interface{}) {
	if uc.logger != nil {
		uc.logger.Error(format, args...)
	}
}
