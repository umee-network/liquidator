package core

import "github.com/knadh/koanf"

// OrderingFunc receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral. From there it should decide what borrow
// denominations the liquidator is interested in repaying, and what collateral
// rewards it will receive. For example, a liquidation target with three
// borrowed denominations and two collateral denominations will have six possible
// combinations of (repay denom, reward denom) that could be made into liquidation
// orders. The liquidator may not possess every repayment denom, or be interested
// in every collateral denom. Furthermore, if one liquidation brings the borrower
// back to health, then the remaining ones will no longer be available. The
// OrderingFunc should discard unavailable or undesirable possibilities and list
// the remaining potential liquidations in order of priority. Custom implementations
// may also choose to select or skip targets based on address, for example to
// avoid the downstream computational cost of recommending the same target
// many times in a row. OrderingFunc behavior should depend heavily on the fields
// of the input configuration.
type OrderingFunc func(*koanf.Koanf, LiquidationTarget) ([]LiquidationOrder, error)

// EmptyOrderingFunc is an OrderingFunc that never returns any Liquidation Orders
func EmptyOrderingFunc(_ *koanf.Koanf, _ LiquidationTarget) ([]LiquidationOrder, error) {
	return nil, nil
}
