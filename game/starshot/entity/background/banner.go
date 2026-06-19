package background

import (
	"image/color"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/util"
)

// Banner displays static text on the screen
type Banner struct {
	// Content
	text      string
	fontFace  *text.GoTextFace
	textColor color.RGBA

	// Layout
	x, y int // Position (top-left corner)

	// Optional background
	bgColor *color.RGBA // nil = no background box
	padding int

	// Lifetime
	frameCount int
	maxFrames  int // 0 = permanent, >0 = auto-remove after duration

	// Cached dimensions
	width, height float64
}

// BannerOptions provides advanced configuration for banner creation
type BannerOptions struct {
	Text            string
	X, Y            int
	FontSize        float64
	TextColor       color.RGBA
	BackgroundColor *color.RGBA
	Padding         int
	DurationFrames  int // 0 = permanent
}

// NewBanner creates a simple permanent banner centered at the given position
func NewBanner(text string, x, y int, fontSize float64, c color.RGBA) (*Banner, error) {
	return NewBannerWithOptions(BannerOptions{
		Text:           text,
		X:              x,
		Y:              y,
		FontSize:       fontSize,
		TextColor:      c,
		DurationFrames: 0, // permanent
	})
}

// NewBannerWithOptions creates a banner with full configuration options
func NewBannerWithOptions(opts BannerOptions) (*Banner, error) {
	fm, err := util.GetDefaultFontManager()
	if err != nil {
		return nil, err
	}

	face := fm.GetFace(opts.FontSize)

	// Measure text dimensions for caching
	width, height := text.Measure(opts.Text, face, 0)

	return &Banner{
		text:       opts.Text,
		fontFace:   face,
		textColor:  opts.TextColor,
		x:          opts.X,
		y:          opts.Y,
		bgColor:    opts.BackgroundColor,
		padding:    opts.Padding,
		maxFrames:  opts.DurationFrames,
		frameCount: 0,
		width:      width,
		height:     height,
	}, nil
}

// Entity interface implementation

func (b *Banner) Type() def.EntityType {
	return def.EntityTypeBackground
}

func (b *Banner) Location() (x, y int) {
	return b.x, b.y
}

func (b *Banner) Dimensions() (width, height int) {
	return int(b.width), int(b.height)
}

func (b *Banner) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (b *Banner) Act(scene def.Scene) {
	if b.maxFrames > 0 {
		b.frameCount++
	}
}

func (b *Banner) Draw(img *ebit.Image) {
	if b.bgColor != nil {
		b.drawBackgroundBox(img)
	}

	opts := &text.DrawOptions{}
	opts.GeoM.Translate(float64(b.x)-b.width/2, float64(b.y))
	opts.ColorScale.ScaleWithColor(b.textColor)
	text.Draw(img, b.text, b.fontFace, opts)
}

func (b *Banner) drawBackgroundBox(img *ebit.Image) {
	x1 := int(float64(b.x)-b.width/2) - b.padding
	y1 := b.y - b.padding
	x2 := int(float64(b.x)+b.width/2) + b.padding
	y2 := int(float64(b.y)+b.height) + b.padding

	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			img.Set(x, y, *b.bgColor)
		}
	}
}

func (b *Banner) CanBeRemoved() bool {
	if b.maxFrames == 0 {
		return false
	}

	return b.frameCount >= b.maxFrames
}
