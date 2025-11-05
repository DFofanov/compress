package entities

// CompressionConfig представляет конфигурацию сжатия
type CompressionConfig struct {
	Level             int    // Уровень сжатия (10-90)
	ImageQuality      int    // Качество изображений (10-100)
	ImageCompression  bool   // Сжимать изображения
	RemoveDuplicates  bool   // Удалять дубликаты объектов
	CompressStreams   bool   // Сжимать потоки данных
	RemoveMetadata    bool   // Удалять метаданные
	RemoveAnnotations bool   // Удалять аннотации
	RemoveAttachments bool   // Удалять вложения
	OptimizeForWeb    bool   // Оптимизировать для веб
	UniPDFLicenseKey  string // Лицензионный ключ для UniPDF
}

// NewCompressionConfig создает конфигурацию сжатия на основе уровня
func NewCompressionConfig(level int) *CompressionConfig {
	return NewCompressionConfigWithLicense(level, "")
}

// NewCompressionConfigWithLicense создает конфигурацию сжатия с лицензионным ключом
func NewCompressionConfigWithLicense(level int, licenseKey string) *CompressionConfig {
	if level < 10 {
		level = 10
	}
	if level > 90 {
		level = 90
	}

	config := &CompressionConfig{
		Level:            level,
		RemoveDuplicates: true,
		CompressStreams:  true,
		OptimizeForWeb:   true,
		UniPDFLicenseKey: licenseKey,
	}

	switch {
	case level <= 20: // Слабое сжатие
		config.ImageQuality = 90
		config.ImageCompression = true
		config.RemoveMetadata = false
		config.RemoveAnnotations = false
		config.RemoveAttachments = false

	case level <= 40: // Умеренное сжатие
		config.ImageQuality = 75
		config.ImageCompression = true
		config.RemoveMetadata = true
		config.RemoveAnnotations = false
		config.RemoveAttachments = false

	case level <= 60: // Среднее сжатие
		config.ImageQuality = 60
		config.ImageCompression = true
		config.RemoveMetadata = true
		config.RemoveAnnotations = true
		config.RemoveAttachments = false

	case level <= 80: // Высокое сжатие
		config.ImageQuality = 40
		config.ImageCompression = true
		config.RemoveMetadata = true
		config.RemoveAnnotations = true
		config.RemoveAttachments = true

	default: // Максимальное сжатие (81-90%)
		config.ImageQuality = 25
		config.ImageCompression = true
		config.RemoveMetadata = true
		config.RemoveAnnotations = true
		config.RemoveAttachments = true
	}

	return config
}

// Validate проверяет корректность конфигурации
func (c *CompressionConfig) Validate() error {
	if c.Level < 10 || c.Level > 90 {
		return ErrInvalidCompressionLevel
	}
	if c.ImageQuality < 10 || c.ImageQuality > 100 {
		return ErrInvalidImageQuality
	}
	return nil
}
