package model

import (
	"fmt"
	"os"
)

type FileModelRepository struct{}

func NewFileModelRepository() *FileModelRepository {
	return &FileModelRepository{}
}

func (r *FileModelRepository) CreateModel(modelPath string) error {
	if err := os.MkdirAll(modelPath, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}
	return nil
}

// Exists reports whether modelPath is a model directory. The fork drops
// index.yaml, so we check that the path is a directory rather than that an
// index file exists.
func (r *FileModelRepository) Exists(modelPath string) bool {
	info, err := os.Stat(modelPath)
	return err == nil && info.IsDir()
}
