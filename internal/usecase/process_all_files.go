package usecases

import (
	"fmt"
	"path/filepath"
	"strings"

	"compress/internal/domain/entities"
	"compress/internal/domain/repositories"
	"compress/internal/infrastructure/compressors"
)

// ProcessAllFilesUseCase сценарий для обработки всех поддерживаемых типов файлов
type ProcessAllFilesUseCase struct {
	pdfProcessor   *ProcessPDFsUseCase
	imageProcessor *CompressImageUseCase
	logger         repositories.Logger
}

// NewProcessAllFilesUseCase создает новый сценарий обработки всех файлов
func NewProcessAllFilesUseCase(
	pdfProcessor *ProcessPDFsUseCase,
	imageProcessor *CompressImageUseCase,
	logger repositories.Logger,
) *ProcessAllFilesUseCase {
	return &ProcessAllFilesUseCase{
		pdfProcessor:   pdfProcessor,
		imageProcessor: imageProcessor,
		logger:         logger,
	}
}

// Execute выполняет обработку всех поддерживаемых файлов
func (uc *ProcessAllFilesUseCase) Execute(config *entities.Config) error {
	uc.logger.Info("Начинаем обработку файлов")
	uc.logger.Info("Исходная директория: %s", config.Scanner.SourceDirectory)

	var processedPDFs, processedImages bool

	// Обрабатываем PDF файлы
	if uc.shouldProcessPDFs(config) {
		uc.logger.Info("Обработка PDF файлов...")
		err := uc.pdfProcessor.Execute(config)
		if err != nil {
			uc.logger.Error("Ошибка обработки PDF файлов: %v", err)
			return fmt.Errorf("ошибка обработки PDF файлов: %w", err)
		}
		processedPDFs = true
		uc.logger.Info("Обработка PDF файлов завершена")
	}

	// Обрабатываем изображения
	if uc.shouldProcessImages(config) {
		uc.logger.Info("Обработка изображений...")
		result, err := uc.imageProcessor.ProcessImagesInDirectory(
			config.Scanner.SourceDirectory,
			config.Scanner.TargetDirectory,
			&config.Compression,
			config.Scanner.ReplaceOriginal,
		)
		if err != nil {
			uc.logger.Error("Ошибка обработки изображений: %v", err)
			return fmt.Errorf("ошибка обработки изображений: %w", err)
		}

		// Логируем результаты обработки изображений
		uc.logger.Info("Обработка изображений завершена. Всего файлов: %d, Успешно: %d, Ошибок: %d",
			result.TotalFiles, result.SuccessfulFiles, len(result.FailedFiles))

		for _, failed := range result.FailedFiles {
			uc.logger.Error("Не удалось обработать изображение %s: %v", failed.FilePath, failed.Error)
		}

		processedImages = true
	}

	if !processedPDFs && !processedImages {
		uc.logger.Warning("Не выбрано ни одного типа файлов для обработки")
		return fmt.Errorf("не выбрано ни одного типа файлов для обработки")
	}

	uc.logger.Info("Обработка всех файлов завершена успешно")
	return nil
}

// shouldProcessPDFs проверяет, нужно ли обрабатывать PDF файлы
func (uc *ProcessAllFilesUseCase) shouldProcessPDFs(config *entities.Config) bool {
	// PDF файлы обрабатываются всегда, если есть алгоритм сжатия
	return config.Compression.Algorithm != ""
}

// shouldProcessImages проверяет, нужно ли обрабатывать изображения
func (uc *ProcessAllFilesUseCase) shouldProcessImages(config *entities.Config) bool {
	return config.Compression.EnableJPEG || config.Compression.EnablePNG
}

// GetSupportedFileTypes возвращает список поддерживаемых типов файлов
func (uc *ProcessAllFilesUseCase) GetSupportedFileTypes(config *entities.Config) []string {
	var types []string

	if uc.shouldProcessPDFs(config) {
		types = append(types, "PDF")
	}

	if config.Compression.EnableJPEG {
		types = append(types, "JPEG")
	}

	if config.Compression.EnablePNG {
		types = append(types, "PNG")
	}

	return types
}

// IsFileSupported проверяет, поддерживается ли данный файл для обработки
func (uc *ProcessAllFilesUseCase) IsFileSupported(filename string, config *entities.Config) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	// Проверяем PDF
	if ext == ".pdf" && uc.shouldProcessPDFs(config) {
		return true
	}

	// Проверяем изображения
	if compressors.IsImageFile(filename) {
		format := compressors.GetImageFormat(filename)
		switch format {
		case "jpeg":
			return config.Compression.EnableJPEG
		case "png":
			return config.Compression.EnablePNG
		}
	}

	return false
}
