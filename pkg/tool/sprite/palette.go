package sprite

import (
	"errors"
	"image/color"
	"unicode/utf8"

	"golang.org/x/text/width"
)

var (
	ErrInvalidColorKey = errors.New("invalid color key: must be a single character")
	errKeyOccupied     = errors.New("key already occupied in palette")
)

// ColorKey is a single-rune string that identifies a color entry in a Palette.
//
// Validity requires all three of:
//   - exactly one Unicode code point (rune count == 1),
//   - valid UTF-8 (no surrogate halves U+D800–U+DFFF or other non-scalar values),
//   - narrow display width (East Asian Width ≠ Wide or Fullwidth), so that a
//     matrix row printed as a plain string aligns correctly in monospace output.
type ColorKey string

func (ck ColorKey) valid() bool {
	s := string(ck)
	if len(s) == 0 {
		return false
	}
	if !utf8.ValidString(s) {
		return false
	}
	r, size := utf8.DecodeRuneInString(s)
	if size != len(s) {
		return false // more than one rune
	}
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return false
	}
	return true
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
// (U+0030), skipping surrogates (U+D800–U+DFFF) and wide/fullwidth code points.
// The usable space is the set of narrow Unicode scalar values — hundreds of
// thousands of slots — so exhaustion is not a practical concern for game sprites.
// If the space is somehow exhausted, nextKey panics.
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
// internal rune counter. It skips:
//   - surrogate halves (U+D800–U+DFFF) and other non-scalar values, via utf8.ValidRune,
//   - wide and fullwidth code points, via golang.org/x/text/width,
//   - runes already reserved in this palette.
//
// Panics if the narrow Unicode scalar space is exhausted, which is not a
// realistic concern for game sprite palettes.
func (p *Palette) nextKey() ColorKey {
	for p.nextRune <= utf8.MaxRune {
		r := p.nextRune
		p.nextRune++
		if !utf8.ValidRune(r) {
			continue
		}
		switch width.LookupRune(r).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			continue
		}
		ck := ColorKey(string(r))
		if _, reserved := p.reserved[ck]; !reserved {
			return ck
		}
	}
	panic("sprite: palette key space exhausted")
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

// setKeyed pins c to exactly ck in the palette, bypassing the auto-increment
// mechanism. It is used by file loading to honour the key names written in the
// sprite definition file. The key is marked as reserved so nextKey will never
// reassign the slot. Note: setKeyed does not update the registry (the color→key
// deduplication map used by Add), so calling Add with the same RGBA value later
// will produce a fresh auto-assigned key rather than returning ck.
func (p *Palette) setKeyed(ck ColorKey, c color.RGBA) error {
	if !ck.valid() {
		return ErrInvalidColorKey
	}
	if _, exists := p.colors[ck]; exists {
		return errKeyOccupied
	}
	if _, exists := p.reserved[ck]; exists {
		return errKeyOccupied
	}
	p.colors[ck] = c
	p.reserved[ck] = struct{}{}
	return nil
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
