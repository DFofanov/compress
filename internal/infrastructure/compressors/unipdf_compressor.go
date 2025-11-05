package compressors

import (
	"fmt"
	"os"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/model/optimize"

	"compress/internal/domain/entities"
)

// UniPDFCompressor реализация компрессора с использованием UniPDF
type UniPDFCompressor struct{}

// NewUniPDFCompressor создает новый UniPDF компрессор
func NewUniPDFCompressor() *UniPDFCompressor {
	return &UniPDFCompressor{}
}

// Compress сжимает PDF файл используя UniPDF библиотеку
func (u *UniPDFCompressor) Compress(inputPath, outputPath string, config *entities.CompressionConfig) (*entities.CompressionResult, error) {
	fmt.Printf("🔄 Сжатие PDF с уровнем %d%% (UniPDF)...\n", config.Level)

	// Инициализируем логгер
	common.SetLogger(common.NewConsoleLogger(common.LogLevelInfo))

	// Проверяем лицензионный ключ из конфигурации или переменной окружения
	licenseKey := config.UniPDFLicenseKey
	if licenseKey == "" {
		licenseKey = os.Getenv("UNIDOC_LICENSE_API_KEY")
	}

	if licenseKey == "" {
		return &entities.CompressionResult{
			OriginalSize: 0,
			Success:      false,
			Error:        fmt.Errorf("UniPDF требует лицензионный ключ. Установите его в конфигурации или в переменной UNIDOC_LICENSE_API_KEY. Используйте алгоритм 'pdfcpu' для бесплатной обработки или получите ключ на https://cloud.unidoc.io"),
		}, fmt.Errorf("UniPDF лицензия не настроена. Установите лицензионный ключ в конфигурации или используйте алгоритм 'pdfcpu'")
	}

	// Устанавливаем лицензионный ключ
	fmt.Printf("🔑 Устанавливаем лицензионный ключ UniPDF...\n")
	os.Setenv("UNIDOC_LICENSE_API_KEY", licenseKey) // Получаем исходный размер файла
	originalInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения информации об исходном файле: %w", err)
	}

	// Открываем исходный PDF файл
	pdfReader, file, err := model.NewPdfReaderFromFile(inputPath, nil)
	if err != nil {
		return &entities.CompressionResult{
			OriginalSize: originalInfo.Size(),
			Success:      false,
			Error:        err,
		}, fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	// Создаем writer с оптимизацией
	pdfWriter := model.NewPdfWriter()

	// Настраиваем оптимизацию в зависимости от уровня сжатия
	optimizer := optimize.New(optimize.Options{
		CombineDuplicateDirectObjects:   true,
		CombineIdenticalIndirectObjects: true,
		ImageUpperPPI:                   float64(150 - config.Level), // чем выше уровень, тем ниже PPI
		ImageQuality:                    100 - config.Level,          // чем выше уровень, тем ниже качество
	})

	pdfWriter.SetOptimizer(optimizer)

	// Копируем страницы
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return &entities.CompressionResult{
			OriginalSize: originalInfo.Size(),
			Success:      false,
			Error:        err,
		}, fmt.Errorf("ошибка получения количества страниц: %w", err)
	}

	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return &entities.CompressionResult{
				OriginalSize: originalInfo.Size(),
				Success:      false,
				Error:        err,
			}, fmt.Errorf("ошибка получения страницы %d: %w", i, err)
		}

		err = pdfWriter.AddPage(page)
		if err != nil {
			return &entities.CompressionResult{
				OriginalSize: originalInfo.Size(),
				Success:      false,
				Error:        err,
			}, fmt.Errorf("ошибка добавления страницы %d: %w", i, err)
		}
	}

	// Сохраняем оптимизированный файл
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return &entities.CompressionResult{
			OriginalSize: originalInfo.Size(),
			Success:      false,
			Error:        err,
		}, fmt.Errorf("ошибка создания выходного файла: %w", err)
	}
	defer outputFile.Close()

	err = pdfWriter.Write(outputFile)
	if err != nil {
		return &entities.CompressionResult{
			OriginalSize: originalInfo.Size(),
			Success:      false,
			Error:        err,
		}, fmt.Errorf("ошибка записи файла: %w", err)
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
