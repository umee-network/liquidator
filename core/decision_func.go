package core

import "github.com/knadh/koanf"

// DecisionFunc should evaluate an estimated liquidation outcome input as
// a LiquidationOrder and return true if the liquidation should be attempted.
// False returns are for liquidations that should be skipped. The behavior of
// DecisionFunc can depend on the fields of the input configuration, and
// should also monitor the liquidator's available gas and current balances of
// relevant tokens.
type DecisionFunc func(*koanf.Koanf, LiquidationOrder) (bool, error)

// EmptyDecisionFunc is a DecisionFunc that says no to all potential liquidations
func EmptyDecisionFunc(_ *koanf.Koanf, _ LiquidationOrder) (bool, error) {
	return false, nil
}
