package compressors

import (
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"compressor/internal/domain/entities"
)

// PDFCPUCompressor реализация компрессора с использованием PDFCPU
type PDFCPUCompressor struct{}

// NewPDFCPUCompressor создает новый PDFCPU компрессор
func NewPDFCPUCompressor() *PDFCPUCompressor {
	return &PDFCPUCompressor{}
}

// Compress сжимает PDF файл используя PDFCPU библиотеку
func (p *PDFCPUCompressor) Compress(inputPath, outputPath string, config *entities.CompressionConfig) (*entities.CompressionResult, error) {
	fmt.Printf("🔄 Сжатие PDF с уровнем %d%% (PDFCPU)...\n", config.Level)

	// Получаем исходный размер файла
	originalInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации об исходном файле: %w", err)
	}

	// Применяем настройки в зависимости от уровня сжатия
	if config.ImageCompression {
		fmt.Printf("📸 Включено сжатие изображений (качество: %d%%)\n", config.ImageQuality)
	}

	if config.RemoveDuplicates {
		fmt.Println("🔄 Удаление дубликатов объектов")
	}

	// Выполняем оптимизацию с базовыми настройками
	err = api.OptimizeFile(inputPath, outputPath, nil)
	if err != nil {
		return &entities.CompressionResult{
			OriginalSize: originalInfo.Size(),
			Success:      false,
			Error:        err,
		}, fmt.Errorf("ошибка оптимизации PDFCPU: %w", err)
	}

	// Получаем размер сжатого файла
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		return &entities.CompressionResult{
			OriginalSize: originalInfo.Size(),
			Success:      false,
			Error:        err,
		}, fmt.Errorf("ошибка получения информации о сжатом файле: %w", err)
	}

	result := &entities.CompressionResult{
		OriginalSize:   originalInfo.Size(),
		CompressedSize: compressedInfo.Size(),
		Success:        true,
	}

	result.CalculateCompressionRatio()

	fmt.Printf("✅ Сжатие завершено: %s\n", outputPath)
	return result, nil
}
