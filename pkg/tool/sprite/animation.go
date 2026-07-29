package sprite

import (
	"errors"
	"image/color"
)

var (
	ErrInvalidFrameDuration = errors.New("invalid frame duration: must be positive")
	ErrEmptyFrames          = errors.New("frames cannot be empty")
)

type AnimationSequence struct {
	palette       *Palette
	frames        []ColorKey
	frameDuration int
	currentFrame  int
	currentTick   int
}

func NewAnimationSequence(palette *Palette, frames []ColorKey, frameDuration int) (*AnimationSequence, error) {
	if frameDuration <= 0 {
		return nil, ErrInvalidFrameDuration
	}
	if len(frames) == 0 {
		return nil, ErrEmptyFrames
	}

	return &AnimationSequence{
		palette:       palette,
		frames:        frames,
		frameDuration: frameDuration,
		currentFrame:  0,
		currentTick:   0,
	}, nil
}

func (a *AnimationSequence) getColor() color.RGBA {
	// invariant: frames are ColorKeys added to the palette at construction time; ok is guaranteed.
	c, _ := a.palette.Get(a.frames[a.currentFrame])
	return c
}

func (a *AnimationSequence) Advance() {
	// We avoid modulo for performance
	a.currentTick++
	if a.currentTick >= a.frameDuration {
		a.currentTick = 0
		a.currentFrame++
		if a.currentFrame >= len(a.frames) {
			a.currentFrame = 0
		}
	}
}

// blend composites this animation sequence over other using Porter-Duff "src over
// dst" blending and returns the result as a new AnimationSequence.
//
// # Timing alignment
//
// The blended sequence uses frameDuration = GCD(a.frameDuration, other.frameDuration).
// Each source sequence is first expanded to this finer resolution (N frames at
// duration D become N×(D/GCD) frames at GCD), then the two expanded sequences are
// tiled to their LCM length and composited frame-by-frame. After blending, the
// frame slice is reduced to its minimal repeating period: if the full sequence is a
// tiling of a shorter prefix, only that prefix is stored.
//
// # Space complexity
//
// Before period reduction, the blended sequence contains
//
//	LCM(len(a.frames)×a.frameDuration, len(other.frames)×other.frameDuration) / GCD(a.frameDuration, other.frameDuration)
//
// frames. When the two frame durations are coprime (GCD = 1), this degenerates to
// full tick-level expansion and the frame count equals the product of the two cycle
// lengths in ticks. For predictable space use, prefer frame durations that share
// common factors — powers of 2, or multiples of a common base, work well. Coprime
// durations are permitted but may produce unexpectedly large blended sequences.
func (a *AnimationSequence) blend(other *AnimationSequence) *AnimationSequence {
	// Step 1: GCD-normalize frame durations so both sequences advance at the same rate.
	newDuration := gcd(a.frameDuration, other.frameDuration)

	aExpansion := a.frameDuration / newDuration
	aExpanded := make([]ColorKey, len(a.frames)*aExpansion)
	for i, ck := range a.frames {
		for k := 0; k < aExpansion; k++ {
			aExpanded[i*aExpansion+k] = ck
		}
	}

	bExpansion := other.frameDuration / newDuration
	bExpanded := make([]ColorKey, len(other.frames)*bExpansion)
	for i, ck := range other.frames {
		for k := 0; k < bExpansion; k++ {
			bExpanded[i*bExpansion+k] = ck
		}
	}

	// Step 2: blend the two expanded sequences tiled to their LCM length.
	blendedLen := lcm(len(aExpanded), len(bExpanded))
	newFrames := make([]ColorKey, blendedLen)
	for i := range newFrames {
		ck1 := aExpanded[i%len(aExpanded)]
		color1, _ := a.palette.Get(ck1) // invariant: frame keys are always in the palette; ok is guaranteed.
		ck2 := bExpanded[i%len(bExpanded)]
		color2, _ := other.palette.Get(ck2) // invariant: same.
		newColor := alphaComposite(color1, color2)
		newKey, _ := a.palette.Add(newColor) // bool (new vs existing) discarded
		newFrames[i] = newKey
	}

	// Step 3: shorten to the minimal repeating period.
	newFrames = minimalPeriod(newFrames)

	return &AnimationSequence{
		palette:       a.palette,
		frames:        newFrames,
		frameDuration: newDuration,
		currentFrame:  0,
		currentTick:   0,
	}
}

// minimalPeriod returns the shortest prefix of frames that tiles the full slice.
// Only divisors of len(frames) are tested. If no shorter period exists, the
// original slice is returned unchanged.
func minimalPeriod(frames []ColorKey) []ColorKey {
	n := len(frames)
	for p := 1; p < n; p++ {
		if n%p != 0 {
			continue
		}
		if isTiledBy(frames, p) {
			return frames[:p]
		}
	}
	return frames
}

// isTiledBy reports whether frames is a repetition of its first p elements.
func isTiledBy(frames []ColorKey, p int) bool {
	for i := p; i < len(frames); i++ {
		if frames[i] != frames[i%p] {
			return false
		}
	}
	return true
}
