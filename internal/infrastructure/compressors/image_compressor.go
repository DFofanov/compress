package compressors

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

// ImageCompressor интерфейс для сжатия изображений
type ImageCompressor interface {
	CompressJPEG(inputPath, outputPath string, quality int) error
	CompressPNG(inputPath, outputPath string, quality int) error
}

// DefaultImageCompressor реализация компрессора изображений
type DefaultImageCompressor struct{}

// NewImageCompressor создает новый компрессор изображений
func NewImageCompressor() ImageCompressor {
	return &DefaultImageCompressor{}
}

// CompressJPEG сжимает JPEG файл с указанным качеством
func (c *DefaultImageCompressor) CompressJPEG(inputPath, outputPath string, quality int) error {
	// Открываем исходный файл
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %w", inputPath, err)
	}
	defer inputFile.Close()

	// Декодируем изображение
	img, err := jpeg.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("не удалось декодировать JPEG файл %s: %w", inputPath, err)
	}

	// Получаем размер исходного файла для сравнения
	inputFileInfo, err := inputFile.Stat()
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о файле %s: %w", inputPath, err)
	}
	originalSize := inputFileInfo.Size()

	// Вычисляем новый размер на основе качества
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Более агрессивное уменьшение размера для достижения реального сжатия
	// quality 10 -> 0.5 (50%), quality 50 -> 0.9 (90%)
	scaleFactor := 0.5 + float64(quality-10)/40.0*0.4
	if scaleFactor > 1.0 {
		scaleFactor = 1.0
	}

	newWidth := uint(float64(width) * scaleFactor)
	newHeight := uint(float64(height) * scaleFactor)

	// Изменяем размер изображения только если есть реальная польза
	var finalImg image.Image
	if newWidth < uint(width) && newHeight < uint(height) {
		finalImg = resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	} else {
		finalImg = img
	}

	// Маппинг качества: 10->30, 30->55, 50->75 (более консервативно)
	jpegQuality := 20 + int(float64(quality-10)/40.0*55.0)
	if jpegQuality < 20 {
		jpegQuality = 20
	}
	if jpegQuality > 75 {
		jpegQuality = 75
	}

	// Создаем временный файл для проверки результата
	tmpPath := outputPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("не удалось создать временный файл: %w", err)
	}

	// Кодируем во временный файл
	options := &jpeg.Options{Quality: jpegQuality}
	err = jpeg.Encode(tmpFile, finalImg, options)
	tmpFile.Close()

	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось закодировать JPEG: %w", err)
	}

	// Проверяем размер результата
	tmpInfo, err := os.Stat(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось получить информацию о временном файле: %w", err)
	}

	// Если сжатие неэффективно (файл больше или почти такой же), копируем оригинал
	if tmpInfo.Size() >= originalSize*95/100 {
		os.Remove(tmpPath)
		// Копируем оригинал
		inputFile.Seek(0, 0)
		outputFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("не удалось создать выходной файл: %w", err)
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, inputFile)
		if err != nil {
			return fmt.Errorf("не удалось скопировать файл: %w", err)
		}
		return nil
	}

	// Переименовываем временный файл в выходной
	err = os.Rename(tmpPath, outputPath)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось переименовать временный файл: %w", err)
	}

	return nil
}

// CompressPNG сжимает PNG файл с указанным качеством
func (c *DefaultImageCompressor) CompressPNG(inputPath, outputPath string, quality int) error {
	// Открываем исходный файл
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл %s: %w", inputPath, err)
	}
	defer inputFile.Close()

	// Получаем размер исходного файла для сравнения
	inputFileInfo, err := inputFile.Stat()
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о файле %s: %w", inputPath, err)
	}
	originalSize := inputFileInfo.Size()

	// Декодируем изображение
	img, err := png.Decode(inputFile)
	if err != nil {
		return fmt.Errorf("не удалось декодировать PNG файл %s: %w", inputPath, err)
	}

	// Вычисляем новый размер на основе качества
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Более консервативное масштабирование для PNG
	// quality 10 -> 0.6 (60%), quality 50 -> 0.9 (90%)
	scaleFactor := 0.6 + float64(quality-10)/40.0*0.3
	if scaleFactor > 1.0 {
		scaleFactor = 1.0
	}

	newWidth := uint(float64(width) * scaleFactor)
	newHeight := uint(float64(height) * scaleFactor)

	// Не изменяем размер для маленьких изображений
	if width < 400 && height < 400 {
		newWidth = uint(width)
		newHeight = uint(height)
	}

	// Изменяем размер изображения только если это даст выигрыш
	var finalImg image.Image
	if newWidth < uint(width) && newHeight < uint(height) {
		finalImg = resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	} else {
		finalImg = img
	}

	// Создаем временный файл для проверки результата
	tmpPath := outputPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("не удалось создать временный файл: %w", err)
	}

	// Для PNG используем максимальное сжатие
	encoder := &png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	err = encoder.Encode(tmpFile, finalImg)
	tmpFile.Close()

	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось закодировать PNG: %w", err)
	}

	// Проверяем размер результата
	tmpInfo, err := os.Stat(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось получить информацию о временном файле: %w", err)
	}

	// Если сжатие неэффективно (файл больше или почти такой же), копируем оригинал
	if tmpInfo.Size() >= originalSize*95/100 {
		os.Remove(tmpPath)
		// Копируем оригинал
		inputFile.Seek(0, 0)
		outputFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("не удалось создать выходной файл: %w", err)
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, inputFile)
		if err != nil {
			return fmt.Errorf("не удалось скопировать файл: %w", err)
		}
		return nil
	}

	// Переименовываем временный файл в выходной
	err = os.Rename(tmpPath, outputPath)
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось переименовать временный файл: %w", err)
	}

	return nil
}

// IsImageFile проверяет, является ли файл изображением поддерживаемого формата
func IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

// GetImageFormat возвращает формат изображения по расширению файла
func GetImageFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	default:
		return ""
	}
}
