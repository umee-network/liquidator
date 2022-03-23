package liquidator

import (
	"context"
)

// defaultGetLiquidationTargets queries the chain for all eligible liquidation
// targets and their total borrowed and collateral, converting collateral uTokens
// to equivalent base tokens.
func defaultGetLiquidationTargets(ctx context.Context) ([]LiquidationTarget, error) {
	// TODO: body
	return nil, nil
}

// defaultSelectLiquidationDenoms receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral. From there it should decide what borrow
// denominations the liquidator is interested in repaying, and what collateral
// rewards it wants to receive. For example, a liquidation target with three
// borrowed denominations and two collateral denominations will have six possible
// combinations of (repay denom, reward denom) that could be made into liquidation
// orders. The liquidator may not possess every repayment denom, or be interested
// in every collateral denom. Furthermore, if one liquidation brings the borrower
// back to health, then the remaining ones will no longer be available. The
// defaultSelectLiquidationDenoms function should choose the reward and repay denoms from
// available options. Also returns a boolean, which can be set to false if no
// workable denominations were discovered.
func defaultSelectLiquidationDenoms(ctx context.Context, target LiquidationTarget) (LiquidationOrder, bool, error) {
	// TODO: body
	return LiquidationOrder{}, false, nil
}

// defaultEstimateLiquidationOutcome simulates the result of a MsgLiquidate in selected
// denominations as closely as possible to how it would be executed by the leverage
// module on the umee chain.
func defaultEstimateLiquidationOutcome(ctx context.Context, intent LiquidationOrder) (LiquidationOrder, error) {
	// TODO: body
	return LiquidationOrder{}, nil
}

// defaultApproveLiquidation decides whether an estimated liquidation outcome is worth it
func defaultApproveLiquidation(ctx context.Context, estimate LiquidationOrder) (bool, error) {
	// TODO: body
	return false, nil
}

// defaultExecuteLiquidation attempts to execute a chosen liquidation, and reports back
// the actual repaid and reward amounts if successful
func defaultExecuteLiquidation(ctx context.Context, intent LiquidationOrder) (LiquidationOrder, error) {
	// TODO: body
	return LiquidationOrder{}, nil
}
