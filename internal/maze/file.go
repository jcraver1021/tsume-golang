package maze

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

func WriteImageToFile(m *image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, *m); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}
