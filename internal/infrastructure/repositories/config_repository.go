package repositories

import (
	"compress/internal/domain/entities"
)

// ConfigRepository реализация репозитория конфигурации
type ConfigRepository struct{}

// NewConfigRepository создает новый репозиторий конфигурации
func NewConfigRepository() *ConfigRepository {
	return &ConfigRepository{}
}

// GetCompressionConfig получает конфигурацию сжатия по уровню
func (r *ConfigRepository) GetCompressionConfig(level int) (*entities.CompressionConfig, error) {
	config := entities.NewCompressionConfig(level)
	return config, nil
}

// ValidateConfig валидирует конфигурацию
func (r *ConfigRepository) ValidateConfig(config *entities.CompressionConfig) error {
	return config.Validate()
}
