package repositories

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"compress/internal/domain/entities"
)

// FileSystemRepository реализация репозитория для работы с файловой системой
type FileSystemRepository struct{}

// NewFileSystemRepository создает новый репозиторий файловой системы
func NewFileSystemRepository() *FileSystemRepository {
	return &FileSystemRepository{}
}

// GetFileInfo получает информацию о PDF файле
func (r *FileSystemRepository) GetFileInfo(path string) (*entities.PDFDocument, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &entities.PDFDocument{
		Path:         path,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		Pages:        0, // TODO: Можно добавить определение количества страниц
	}, nil
}

// FileExists проверяет существование файла
func (r *FileSystemRepository) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirectory создает директорию
func (r *FileSystemRepository) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// ListPDFFiles возвращает список PDF файлов в директории и всех подпапках
func (r *FileSystemRepository) ListPDFFiles(directory string) ([]string, error) {
	var pdfFiles []string

	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".pdf") {
			pdfFiles = append(pdfFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(pdfFiles)
	return pdfFiles, nil
}
