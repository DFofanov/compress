package usecases

import (
	"fmt"
	"path/filepath"

	"compress/internal/domain/entities"
	"compress/internal/domain/repositories"
)

// CompressPDFUseCase сценарий сжатия одного PDF файла
type CompressPDFUseCase struct {
	compressor repositories.PDFCompressor
	fileRepo   repositories.FileRepository
	configRepo repositories.ConfigRepository
}

// NewCompressPDFUseCase создает новый сценарий сжатия PDF
func NewCompressPDFUseCase(
	compressor repositories.PDFCompressor,
	fileRepo repositories.FileRepository,
	configRepo repositories.ConfigRepository,
) *CompressPDFUseCase {
	return &CompressPDFUseCase{
		compressor: compressor,
		fileRepo:   fileRepo,
		configRepo: configRepo,
	}
}

// Execute выполняет сжатие PDF файла
func (uc *CompressPDFUseCase) Execute(inputPath string, outputPath string, compressionLevel int) (*entities.CompressionResult, error) {
	// Проверяем существование входного файла
	if !uc.fileRepo.FileExists(inputPath) {
		return nil, entities.ErrFileNotFound
	}

	// Получаем информацию о файле
	fileInfo, err := uc.fileRepo.GetFileInfo(inputPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации о файле: %w", err)
	}

	// Создаем конфигурацию сжатия
	config, err := uc.configRepo.GetCompressionConfig(compressionLevel)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания конфигурации: %w", err)
	}

	// Валидируем конфигурацию
	if err := uc.configRepo.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}

	// Генерируем имя выходного файла, если не указано
	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		base := inputPath[:len(inputPath)-len(ext)]
		outputPath = base + "_compressed" + ext
	}

	// Выполняем сжатие
	result, err := uc.compressor.Compress(inputPath, outputPath, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка сжатия файла: %w", err)
	}

	// Устанавливаем исходный размер
	result.OriginalSize = fileInfo.Size
	result.CalculateCompressionRatio()

	return result, nil
}
