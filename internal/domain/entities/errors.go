package entities

import "errors"

// Доменные ошибки
var (
	ErrInvalidCompressionLevel = errors.New("уровень сжатия должен быть от 10 до 90")
	ErrInvalidImageQuality     = errors.New("качество изображения должно быть от 10 до 100")
	ErrInvalidJPEGQuality      = errors.New("качество JPEG должно быть от 10 до 50 с шагом 5")
	ErrInvalidPNGQuality       = errors.New("качество PNG должно быть от 10 до 50 с шагом 5")
	ErrFileNotFound            = errors.New("файл не найден")
	ErrInvalidFileFormat       = errors.New("неверный формат файла")
	ErrCompressionFailed       = errors.New("ошибка сжатия файла")
	ErrDirectoryNotFound       = errors.New("директория не найдена")
	ErrNoFilesFound            = errors.New("PDF файлы не найдены")
)
