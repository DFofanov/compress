package controllers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"compressor/internal/domain/entities"
	usecases "compressor/internal/usecase"
)

// CLIController контроллер для командной строки
//
// ⚠️ DEPRECATED / LEGACY CODE ⚠️
//
// Данный контроллер НЕ используется в текущей версии приложения.
// Приложение использует только TUI интерфейс (internal/presentation/tui/manager.go).
// Сохранен для возможной будущей поддержки CLI режима или миграции на cobra/viper.
//
// Рекомендация: при необходимости CLI использовать флаги в main.go вместо этого контроллера.
type CLIController struct {
	compressPDFUseCase       *usecases.CompressPDFUseCase
	compressDirectoryUseCase *usecases.CompressDirectoryUseCase
}

// NewCLIController создает новый CLI контроллер
func NewCLIController(
	compressPDFUseCase *usecases.CompressPDFUseCase,
	compressDirectoryUseCase *usecases.CompressDirectoryUseCase,
) *CLIController {
	return &CLIController{
		compressPDFUseCase:       compressPDFUseCase,
		compressDirectoryUseCase: compressDirectoryUseCase,
	}
}

// HandleSingleFile обрабатывает сжатие одного файла
func (c *CLIController) HandleSingleFile(inputPath, outputPath string) error {
	fmt.Println("🔥 PDF Compressor - Сжатие PDF файлов")
	fmt.Println("====================================")

	// Запрашиваем уровень сжатия
	compressionLevel := c.askForCompressionLevel()

	fmt.Printf("\n🚀 Начинаем сжатие файла: %s\n", inputPath)

	// Выполняем сжатие
	result, err := c.compressPDFUseCase.Execute(inputPath, outputPath, compressionLevel)
	if err != nil {
		return fmt.Errorf("ошибка сжатия: %w", err)
	}

	// Показываем результаты
	c.showCompressionResult(result, outputPath)

	return nil
}

// HandleDirectory обрабатывает сжатие директории
func (c *CLIController) HandleDirectory(inputDir, outputDir string) error {
	fmt.Println("🔥 PDF Compressor - Сжатие директории PDF файлов")
	fmt.Println("================================================")

	// Запрашиваем уровень сжатия
	compressionLevel := c.askForCompressionLevel()

	fmt.Printf("\n🚀 Начинаем сжатие директории: %s\n", inputDir)

	// Выполняем сжатие
	result, err := c.compressDirectoryUseCase.Execute(inputDir, outputDir, compressionLevel)
	if err != nil {
		return fmt.Errorf("ошибка сжатия директории: %w", err)
	}

	// Показываем результаты
	c.showDirectoryResult(result)

	return nil
}

// askForCompressionLevel запрашивает уровень сжатия у пользователя
func (c *CLIController) askForCompressionLevel() int {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🎯 Выберите уровень сжатия:")
	fmt.Println("10-20%: Слабое сжатие (высокое качество)")
	fmt.Println("21-40%: Умеренное сжатие (хорошее качество)")
	fmt.Println("41-60%: Среднее сжатие (среднее качество)")
	fmt.Println("61-80%: Высокое сжатие (низкое качество)")
	fmt.Println("81-90%: Максимальное сжатие (очень низкое качество)")

	for {
		fmt.Print("\nВведите процент сжатия (10-90): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("❌ Ошибка ввода")
			continue
		}

		input = strings.TrimSpace(input)
		level, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("❌ Введите число")
			continue
		}

		if level < 10 || level > 90 {
			fmt.Println("❌ Уровень сжатия должен быть от 10 до 90")
			continue
		}

		return level
	}
}

// showCompressionResult показывает результат сжатия файла
func (c *CLIController) showCompressionResult(result *entities.CompressionResult, outputPath string) {
	fmt.Println("\n📊 Результаты сжатия:")
	fmt.Printf("Исходный размер: %.2f MB\n", float64(result.OriginalSize)/1024/1024)
	fmt.Printf("Сжатый размер: %.2f MB\n", float64(result.CompressedSize)/1024/1024)
	fmt.Printf("Сжатие: %.1f%%\n", result.CompressionRatio)
	fmt.Printf("Сэкономлено: %.2f MB\n", float64(result.SavedSpace)/1024/1024)

	if result.IsEffective() {
		fmt.Println("✅ Сжатие выполнено успешно!")
	} else {
		fmt.Println("⚠️ Файл не был сжат (возможно, уже оптимизирован)")
	}

	fmt.Printf("\n🎉 Готово! Сжатый файл сохранен как: %s\n", outputPath)
}

// showDirectoryResult показывает результат сжатия директории
func (c *CLIController) showDirectoryResult(result *usecases.DirectoryCompressionResult) {
	fmt.Printf("\n📊 Результаты сжатия директории:\n")
	fmt.Printf("Всего файлов: %d\n", result.TotalFiles)
	fmt.Printf("Успешно сжато: %d\n", result.SuccessCount)
	fmt.Printf("Ошибок: %d\n", result.FailedCount)

	// Показываем статистику по каждому файлу
	for i, fileResult := range result.Results {
		fmt.Printf("\n[%d] Сжатие: %.1f%%, Сэкономлено: %.2f MB\n",
			i+1, fileResult.CompressionRatio, float64(fileResult.SavedSpace)/1024/1024)
	}

	// Показываем ошибки, если есть
	if len(result.Errors) > 0 {
		fmt.Println("\n❌ Ошибки:")
		for i, err := range result.Errors {
			fmt.Printf("[%d] %v\n", i+1, err)
		}
	}

	fmt.Printf("\n🎉 Обработка завершена! Успешно сжато: %d/%d файлов\n",
		result.SuccessCount, result.TotalFiles)
}
