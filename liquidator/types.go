package liquidator

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LiquidationTarget contains the address of a borrower who is over their borrow limit, as
// well as their current borrowed amounts and collateral in all denominations.
// Collateral uTokens MUST be expressed as their equivalent value in base tokens.
type LiquidationTarget struct {
	Addr       sdk.AccAddress
	Borrowed   sdk.Coins
	Collateral sdk.Coins
}

// LiquidationOrder expresses the intent to perform, or the outcome of, a liquidation.
// Collateral reward amounts MUST be expressed in base tokens equivalent to their
// underlying uToken amount.
type LiquidationOrder struct {
	Addr   sdk.AccAddress
	Repay  sdk.Coin
	Reward sdk.Coin
}

// TargetFunc must return a list of eligible liquidation targets.
type TargetFunc func(context.Context) ([]LiquidationTarget, error)

// SelectFunc must convert a liquidation target to a desired liquidation order by selecting
// reward and repay denominations. It should return false if no available liquidation is desired.
// The coin amounts in the returned order will be treated as maximum amounts in liquidation
// outcome estimation and actual execution.
type SelectFunc func(context.Context, LiquidationTarget) (LiquidationOrder, bool, error)

// EstimateFunc must take a liquidation order representing the liquidator's intent
// and estimate what the actual outcome of the transaction would be when processed
// by leverage module. It returns the estimated repay and reward amounts.
type EstimateFunc func(context.Context, LiquidationOrder) (LiquidationOrder, error)

// ApproveFunc must take a liquidation order representing the estimated outcome
// of a transaction, and return true if the order should be executed.
type ApproveFunc func(context.Context, LiquidationOrder) (bool, error)

// ExecuteFunc must take a liquidation order an execute it on the blockchain. The
// input order should represent the liquidator's intent (not the potentially
// lower estimated outcome). Returns the actual reward and repayment amounts.
type ExecuteFunc func(context.Context, LiquidationOrder) (LiquidationOrder, error)
