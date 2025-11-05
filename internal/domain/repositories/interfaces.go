package repositories

import (
	"compress/internal/domain/entities"
)

// PDFCompressor интерфейс для сжатия PDF файлов
type PDFCompressor interface {
	Compress(inputPath, outputPath string, config *entities.CompressionConfig) (*entities.CompressionResult, error)
}

// FileRepository интерфейс для работы с файловой системой
type FileRepository interface {
	GetFileInfo(path string) (*entities.PDFDocument, error)
	FileExists(path string) bool
	CreateDirectory(path string) error
	ListPDFFiles(directory string) ([]string, error)
}

// ConfigRepository интерфейс для работы с конфигурацией
type ConfigRepository interface {
	GetCompressionConfig(level int) (*entities.CompressionConfig, error)
	ValidateConfig(config *entities.CompressionConfig) error
}
