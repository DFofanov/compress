package config

import (
	"compressor/internal/domain/entities"
	"os"

	"gopkg.in/yaml.v3"
)

// Repository реализация репозитория конфигурации
type Repository struct{}

// NewRepository создает новый репозиторий конфигурации
func NewRepository() *Repository {
	return &Repository{}
}

// Load загружает конфигурацию из файла
func (r *Repository) Load(configPath string) (*entities.Config, error) {
	// Если файл не существует, создаем конфигурацию по умолчанию
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return r.createDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config entities.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save сохраняет конфигурацию в файл
func (r *Repository) Save(configPath string, config *entities.Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// createDefaultConfig создает конфигурацию по умолчанию
func (r *Repository) createDefaultConfig() *entities.Config {
	return &entities.Config{
		Scanner: entities.ScannerConfig{
			SourceDirectory: "./pdfs",
			TargetDirectory: "./compressed",
			ReplaceOriginal: false,
		},
		Compression: entities.AppCompressionConfig{
			Level:     50,
			Algorithm: "pdfcpu",
			AutoStart: false,
		},
		Processing: entities.ProcessingConfig{
			ParallelWorkers: 2,
			TimeoutSeconds:  30,
			RetryAttempts:   3,
		},
		Output: entities.OutputConfig{
			LogLevel:     "info",
			ProgressBar:  true,
			LogToFile:    true,
			LogFileName:  "compressor.log",
			LogMaxSizeMB: 10,
		},
	}
}
