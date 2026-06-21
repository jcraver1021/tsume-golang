package player

import (
	"path"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
)


// PlayerAction represents player input state
type PlayerAction struct {
	MoveUp    bool
	MoveDown  bool
	MoveLeft  bool
	MoveRight bool
	Shoot     bool
}

// PlayerController is an interface for entities that can respond to player input
type PlayerController interface {
	def.Entity
	SetPlayerAction(action PlayerAction)
}

const defaultPlayerSpeed = 5

type Player struct {
	x, y          int
	width, height int
	
	hull *Hull
	engine *Engine
	sprite *draw.ColorMatrix

	playerAction PlayerAction
}

// NewPlayer creates a new ColorMatrix-based player
func NewPlayer(x, y int) (*Player, error) {
	// Load core hull
	hullData, err := spriteFiles.ReadFile("sprites/core_hull.yaml")
	if err != nil {
		return nil, err
	}

	hull, err := draw.ColorMatrixFromBytes(hullData)
	if err != nil {
		return nil, err
	}

	// Load basic engine
	engineData, err := spriteFiles.ReadFile("sprites/basic_engine.yaml")
	if err != nil {
		return nil, err
	}

	engine, err := draw.ColorMatrixFromBytes(engineData)
	if err != nil {
		return nil, err
	}

	// Compose hull + engine (engine overlays hull)
	if err := hull.Compose(engine, 0, 0); err != nil {
		return nil, err
	}

	width, height := hull.Dimensions()

	p := &Player{
		x:      x,
		y:      y,
		width:  width,
		height: height,
	}

	// Load defaults
	p.hull, err = loadDefaultHull()
	if err != nil {
		return nil, err
	}

	p.engine, err = loadDefaultEngine()
	if err != nil {
		return nil, err
	}

	// Compose sprites
	p.composePlayerSprites()

	return p, nil
}

func loadDefaultHull() (*Hull, error) {
	return BasicHull()
}

func loadDefaultEngine() (*Engine, error) {
	return BasicEngine()
}

func (p *Player) composePlayerSprites() error {
	// Load hull
	if p.hull == nil {
		return nil
	}
	hull := p.hull.sprite

	// Compose hull + engine (engine overlays hull)
	if p.engine != nil {
		offsetX, offsetY := p.computeEngineMountOffset()
		if err := hull.Compose(p.engine.sprite, offsetX, offsetY); err != nil {
			return err
		}
	}

	p.sprite = hull
	return nil
}

func (p *Player) computeEngineMountOffset() (offsetX, offsetY int) {
	if p.engine == nil {
		return 0, 0
	}

	switch p.engine.EngineMount {
	case EngineMountCenter:
		offsetX = (p.width - p.engine.sprite.Width()) / 2
		offsetY = p.height - p.engine.sprite.Height()
	default:
		offsetX = 0
		offsetY = 0
	}

	return offsetX, offsetY
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

func (p *Player) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(p.x+p.width < ox || p.x > ox+ow || p.y+p.height < oy || p.y > oy+oh)
}

func (p *Player) SetPlayerAction(action PlayerAction) {
	p.playerAction = action
}

func (p *Player) Act(b def.Scene) {
	if p.playerAction.MoveUp {
		p.y -= p.engine.vUp
	}
	if p.playerAction.MoveDown {
		p.y += p.engine.vDown
	}
	if p.playerAction.MoveLeft {
		p.x -= p.engine.vLeft
	}
	if p.playerAction.MoveRight {
		p.x += p.engine.vRight
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
	// Render sprite (advances animations automatically)
	pixels := p.sprite.Render()

	for row := range pixels {
		for col := range pixels[row] {
			color := pixels[row][col]
			if color.A > 0 { // Only draw non-transparent pixels
				img.Set(p.x+col, p.y+row, color)
			}
		}
	}
}

func (p *Player) CanBeRemoved() bool {
	return false
}

// AddComponent allows dynamic composition of power-ups
func (p *Player) AddComponent(componentPath string) error {
	data, err := spriteFiles.ReadFile(path.Join("sprites", componentPath))
	if err != nil {
		return err
	}

	component, err := draw.ColorMatrixFromBytes(data)
	if err != nil {
		return err
	}

	// Compose the new component onto the existing sprite
	return p.sprite.Compose(component, 0, 0)
}
