package player

// Example power-up functions showing how to modify player configuration

// UpgradeToCentralCannon adds a nose-mounted cannon to the player
func (p *Player) UpgradeToCentralCannon() {
	// Check if already has central cannon
	for _, comp := range p.components {
		if comp.Name == "central_cannon" {
			return // Already equipped
		}
	}
	p.AddComponent(CentralCannon())
}

// UpgradeToWingGuns adds wing-mounted gun pods to the player
func (p *Player) UpgradeToWingGuns() {
	// Check if already has wing guns
	for _, comp := range p.components {
		if comp.Name == "wing_guns" {
			return // Already equipped
		}
	}
	p.AddComponent(WingGuns())
}

// RemoveWeaponUpgrades removes all weapon components
func (p *Player) RemoveWeaponUpgrades() {
	p.RemoveComponent("central_cannon")
	p.RemoveComponent("wing_guns")
}

// Example: How to create a fully upgraded player
// player := NewPlayer(x, y)
// player.UpgradeToCentralCannon()
// player.UpgradeToWingGuns()
//
// The visual representation will automatically show:
// - Core hull with cockpit
// - Animated engine glow (pulsing)
// - Orange central cannon on nose
// - Red/orange gun pods on wings
