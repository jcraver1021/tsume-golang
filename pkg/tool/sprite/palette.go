package sprite

import (
	"errors"
	"image/color"
)

var (
	ErrInvalidColorKey = errors.New("invalid color key: must be a single character")
)

// ColorKey is a single-byte string that identifies a color entry in a Palette.
//
// Validity is defined by byte length (len(ck) == 1), not rune count, so only
// ASCII code points U+0000–U+007F produce valid keys. Runes above U+007F
// encode to two or more bytes in UTF-8 and are therefore invalid.
type ColorKey string

func (ck ColorKey) valid() bool {
	return len(ck) == 1
}

func fromString(s string) (ColorKey, error) {
	ck := ColorKey(s)
	if !ck.valid() {
		return "", ErrInvalidColorKey
	}

	return ck, nil
}

// Palette represents a collection of colors identified by single-character keys.
//
// Key space: auto-generated keys are assigned sequentially starting from '0'
// (U+0030). Because valid keys must be single-byte ASCII (U+0000–U+007F),
// there are at most 80 slots available before key generation silently overflows
// into multi-byte territory (see nextKey). Each call to Add for a new color and
// each call to Reserve draws from this pool. For typical game sprites — which
// have small palettes — this limit is unlikely to be reached, but it should be
// treated as a hard cap when designing data pipelines that construct palettes
// programmatically.
type Palette struct {
	colors   map[ColorKey]color.RGBA
	registry map[color.RGBA]ColorKey
	reserved map[ColorKey]struct{}
	nextRune rune
}

func NewPalette() *Palette {
	return &Palette{
		colors:   make(map[ColorKey]color.RGBA),
		registry: make(map[color.RGBA]ColorKey),
		reserved: make(map[ColorKey]struct{}),
		nextRune: '0',
	}
}

// nextKey returns the next available auto-generated key and advances the
// internal rune counter. Once nextRune exceeds U+007F the generated key will
// be multi-byte UTF-8 and will fail valid(), causing silent misbehavior in any
// code that validates keys (Reserve, fromString). Callers should ensure the
// total number of Add and Reserve calls stays within the ~80-key ASCII budget.
func (p *Palette) nextKey() ColorKey {
	ck := ColorKey(string(p.nextRune))
	p.nextRune++
	for {
		if _, reserved := p.reserved[ck]; !reserved {
			break
		}
		ck = ColorKey(string(p.nextRune))
		p.nextRune++
	}
	return ck
}

func (p *Palette) Reserve(ck ColorKey) (ColorKey, error) {
	if !ck.valid() {
		return "", ErrInvalidColorKey
	}

	needNew := false
	if _, exists := p.colors[ck]; exists {
		needNew = true
	}
	if _, exists := p.reserved[ck]; exists {
		needNew = true
	}

	if needNew {
		ck = p.nextKey()
	}

	p.reserved[ck] = struct{}{}

	return ck, nil
}

func (p *Palette) Add(c color.RGBA) (ColorKey, bool) {
	if ck, exists := p.registry[c]; exists {
		return ck, false
	}

	ck := p.nextKey()
	p.colors[ck] = c
	p.registry[c] = ck
	return ck, true
}

func (p *Palette) Get(ck ColorKey) (color.RGBA, bool) {
	c, exists := p.colors[ck]
	return c, exists
}

// clone returns a deep copy of the palette with the same key assignments
// and auto-increment state. Callers that own AnimationSequence values must
// update those sequences to reference the new palette.
func (p *Palette) clone() *Palette {
	newP := &Palette{
		colors:   make(map[ColorKey]color.RGBA, len(p.colors)),
		registry: make(map[color.RGBA]ColorKey, len(p.registry)),
		reserved: make(map[ColorKey]struct{}, len(p.reserved)),
		nextRune: p.nextRune,
	}
	for k, v := range p.colors {
		newP.colors[k] = v
	}
	for k, v := range p.registry {
		newP.registry[k] = v
	}
	for k := range p.reserved {
		newP.reserved[k] = struct{}{}
	}
	return newP
}
