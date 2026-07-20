package player

import (
	"path"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
)

// PlayerAction represents player input state
type PlayerAction struct {
	MoveUp         bool
	MoveDown       bool
	MoveLeft       bool
	MoveRight      bool
	ShootPrimary   bool
	ShootSecondary bool
}

// PlayerController is an interface for entities that can respond to player input
// and have their weapons swapped at runtime.
type PlayerController interface {
	def.Entity
	SetPlayerAction(action PlayerAction)
	SetPrimaryWeapon(weapon def.Weapon)
	SetSecondaryWeapon(weapon def.Weapon)
}

const defaultPlayerSpeed = 5

// weaponSprite is implemented by weapons that provide a hull-composited visual.
// MountOffsetX returns the x pixel offset (from the hull's left edge) at which
// the weapon sprite should be composited — weapons declare their own position so
// a center cannon and a wing-mounted gun can coexist without overlapping.
type weaponSprite interface {
	Sprite() *draw.ColorMatrix
	MountOffsetX(hullWidth int) int
	MountOffsetY() int
}

type Player struct {
	x, y          int
	width, height int

	hull   *Hull
	engine *Engine
	sprite *draw.ColorMatrix

	playerAction    PlayerAction
	primaryWeapon   def.Weapon
	secondaryWeapon def.Weapon

	hp    int
	maxHP int

	dead                 bool
	explosionFrameCount  int // Frames since death
	explosionMaxDuration int // Total frames the explosion animation lasts
}

// NewPlayer creates a new ColorMatrix-based player. Either weapon may be nil.
func NewPlayer(x, y int, primaryWeapon, secondaryWeapon def.Weapon) (*Player, error) {
	p := &Player{
		x:               x,
		y:               y,
		primaryWeapon:   primaryWeapon,
		secondaryWeapon: secondaryWeapon,
	}

	// Load defaults
	var err error
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

	// Set dimensions based on the composed sprite
	p.width = p.sprite.Width()
	p.height = p.sprite.Height()

	// Derive starting HP from hull. Future upgrades call AddMaxHP.
	p.maxHP = p.hull.HP
	p.hp = p.maxHP

	return p, nil
}

// AddMaxHP increases the player's maximum (and current) HP by the given amount.
// Call this when equipping a new hull component or upgrade.
func (p *Player) AddMaxHP(bonus int) {
	p.maxHP += bonus
	p.hp += bonus
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

	// Compose each weapon's sprite onto the hull at its declared mount position.
	for _, w := range []def.Weapon{p.primaryWeapon, p.secondaryWeapon} {
		if ws, ok := w.(weaponSprite); ok {
			gunSprite := ws.Sprite()
			if gunSprite != nil {
				_ = hull.Compose(gunSprite, ws.MountOffsetX(hull.Width()), ws.MountOffsetY())
			}
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

func (p *Player) SetPrimaryWeapon(weapon def.Weapon) {
	p.primaryWeapon = weapon
}

func (p *Player) SetSecondaryWeapon(weapon def.Weapon) {
	p.secondaryWeapon = weapon
}

func (p *Player) Act(b def.Scene) {
	if p.dead {
		// Track explosion animation progress
		p.explosionFrameCount++
		return
	}

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

	// Handle shooting via each equipped weapon
	originX := p.x + p.width/2 - 1
	originY := p.y - 8
	fireWeapon(p.primaryWeapon, p.playerAction.ShootPrimary, originX, originY, b)
	fireWeapon(p.secondaryWeapon, p.playerAction.ShootSecondary, originX, originY, b)

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
	if !p.dead {
		return false
	}
	// Remove player after explosion animation completes
	return p.explosionFrameCount >= p.explosionMaxDuration
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

// Mortal interface implementation

func (p *Player) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		ExplosionSize:      def.ExplosionLarge,
		SlowdownMultiplier: 0.3, // 30% speed
		SlowdownDuration:   90,  // ~1.5 seconds at 60 TPS
	}
}

func (p *Player) MarkAsDead(scene def.Scene) {
	p.dead = true
	p.explosionFrameCount = 0
	// Note: explosion composition happens via ComposeExplosion() called externally
}

// ComposeExplosion overlays an explosion sprite on the player
// Called by the game logic after loading the sprite from the effects package
func (p *Player) ComposeExplosion(explosionSprite *draw.ColorMatrix) error {
	// Store original dimensions before composing
	oldWidth := p.width
	oldHeight := p.height

	// Compose explosion over player sprite (expanding if needed)
	if err := p.sprite.ComposeExpanding(explosionSprite); err != nil {
		return err
	}

	// Update dimensions to match new sprite size
	p.width = p.sprite.Width()
	p.height = p.sprite.Height()

	// Recenter player position so explosion is centered on where player was
	// (sprite grew, so we need to shift position back)
	centerShiftX := (p.width - oldWidth) / 2
	centerShiftY := (p.height - oldHeight) / 2
	p.x -= centerShiftX
	p.y -= centerShiftY

	// Set explosion duration (96 frames for large explosion: 8 frames × 12 ticks/frame)
	p.explosionMaxDuration = 96

	return nil
}

func (p *Player) IsDead() bool { return p.dead }

// --- Damageable ---

func (p *Player) TakeDamage(amount int) {
	if p.dead {
		return
	}
	p.hp -= amount
	if p.hp <= 0 {
		p.hp = 0
		p.dead = true
	}
}

func (p *Player) CurrentHP() int { return p.hp }
func (p *Player) MaxHP() int     { return p.maxHP }

func fireWeapon(w def.Weapon, triggered bool, originX, originY int, scene def.Scene) {
	if w == nil {
		return
	}
	w.TickCooldown()
	if triggered && w.Ready() {
		w.Fire(originX, originY, scene)
	}
}
