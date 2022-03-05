package base

import "github.com/knadh/koanf"

// TargetFunc should return a list of potential liquidation targets, as
// well as their current borrowed and collateral balances in all token
// denominations. Collateral uTokens must be expressed as their equivalent
// value in base tokens.
type TargetFunc func(*koanf.Koanf) ([]LiquidationTarget, error)

// EmptyTargetFunc is a TargetFunc that never returns any liquidation targets
func EmptyTargetFunc(_ *koanf.Koanf) ([]LiquidationTarget, error) {
	return nil, nil
}
