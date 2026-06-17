package player

// implements the player entity

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

const (
	defaultPlayerSpeed      = 5
	defaultPlayerSideLength = 32
)

type PlayerAction struct {
	MoveUp    bool
	MoveDown  bool
	MoveLeft  bool
	MoveRight bool
	Shoot     bool
}

type Player struct {
	x      int
	y      int
	width  int
	height int
	speed  int

	playerAction PlayerAction

	// Sprite composition
	components     []*SpriteComponent
	animatedPixels []AnimatedPixel

	// Cache tick from Act() for use in Draw()
	currentTick int
}

func NewPlayer(x, y int) *Player {
	// Default configuration: core hull + basic engine
	components := []*SpriteComponent{
		CoreHull(),
		BasicEngine(),
	}

	animatedPixels := BasicEngineAnimatedPixels()

	return &Player{
		x:              x,
		y:              y,
		width:          defaultPlayerSideLength,
		height:         defaultPlayerSideLength,
		speed:          defaultPlayerSpeed,
		components:     components,
		animatedPixels: animatedPixels,
	}
}

// AddComponent adds a sprite component to the player (for power-ups)
func (p *Player) AddComponent(component *SpriteComponent) {
	p.components = append(p.components, component)
}

// RemoveComponent removes a sprite component by name
func (p *Player) RemoveComponent(name string) {
	for i, comp := range p.components {
		if comp.Name == name {
			p.components = append(p.components[:i], p.components[i+1:]...)
			return
		}
	}
}

// AddAnimatedPixels adds animated pixels (e.g., for weapon effects)
func (p *Player) AddAnimatedPixels(pixels []AnimatedPixel) {
	p.animatedPixels = append(p.animatedPixels, pixels...)
}

func (p *Player) Type() def.EntityType {
	return def.EntityTypePlayer
}

func (p *Player) Location() (x, y int) {
	return p.x, p.y
}

func (p *Player) Dimensions() (width, height int) {
	return p.width, p.height
}

func (p *Player) Onscreen(b def.Scene) def.OnScreen {
	if p.x+p.width < 0 || p.x > b.Width() || p.y+p.height < 0 || p.y > b.Height() {
		return def.OffScreen
	}
	if p.x >= 0 && p.x+p.width <= b.Width() && p.y >= 0 && p.y+p.height <= b.Height() {
		return def.Fully
	}
	return def.Partially
}

func (p *Player) Overlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()

	return !(p.x+p.width < ox || p.x > ox+ow || p.y+p.height < oy || p.y > oy+oh)
}

func (p *Player) SetPlayerAction(action PlayerAction) {
	p.playerAction = action
}

func (p *Player) Act(b def.Scene) {
	// Cache global tick for use in Draw()
	p.currentTick = b.Tick()

	if p.playerAction.MoveUp {
		p.y -= p.speed
	}
	if p.playerAction.MoveDown {
		p.y += p.speed
	}
	if p.playerAction.MoveLeft {
		p.x -= p.speed
	}
	if p.playerAction.MoveRight {
		p.x += p.speed
	}

	// Clamp to screen bounds
	if p.x < 0 {
		p.x = 0
	}
	if p.y < 0 {
		p.y = 0
	}
	if p.x+p.width > b.Width() {
		p.x = b.Width() - p.width
	}
	if p.y+p.height > b.Height() {
		p.y = b.Height() - p.height
	}

	p.playerAction = PlayerAction{} // Reset actions after processing
}

func (p *Player) Draw(img *ebit.Image) {
	palette := GetColorPalette()

	// Draw all components in order
	for _, component := range p.components {
		for row := range component.Height {
			for col := range component.Width {
				if row < len(component.Data) && col < len(component.Data[row]) {
					colorCode := component.Data[row][col]
					if colorCode != ColorEmpty {
						screenX := p.x + component.OffsetX + col
						screenY := p.y + component.OffsetY + row
						img.Set(screenX, screenY, palette[colorCode])
					}
				}
			}
		}
	}

	// Draw animated pixels on top using cached tick
	for _, animPixel := range p.animatedPixels {
		colorCode := animPixel.Sequence.GetColorAtFrame(p.currentTick)
		if colorCode != ColorEmpty {
			img.Set(p.x+animPixel.X, p.y+animPixel.Y, palette[colorCode])
		}
	}
}

func (p *Player) CanBeRemoved() bool {
	return false
}
