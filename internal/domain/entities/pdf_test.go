package entities_test

import (
	"testing"

	"compress/internal/domain/entities"
)

func TestCompressionResult_CalculateCompressionRatio(t *testing.T) {
	tests := []struct {
		name               string
		originalSize       int64
		compressedSize     int64
		expectedRatio      float64
		expectedSavedSpace int64
	}{
		{
			name:               "50% compression",
			originalSize:       1000,
			compressedSize:     500,
			expectedRatio:      50.0,
			expectedSavedSpace: 500,
		},
		{
			name:               "25% compression",
			originalSize:       1000,
			compressedSize:     750,
			expectedRatio:      25.0,
			expectedSavedSpace: 250,
		},
		{
			name:               "No compression",
			originalSize:       1000,
			compressedSize:     1000,
			expectedRatio:      0.0,
			expectedSavedSpace: 0,
		},
		{
			name:               "File got bigger",
			originalSize:       1000,
			compressedSize:     1100,
			expectedRatio:      -10.0,
			expectedSavedSpace: -100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &entities.CompressionResult{
				OriginalSize:   tt.originalSize,
				CompressedSize: tt.compressedSize,
			}

			result.CalculateCompressionRatio()

			if result.CompressionRatio != tt.expectedRatio {
				t.Errorf("Expected compression ratio %f, got %f", tt.expectedRatio, result.CompressionRatio)
			}

			if result.SavedSpace != tt.expectedSavedSpace {
				t.Errorf("Expected saved space %d, got %d", tt.expectedSavedSpace, result.SavedSpace)
			}
		})
	}
}

func TestCompressionResult_IsEffective(t *testing.T) {
	tests := []struct {
		name              string
		result            *entities.CompressionResult
		expectedEffective bool
	}{
		{
			name: "Effective compression",
			result: &entities.CompressionResult{
				OriginalSize:     1000,
				CompressedSize:   500,
				CompressionRatio: 50.0,
				Success:          true,
			},
			expectedEffective: true,
		},
		{
			name: "No compression but successful",
			result: &entities.CompressionResult{
				OriginalSize:     1000,
				CompressedSize:   1000,
				CompressionRatio: 0.0,
				Success:          true,
			},
			expectedEffective: false,
		},
		{
			name: "Good compression but failed",
			result: &entities.CompressionResult{
				OriginalSize:     1000,
				CompressedSize:   500,
				CompressionRatio: 50.0,
				Success:          false,
			},
			expectedEffective: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsEffective(); got != tt.expectedEffective {
				t.Errorf("IsEffective() = %v, want %v", got, tt.expectedEffective)
			}
		})
	}
}
