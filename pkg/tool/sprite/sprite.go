package sprite

import (
	"errors"
	"image/color"
)

var (
	ErrInvalidSpriteMatrix = errors.New("invalid sprite matrix: must have at least one row and one column")
)

// Sprite is an immutable-after-construction 2D grid of ColorKeys with an
// associated Palette and optional per-key AnimationSequences. There is no
// pixel-level write API by design: a sprite's content is fixed at construction
// time (via NewSprite or BlankSprite) and can only be changed by producing a
// new sprite through Compose, ComposeExpanding, or Clone. This keeps sprites
// safe to share across systems without defensive copying.
type Sprite struct {
	matrix             [][]ColorKey
	palette            *Palette
	animationSequences map[ColorKey]*AnimationSequence
}

func BlankSprite(rows, cols int) (*Sprite, error) {
	palette := NewPalette()
	key, _ := palette.Add(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // bool (new vs existing) discarded — only the key is needed
	matrix := make([][]ColorKey, rows)
	for i := range matrix {
		matrix[i] = make([]ColorKey, cols)
		for j := range matrix[i] {
			matrix[i][j] = key
		}
	}

	return NewSprite(matrix, palette, map[ColorKey]*AnimationSequence{})
}

func NewSprite(matrix [][]ColorKey, palette *Palette, animationSequences map[ColorKey]*AnimationSequence) (*Sprite, error) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return nil, ErrInvalidSpriteMatrix
	}

	for _, row := range matrix {
		if len(row) != len(matrix[0]) {
			return nil, ErrInvalidSpriteMatrix
		}
	}

	return &Sprite{
		matrix:             matrix,
		palette:            palette,
		animationSequences: animationSequences,
	}, nil
}

func (s *Sprite) Width() int {
	return len(s.matrix[0])
}

func (s *Sprite) Height() int {
	return len(s.matrix)
}

// Render returns the current pixel grid without mutating any state. It is safe
// to call multiple times on the same sprite to obtain the same frame. Call
// Advance to step the animation clock forward.
func (s *Sprite) Render() [][]color.RGBA {
	rendered := make([][]color.RGBA, s.Height())
	for i := range rendered {
		rendered[i] = make([]color.RGBA, s.Width())
		for j := range rendered[i] {
			if ck := s.matrix[i][j]; ck != "" {
				if seq, ok := s.animationSequences[ck]; ok {
					rendered[i][j] = seq.getColor()
				} else if c, ok := s.palette.Get(ck); ok {
					rendered[i][j] = c
				} else {
					rendered[i][j] = color.RGBA{R: 0, G: 0, B: 0, A: 0}
				}
			}
		}
	}
	return rendered
}

// Advance ticks all animation sequences forward by one step. Call this once
// per game-loop iteration, after Render.
func (s *Sprite) Advance() {
	for _, seq := range s.animationSequences {
		seq.Advance()
	}
}

// Compose overlays other onto s using Porter-Duff "src over dst" blending and
// returns the result as a new Sprite. s and other are not modified.
//
// Offset semantics: dst[i][j] corresponds to other[i+rowOffset][j+colOffset].
// A positive rowOffset shifts other upward relative to s (other's top edge
// moves above s's top edge); a negative rowOffset shifts it downward. The same
// logic applies to colOffset and the horizontal axis.
//
// Out-of-bounds offsets are intentional no-ops: if the offset places other
// entirely outside s's bounds the returned sprite is a pixel-identical copy of
// s. Pixels of other that partially overlap s's edges are clipped; pixels of
// other that fall outside s are ignored.
func (s *Sprite) Compose(other *Sprite, rowOffset, colOffset int) (*Sprite, error) {
	newPalette := NewPalette()
	blankCk, _ := newPalette.Add(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // bool (new vs existing) discarded — first entry in a fresh palette

	withinOther := func(row, col int) bool {
		return row+rowOffset >= 0 && row+rowOffset < other.Height() && col+colOffset >= 0 && col+colOffset < other.Width()
	}

	newMatrix := make([][]ColorKey, s.Height())
	for i := range newMatrix {
		newMatrix[i] = make([]ColorKey, s.Width())
		for j := range newMatrix[i] {
			newMatrix[i][j] = blankCk
		}
	}

	// Unaffected animation coordinates for later
	type Coord struct {
		row int
		col int
	}
	animationCoords := []Coord{}
	newAnimationSequences := map[ColorKey]*AnimationSequence{}

	// First, we will add all unaffected colors from this sprite (destination)
	for i := range s.matrix {
		for j := range s.matrix[i] {
			if !withinOther(i, j) {
				if color, ok := s.palette.Get(s.matrix[i][j]); ok {
					newCk, _ := newPalette.Add(color) // This may or may not be the same key, but it will be the same color
					newMatrix[i][j] = newCk
				} else {
					animationCoords = append(animationCoords, Coord{row: i, col: j})
				}
			} else {
				// Only fully transparent source colors are unaffected
				// (we want destination indices to be lower; we will do fully source and composite later)
				// (we also will scan for blended animations later)
				if srcColor, ok := other.palette.Get(other.matrix[i+rowOffset][j+colOffset]); ok && srcColor.A == 0 {
					// ok is false when dst is an animation at this pixel; the switcher loop
					// below overwrites such cells, so the zero color here is a safe placeholder.
					dstColor, _ := s.palette.Get(s.matrix[i][j])
					newCk, _ := newPalette.Add(dstColor) // bool (new vs existing) discarded
					newMatrix[i][j] = newCk
				}
			}
		}
	}

	// Next, we will add all unaffected colors from the other sprite (source) that overlap with this sprite
	for i := range other.matrix {
		for j := range other.matrix[i] {
			if withinOther(i-rowOffset, j-colOffset) {
				if color, ok := other.palette.Get(other.matrix[i][j]); ok && color.A == 255 {
					newCk, _ := newPalette.Add(color) // bool (new vs existing) discarded
					newMatrix[i-rowOffset][j-colOffset] = newCk
				}
			}
		}
	}

	// Next, we will re-add unaffected animation coordinates from the original sprite into the new matrix
	for _, coord := range animationCoords {
		if seq, ok := s.animationSequences[s.matrix[coord.row][coord.col]]; ok {
			// Matrix keys originate from Add/Reserve on a well-formed palette and are always
			// single-byte ASCII (valid ColorKeys); ErrInvalidColorKey cannot occur here.
			newCk, _ := newPalette.Reserve(s.matrix[coord.row][coord.col])
			newMatrix[coord.row][coord.col] = newCk
			newAnimationSequences[newCk] = seq
		}
	}

	// Finally, we blend the overlap
	type overlapCase int
	const (
		bothColor overlapCase = iota
		srcAnimation
		dstAnimation
		bothAnimation
	)
	switcher := func(i, j int) overlapCase {
		srcIsColor := false
		dstIsColor := false
		if _, ok := other.palette.Get(other.matrix[i][j]); ok {
			srcIsColor = true
		}
		if _, ok := s.palette.Get(s.matrix[i-rowOffset][j-colOffset]); ok {
			dstIsColor = true
		}
		switch {
		case srcIsColor && dstIsColor:
			return bothColor
		case !srcIsColor && dstIsColor:
			return srcAnimation
		case srcIsColor && !dstIsColor:
			return dstAnimation
		case !srcIsColor && !dstIsColor:
			return bothAnimation
		}
		return bothColor // default, should not reach here
	}
	for i := range other.matrix {
		for j := range other.matrix[i] {
			if withinOther(i-rowOffset, j-colOffset) {
				switch switcher(i, j) {
				case bothColor:
					// switcher confirmed both pixels are static colors in their respective palettes; ok is guaranteed.
					srcColor, _ := other.palette.Get(other.matrix[i][j])
					dstColor, _ := s.palette.Get(s.matrix[i-rowOffset][j-colOffset])
					blendedColor := alphaComposite(srcColor, dstColor)
					newCk, _ := newPalette.Add(blendedColor) // bool (new vs existing) discarded
					newMatrix[i-rowOffset][j-colOffset] = newCk
				case srcAnimation:
					// switcher confirmed srcIsColor=false; a well-formed sprite guarantees the
					// key is in animationSequences (absent key would be nil, causing a panic below).
					src, _ := other.animationSequences[other.matrix[i][j]]
					// switcher confirmed dstIsColor=true; ok is guaranteed.
					dstColor, _ := s.palette.Get(s.matrix[i-rowOffset][j-colOffset])
					dstCk, _ := newPalette.Add(dstColor) // bool (new vs existing) discarded
					// Translate src frames into newPalette so blend result lives entirely in newPalette.
					translatedSrcFrames := make([]ColorKey, len(src.frames))
					for fi, srcCk := range src.frames {
						// invariant: AnimationSequence frames are always valid palette keys; ok is guaranteed.
						srcFrameColor, _ := other.palette.Get(srcCk)
						translatedSrcFrames[fi], _ = newPalette.Add(srcFrameColor) // bool (new vs existing) discarded
					}
					// invariant: src.frameDuration > 0 (valid seq); frames non-empty; error impossible.
					translatedSrc, _ := NewAnimationSequence(newPalette, translatedSrcFrames, src.frameDuration)
					dst, _ := NewAnimationSequence(newPalette, []ColorKey{dstCk}, src.frameDuration)
					newAnimation := translatedSrc.blend(dst) // src (other) over dst (s)
					// frames[0] came from palette.Add() so it is always a valid narrow Unicode scalar value; ErrInvalidColorKey impossible.
					newCk, _ := newPalette.Reserve(newAnimation.frames[0]) // arbitrary key choice; Reserve will give us a unique one
					newMatrix[i-rowOffset][j-colOffset] = newCk
					newAnimationSequences[newCk] = newAnimation
				case dstAnimation:
					// switcher confirmed dstIsColor=false; a well-formed sprite guarantees the
					// key is in animationSequences (absent key would be nil, causing a panic below).
					origSeq := s.animationSequences[s.matrix[i-rowOffset][j-colOffset]]
					// Translate dst frames into the new palette without modifying the original sequence.
					translatedDstFrames := make([]ColorKey, len(origSeq.frames))
					for fi, origCk := range origSeq.frames {
						// invariant: AnimationSequence frames are always valid palette keys; ok is guaranteed.
						aColor, _ := s.palette.Get(origCk)
						translatedDstFrames[fi], _ = newPalette.Add(aColor) // bool (new vs existing) discarded
					}
					// invariant: origSeq.frameDuration > 0 (valid seq); translatedFrames has len ≥ 1; error impossible.
					dstSeq, _ := NewAnimationSequence(newPalette, translatedDstFrames, origSeq.frameDuration)
					// Translate src static color into newPalette for a consistent palette on the result.
					// switcher confirmed srcIsColor=true; ok is guaranteed.
					srcStaticColor, _ := other.palette.Get(other.matrix[i][j])
					srcStaticCk, _ := newPalette.Add(srcStaticColor) // bool (new vs existing) discarded
					// invariant: single valid frame; origSeq.frameDuration > 0; error impossible.
					srcSeq, _ := NewAnimationSequence(newPalette, []ColorKey{srcStaticCk}, origSeq.frameDuration)
					newAnimation := srcSeq.blend(dstSeq) // src (other) over dst (s)
					// frames[0] came from palette.Add() so it is always a valid narrow Unicode scalar value; ErrInvalidColorKey impossible.
					newCk, _ := newPalette.Reserve(newAnimation.frames[0]) // arbitrary key choice; Reserve will give us a unique one
					newMatrix[i-rowOffset][j-colOffset] = newCk
					newAnimationSequences[newCk] = newAnimation
				case bothAnimation:
					// switcher confirmed dstIsColor=false; a well-formed sprite guarantees the
					// key is in animationSequences (absent key would be nil, causing a panic below).
					origSeq := s.animationSequences[s.matrix[i-rowOffset][j-colOffset]]
					// Translate dst frames into the new palette without modifying the original sequence.
					translatedDstFrames := make([]ColorKey, len(origSeq.frames))
					for fi, origCk := range origSeq.frames {
						// invariant: AnimationSequence frames are always valid palette keys; ok is guaranteed.
						aColor, _ := s.palette.Get(origCk)
						translatedDstFrames[fi], _ = newPalette.Add(aColor) // bool (new vs existing) discarded
					}
					// invariant: origSeq.frameDuration > 0 (valid seq); translatedFrames has len ≥ 1; error impossible.
					dstSeq, _ := NewAnimationSequence(newPalette, translatedDstFrames, origSeq.frameDuration)
					// switcher confirmed srcIsColor=false; a well-formed sprite guarantees the
					// key is in animationSequences (absent key would be nil, causing a panic in blend below).
					srcSeq := other.animationSequences[other.matrix[i][j]]
					// Translate src frames into newPalette so blend result lives entirely in newPalette.
					translatedSrcFrames := make([]ColorKey, len(srcSeq.frames))
					for fi, srcCk := range srcSeq.frames {
						// invariant: AnimationSequence frames are always valid palette keys; ok is guaranteed.
						srcFrameColor, _ := other.palette.Get(srcCk)
						translatedSrcFrames[fi], _ = newPalette.Add(srcFrameColor) // bool (new vs existing) discarded
					}
					// invariant: srcSeq.frameDuration > 0 (valid seq); frames non-empty; error impossible.
					translatedSrc, _ := NewAnimationSequence(newPalette, translatedSrcFrames, srcSeq.frameDuration)
					newAnimation := translatedSrc.blend(dstSeq) // src (other) over dst (s)
					// frames[0] came from palette.Add() so it is always a valid narrow Unicode scalar value; ErrInvalidColorKey impossible.
					newCk, _ := newPalette.Reserve(newAnimation.frames[0]) // arbitrary key choice; Reserve will give us a unique one
					newMatrix[i-rowOffset][j-colOffset] = newCk
					newAnimationSequences[newCk] = newAnimation
				}
			}
		}
	}

	return &Sprite{
		matrix:             newMatrix,
		palette:            newPalette,
		animationSequences: newAnimationSequences,
	}, nil
}

func (s *Sprite) ComposeExpanding(other *Sprite) (*Sprite, error) {
	baseWidth := s.Width()
	baseHeight := s.Height()
	overlayWidth := other.Width()
	overlayHeight := other.Height()

	newWidth := max(baseWidth, overlayWidth)
	newHeight := max(baseHeight, overlayHeight)

	// Place the smaller sprite in the middle of the larger sprite
	baseOffsetX := (newWidth - baseWidth) / 2
	baseOffsetY := (newHeight - baseHeight) / 2
	overlayOffsetX := (newWidth - overlayWidth) / 2
	overlayOffsetY := (newHeight - overlayHeight) / 2

	// Build an expanded copy of s without mutating it. Clone gives us a deep
	// copy of the palette and all animation sequences pointing to it.
	expanded := s.Clone()

	// Add a transparent key for border cells. If transparent is already in
	// the palette (e.g. from BlankSprite), Add is a no-op and reuses it.
	transparentCk, _ := expanded.palette.Add(color.RGBA{R: 0, G: 0, B: 0, A: 0}) // bool (new vs existing) discarded — only the key is needed

	// Build the expanded matrix, filling border cells with the transparent key
	// so Compose can look them up without a nil dereference.
	newMatrix := make([][]ColorKey, newHeight)
	for i := range newMatrix {
		newMatrix[i] = make([]ColorKey, newWidth)
		for j := range newMatrix[i] {
			newMatrix[i][j] = transparentCk
		}
	}
	for i := range s.matrix {
		for j := range s.matrix[i] {
			newMatrix[i+baseOffsetY][j+baseOffsetX] = s.matrix[i][j]
		}
	}
	expanded.matrix = newMatrix

	// Compose's semantics: dst[i][j] → src[i+rowOffset][j+colOffset], so
	// src[0][0] lands at dst[-rowOffset][-colOffset]. Negating the offset
	// places the smaller sprite at the center of the larger canvas.
	return expanded.Compose(other, -overlayOffsetY, -overlayOffsetX)
}

// Clone returns a deep copy of the sprite. The clone has identical pixel
// content and animation state, but is fully independent: mutations to
// either sprite's palette, matrix, or animations will not affect the other.
func (s *Sprite) Clone() *Sprite {
	newPalette := s.palette.clone()

	newMatrix := make([][]ColorKey, len(s.matrix))
	for i := range s.matrix {
		newMatrix[i] = make([]ColorKey, len(s.matrix[i]))
		copy(newMatrix[i], s.matrix[i])
	}

	newAnimationSequences := make(map[ColorKey]*AnimationSequence, len(s.animationSequences))
	for k, seq := range s.animationSequences {
		newFrames := make([]ColorKey, len(seq.frames))
		copy(newFrames, seq.frames)
		newAnimationSequences[k] = &AnimationSequence{
			palette:       newPalette,
			frames:        newFrames,
			frameDuration: seq.frameDuration,
			currentFrame:  seq.currentFrame,
			currentTick:   seq.currentTick,
		}
	}

	return &Sprite{
		matrix:             newMatrix,
		palette:            newPalette,
		animationSequences: newAnimationSequences,
	}
}
