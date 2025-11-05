package repositories

import "compressor/internal/domain/entities"

// ConfigRepository интерфейс для работы с конфигурацией приложения
type AppConfigRepository interface {
	Load(configPath string) (*entities.Config, error)
	Save(configPath string, config *entities.Config) error
}
