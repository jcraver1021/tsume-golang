package util

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/fonts/PressStart2P-Regular.ttf
var pressStart2PBytes []byte

// FontManager handles font loading and caching
type FontManager struct {
	source *text.GoTextFaceSource
	faces  map[float64]*text.GoTextFace
	mu     sync.RWMutex
}

var (
	defaultManager *FontManager
	managerOnce    sync.Once
)

// GetDefaultFontManager returns the singleton font manager instance
func GetDefaultFontManager() (*FontManager, error) {
	var err error
	managerOnce.Do(func() {
		defaultManager, err = NewFontManager()
	})
	return defaultManager, err
}

// NewFontManager creates a new font manager with the embedded Press Start 2P font
func NewFontManager() (*FontManager, error) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(pressStart2PBytes))
	if err != nil {
		return nil, err
	}

	return &FontManager{
		source: source,
		faces:  make(map[float64]*text.GoTextFace),
	}, nil
}

// GetFace returns a font face for the given size, caching for reuse
func (fm *FontManager) GetFace(size float64) *text.GoTextFace {
	fm.mu.RLock()
	if face, ok := fm.faces[size]; ok {
		fm.mu.RUnlock()
		return face
	}
	fm.mu.RUnlock()

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Double-check in case another goroutine created it while we waited for the lock
	if face, ok := fm.faces[size]; ok {
		return face
	}

	face := &text.GoTextFace{
		Source: fm.source,
		Size:   size,
	}
	fm.faces[size] = face
	return face
}
