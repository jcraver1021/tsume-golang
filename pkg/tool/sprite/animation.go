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

func (a *AnimationSequence) GetColor() color.RGBA {
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

func (a *AnimationSequence) Blend(other *AnimationSequence) *AnimationSequence {
	if a.frameDuration != other.frameDuration {
		// We will decide whether to implement this case later
		panic("frame durations must match to blend animation sequences")
	}

	newFrames := make([]ColorKey, lcm(len(a.frames), len(other.frames)))

	for i := range newFrames {
		frame1 := a.frames[(i % len(a.frames))]
		color1, _ := a.palette.Get(frame1)
		frame2 := other.frames[(i % len(other.frames))]
		color2, _ := other.palette.Get(frame2)
		newColor := alphaComposite(color1, color2)
		newKey, _ := a.palette.Add(newColor)
		newFrames[i] = newKey
	}

	return &AnimationSequence{
		palette:       a.palette,
		frames:        newFrames,
		frameDuration: a.frameDuration,
		currentFrame:  0,
		currentTick:   0,
	}
}
