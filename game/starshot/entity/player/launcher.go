package player

import (
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/projectile"
)

const (
	bombLauncherCooldownFrames = 20 // minimum frames between shots when holding multiple ammo
	bombLauncherMaxAmmo        = 3
	bombLauncherStartAmmo      = 1
)

// Launcher is the common struct for area-effect weapons that consume ammo.
// Different launcher variants are created by different constructors; the
// projectile spawned on each shot varies via the fireFunc field.
type Launcher struct {
	sprite         *draw.ColorMatrix
	cooldown       int
	cooldownFrames int
	ammo           int
	maxAmmo        int
	mountY         int
	fireFunc       func(x, y int, scene def.Scene)
}

// NewBombLauncher returns a slow heavy-bomb launcher that starts with one ammo.
// Call Reload to replenish ammo (e.g. from a pickup entity).
func NewBombLauncher() (*Launcher, error) {
	data, err := spriteFiles.ReadFile("sprites/launcher_bomb.yaml")
	if err != nil {
		return nil, err
	}
	sprite, err := draw.ColorMatrixFromBytes(data)
	if err != nil {
		return nil, err
	}
	return &Launcher{
		sprite:         sprite,
		cooldownFrames: bombLauncherCooldownFrames,
		ammo:           bombLauncherStartAmmo,
		maxAmmo:        bombLauncherMaxAmmo,
		mountY:         8,
		fireFunc: func(x, y int, scene def.Scene) {
			scene.Entities().Add(projectile.NewBomb(x, y))
		},
	}, nil
}

// Add more launcher types here as needed

func (l *Launcher) TickCooldown() {
	if l.cooldown > 0 {
		l.cooldown--
	}
}

// Ready returns true only when ammo is available and the inter-shot cooldown
// has expired. A launcher with zero ammo stays locked until Reload is called.
func (l *Launcher) Ready() bool {
	return l.cooldown == 0 && l.ammo > 0
}

func (l *Launcher) Fire(originX, originY int, scene def.Scene) {
	l.fireFunc(originX, originY, scene)
	l.ammo--
	l.cooldown = l.cooldownFrames
}

// Reload adds up to count ammo, capped at maxAmmo.
func (l *Launcher) Reload(count int) {
	l.ammo += count
	if l.ammo > l.maxAmmo {
		l.ammo = l.maxAmmo
	}
}

// Ammo returns the current ammo count (for HUD display).
func (l *Launcher) Ammo() int { return l.ammo }

// MaxAmmo returns the maximum ammo capacity.
func (l *Launcher) MaxAmmo() int { return l.maxAmmo }

func (l *Launcher) Sprite() *draw.ColorMatrix { return l.sprite }

func (l *Launcher) MountOffsetX(hullWidth int) int {
	return (hullWidth - l.sprite.Width()) / 2
}

func (l *Launcher) MountOffsetY() int { return l.mountY }
