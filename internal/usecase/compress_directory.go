package usecases

import (
	"fmt"
	"path/filepath"

	"compressor/internal/domain/entities"
	"compressor/internal/domain/repositories"
)

// CompressDirectoryUseCase сценарий сжатия всех PDF файлов в директории
type CompressDirectoryUseCase struct {
	compressor repositories.PDFCompressor
	fileRepo   repositories.FileRepository
	configRepo repositories.ConfigRepository
}

// NewCompressDirectoryUseCase создает новый сценарий сжатия директории
func NewCompressDirectoryUseCase(
	compressor repositories.PDFCompressor,
	fileRepo repositories.FileRepository,
	configRepo repositories.ConfigRepository,
) *CompressDirectoryUseCase {
	return &CompressDirectoryUseCase{
		compressor: compressor,
		fileRepo:   fileRepo,
		configRepo: configRepo,
	}
}

// DirectoryCompressionResult результат сжатия директории
type DirectoryCompressionResult struct {
	TotalFiles   int
	SuccessCount int
	FailedCount  int
	Results      []*entities.CompressionResult
	Errors       []error
}

// Execute выполняет сжатие всех PDF файлов в директории
func (uc *CompressDirectoryUseCase) Execute(inputDir, outputDir string, compressionLevel int) (*DirectoryCompressionResult, error) {
	// Проверяем существование входной директории
	if !uc.fileRepo.FileExists(inputDir) {
		return nil, entities.ErrDirectoryNotFound
	}

	// Создаем выходную директорию
	if err := uc.fileRepo.CreateDirectory(outputDir); err != nil {
		return nil, fmt.Errorf("ошибка создания выходной директории: %w", err)
	}

	// Получаем список PDF файлов
	files, err := uc.fileRepo.ListPDFFiles(inputDir)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка файлов: %w", err)
	}

	if len(files) == 0 {
		return nil, entities.ErrNoFilesFound
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

	result := &DirectoryCompressionResult{
		TotalFiles: len(files),
		Results:    make([]*entities.CompressionResult, 0, len(files)),
		Errors:     make([]error, 0),
	}

	// Обрабатываем каждый файл
	for _, inputFile := range files {
		fileName := filepath.Base(inputFile)
		outputFile := filepath.Join(outputDir, fmt.Sprintf("compressed_%s", fileName))

		// Получаем информацию о файле
		fileInfo, err := uc.fileRepo.GetFileInfo(inputFile)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("ошибка получения информации о файле %s: %w", fileName, err))
			result.FailedCount++
			continue
		}

		// Выполняем сжатие
		compressionResult, err := uc.compressor.Compress(inputFile, outputFile, config)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("ошибка сжатия файла %s: %w", fileName, err))
			result.FailedCount++
			continue
		}

		// Устанавливаем исходный размер и вычисляем коэффициент сжатия
		compressionResult.OriginalSize = fileInfo.Size
		compressionResult.CalculateCompressionRatio()

		result.Results = append(result.Results, compressionResult)
		result.SuccessCount++
	}

	return result, nil
}
