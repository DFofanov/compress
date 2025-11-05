package usecases

import (
	"fmt"
	"os"
	"path/filepath"

	"compressor/internal/domain/entities"
	"compressor/internal/domain/repositories"
	"compressor/internal/infrastructure/compressors"
)

// CompressImageUseCase обрабатывает сжатие изображений
type CompressImageUseCase struct {
	logger     repositories.Logger
	compressor compressors.ImageCompressor
}

// NewCompressImageUseCase создает новый UseCase для сжатия изображений
func NewCompressImageUseCase(logger repositories.Logger, compressor compressors.ImageCompressor) *CompressImageUseCase {
	return &CompressImageUseCase{
		logger:     logger,
		compressor: compressor,
	}
}

// CompressImage сжимает одно изображение
func (uc *CompressImageUseCase) CompressImage(inputPath, outputPath string, config *entities.AppCompressionConfig) error {
	format := compressors.GetImageFormat(inputPath)
	if format == "" {
		return fmt.Errorf("неподдерживаемый формат изображения: %s", inputPath)
	}

	// Проверяем, включено ли сжатие для данного формата
	switch format {
	case "jpeg":
		if !config.EnableJPEG {
			uc.logger.Info(fmt.Sprintf("Пропуск JPEG файла (сжатие отключено): %s", inputPath))
			return nil
		}
		return uc.compressor.CompressJPEG(inputPath, outputPath, config.JPEGQuality)
	case "png":
		if !config.EnablePNG {
			uc.logger.Info(fmt.Sprintf("Пропуск PNG файла (сжатие отключено): %s", inputPath))
			return nil
		}
		return uc.compressor.CompressPNG(inputPath, outputPath, config.PNGQuality)
	default:
		return fmt.Errorf("неподдерживаемый формат изображения: %s", format)
	}
}

// ProcessImagesInDirectory обрабатывает все изображения в директории
func (uc *CompressImageUseCase) ProcessImagesInDirectory(sourceDir, targetDir string, config *entities.AppCompressionConfig, replaceOriginal bool) (*ProcessingResult, error) {
	result := &ProcessingResult{
		ProcessedFiles:  make([]string, 0),
		FailedFiles:     make([]ProcessingError, 0),
		SuccessfulFiles: 0,
		TotalFiles:      0,
	}

	// Если включены изображения, проверяем настройки
	if !config.EnableJPEG && !config.EnablePNG {
		uc.logger.Info("Сжатие изображений отключено в конфигурации")
		return result, nil
	}

	// Рекурсивно обходим директорию
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			uc.logger.Error(fmt.Sprintf("Ошибка доступа к файлу %s: %v", path, err))
			return nil // Продолжаем обработку других файлов
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Проверяем, является ли файл изображением
		if !compressors.IsImageFile(path) {
			return nil // Не изображение, пропускаем
		}

		result.TotalFiles++

		// Определяем путь выходного файла
		var outputPath string
		if replaceOriginal {
			outputPath = path
		} else {
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				uc.logger.Error(fmt.Sprintf("Не удалось получить относительный путь для %s: %v", path, err))
				result.FailedFiles = append(result.FailedFiles, ProcessingError{
					FilePath: path,
					Error:    err,
				})
				return nil
			}
			outputPath = filepath.Join(targetDir, relPath)

			// Создаем директорию для выходного файла
			outputDir := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				uc.logger.Error(fmt.Sprintf("Не удалось создать директорию %s: %v", outputDir, err))
				result.FailedFiles = append(result.FailedFiles, ProcessingError{
					FilePath: path,
					Error:    err,
				})
				return nil
			}
		}

		// Сжимаем изображение
		uc.logger.Info(fmt.Sprintf("Сжатие изображения: %s", path))
		err = uc.CompressImage(path, outputPath, config)
		if err != nil {
			uc.logger.Error(fmt.Sprintf("Ошибка сжатия изображения %s: %v", path, err))
			result.FailedFiles = append(result.FailedFiles, ProcessingError{
				FilePath: path,
				Error:    err,
			})
		} else {
			result.ProcessedFiles = append(result.ProcessedFiles, path)
			result.SuccessfulFiles++
			uc.logger.Info(fmt.Sprintf("Изображение успешно сжато: %s", path))
		}

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("ошибка обхода директории %s: %w", sourceDir, err)
	}

	return result, nil
}

// ProcessingResult результат обработки изображений
type ProcessingResult struct {
	ProcessedFiles  []string
	FailedFiles     []ProcessingError
	SuccessfulFiles int
	TotalFiles      int
}

// ProcessingError ошибка обработки файла
type ProcessingError struct {
	FilePath string
	Error    error
}

// GetSupportedImageExtensions возвращает список поддерживаемых расширений изображений
func GetSupportedImageExtensions() []string {
	return []string{".jpg", ".jpeg", ".png"}
}

// CountImageFiles подсчитывает количество изображений в директории
func CountImageFiles(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Игнорируем ошибки доступа к файлам
		}

		if !info.IsDir() && compressors.IsImageFile(path) {
			count++
		}

		return nil
	})

	return count, err
}
