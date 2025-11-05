package entities_test

import (
	"fmt"
	"testing"

	"compress/internal/domain/entities"
)

func TestNewCompressionConfig(t *testing.T) {
	tests := []struct {
		name          string
		level         int
		expectedLevel int
	}{
		{"Normal level", 50, 50},
		{"Too low level", 5, 10},
		{"Too high level", 95, 90},
		{"Minimum level", 10, 10},
		{"Maximum level", 90, 90},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := entities.NewCompressionConfig(tt.level)
			if config.Level != tt.expectedLevel {
				t.Errorf("Expected level %d, got %d", tt.expectedLevel, config.Level)
			}
		})
	}
}

func TestCompressionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *entities.CompressionConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &entities.CompressionConfig{
				Level:        50,
				ImageQuality: 75,
			},
			wantErr: false,
		},
		{
			name: "Invalid compression level - too low",
			config: &entities.CompressionConfig{
				Level:        5,
				ImageQuality: 75,
			},
			wantErr: true,
		},
		{
			name: "Invalid compression level - too high",
			config: &entities.CompressionConfig{
				Level:        95,
				ImageQuality: 75,
			},
			wantErr: true,
		},
		{
			name: "Invalid image quality - too low",
			config: &entities.CompressionConfig{
				Level:        50,
				ImageQuality: 5,
			},
			wantErr: true,
		},
		{
			name: "Invalid image quality - too high",
			config: &entities.CompressionConfig{
				Level:        50,
				ImageQuality: 105,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompressionConfigLevels(t *testing.T) {
	tests := []struct {
		level                int
		expectedImageQuality int
		expectedMetadata     bool
		expectedAnnotations  bool
		expectedAttachments  bool
	}{
		{15, 90, false, false, false}, // Слабое сжатие
		{30, 75, true, false, false},  // Умеренное сжатие
		{50, 60, true, true, false},   // Среднее сжатие
		{70, 40, true, true, true},    // Высокое сжатие
		{85, 25, true, true, true},    // Максимальное сжатие
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Level %d", tt.level), func(t *testing.T) {
			config := entities.NewCompressionConfig(tt.level)

			if config.ImageQuality != tt.expectedImageQuality {
				t.Errorf("Expected ImageQuality %d, got %d", tt.expectedImageQuality, config.ImageQuality)
			}

			if config.RemoveMetadata != tt.expectedMetadata {
				t.Errorf("Expected RemoveMetadata %v, got %v", tt.expectedMetadata, config.RemoveMetadata)
			}

			if config.RemoveAnnotations != tt.expectedAnnotations {
				t.Errorf("Expected RemoveAnnotations %v, got %v", tt.expectedAnnotations, config.RemoveAnnotations)
			}

			if config.RemoveAttachments != tt.expectedAttachments {
				t.Errorf("Expected RemoveAttachments %v, got %v", tt.expectedAttachments, config.RemoveAttachments)
			}
		})
	}
}
