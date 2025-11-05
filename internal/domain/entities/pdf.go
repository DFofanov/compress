package entities

import (
	"time"
)

// PDFDocument представляет PDF документ
type PDFDocument struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
	Pages        int
}

// CompressionResult представляет результат сжатия
type CompressionResult struct {
	CurrentFile      string
	OriginalSize     int64
	CompressedSize   int64
	CompressionRatio float64
	SavedSpace       int64
	Success          bool
	Error            error
}

// CalculateCompressionRatio вычисляет коэффициент сжатия
func (cr *CompressionResult) CalculateCompressionRatio() {
	if cr.OriginalSize > 0 {
		cr.CompressionRatio = ((float64(cr.OriginalSize) - float64(cr.CompressedSize)) / float64(cr.OriginalSize)) * 100
		cr.SavedSpace = cr.OriginalSize - cr.CompressedSize
	}
}

// IsEffective проверяет, было ли сжатие эффективным
func (cr *CompressionResult) IsEffective() bool {
	return cr.Success && cr.CompressionRatio > 0
}
