package ui

import (
	"fmt"
	"image/color"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"tsumegolang/game/starshot/def"
)

const (
	hudHeight = 20
	hudY      = 4 // vertical text offset within the banner
)

// ammoReader is satisfied by *player.Player.
type ammoReader interface {
	SecondaryAmmo() (current, max int, hasWeapon bool)
}

// HUD is a persistent UI entity that renders the status banner at the top of
// the screen. It reads wave and score from a GameStateReader and queries the
// scene each frame for the player's secondary-weapon ammo.
type HUD struct {
	state def.GameStateReader

	// cached each Act so Draw has current values without a scene reference
	secondaryAmmo    int
	secondaryAmmoMax int
	hasSecondary     bool
}

func NewHUD(state def.GameStateReader) *HUD {
	return &HUD{state: state}
}

// --- Entity interface ---

func (h *HUD) Type() def.EntityType {
	return def.EntityTypeUI
}

func (h *HUD) Location() (int, int) {
	return 0, 0
}

func (h *HUD) Dimensions() (int, int) {
	return def.ScreenWidth, hudHeight
}

func (h *HUD) BoundingBoxOverlaps(_ def.Entity) bool {
	return false
}

func (h *HUD) CanBeRemoved() bool {
	return false
}

func (h *HUD) Act(scene def.Scene) {
	for _, p := range scene.Entities().Get(def.EntityTypePlayer) {
		if ar, ok := p.(ammoReader); ok {
			h.secondaryAmmo, h.secondaryAmmoMax, h.hasSecondary = ar.SecondaryAmmo()
			break
		}
	}
}

func (h *HUD) Draw(img *ebit.Image) {
	// Dark translucent banner background
	vector.FillRect(img, 0, 0, float32(def.ScreenWidth), float32(hudHeight),
		color.RGBA{0, 0, 16, 200}, false)

	wave := fmt.Sprintf("Wave %d", h.state.GetWave())
	score := fmt.Sprintf("Score: %d", h.state.GetScore())

	var ammo string
	if h.hasSecondary {
		ammo = fmt.Sprintf("Ammo: %d/%d", h.secondaryAmmo, h.secondaryAmmoMax)
	} else {
		ammo = "Ammo: --"
	}

	ebitenutil.DebugPrintAt(img, wave, 10, hudY)
	ebitenutil.DebugPrintAt(img, score, def.ScreenWidth/2-len(score)*3, hudY)
	ebitenutil.DebugPrintAt(img, ammo, def.ScreenWidth-len(ammo)*6-10, hudY)
}
