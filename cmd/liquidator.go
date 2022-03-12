package cmd

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LiquidationTarget contains the address of a borrower who is over their borrow limit, as
// well as their current borrowed amounts and collateral in all denominations.
// CONTRACT: Collateral uTokens must be expressed as their equivalent value in base tokens.
type LiquidationTarget struct {
	Addr       sdk.AccAddress
	Borrowed   sdk.Coins
	Collateral sdk.Coins
}

// LiquidationOrder expresses the intent to perform, or the outcome of, a liquidation.
// Collateral reward amounts are always expressed in base tokens equivalent to their
// underlying uToken amount.
type LiquidationOrder struct {
	Addr   sdk.AccAddress
	Repay  sdk.Coin
	Reward sdk.Coin
}

// startLiquidator causes the liquidator to continuously look for liquidation targets, decide on which
// borrowed and collateral denominations to attempt to liquidate, and attempt to execute any
// liquidations whose estimated outcomes are approved by its configured decisionmaking. Returns
// only when context is cancelled.
func startLiquidator(
	ctx context.Context,
	cancelFunc context.CancelFunc,
) error {

	// loop as long as ctx is not cancelled
	for ctx.Err() == nil {
		// get a list of eligible liquidation targets
		targets, err := getLiquidationTargets(ctx)
		if err != nil {
			logger.Err(err)
			continue
		}

		for _, target := range targets {
			if ctx.Err() == nil {
				// select one reward denom and one repay denom to consider on target address
				intent, ok, err := selectLiquidationDenoms(ctx, target)
				if err != nil {
					logger.Err(err)
					continue
				}
				if !ok {
					continue
				}
				// estimate actual liquidation outcome if chosen denoms were to be liquidated
				estimate, err := estimateLiquidationOutcome(ctx, intent)
				if err != nil {
					logger.Err(err)
					continue
				}
				// decide whether to liquidate based on estimated outcome
				ok, err = decisionFunc(ctx, estimate)
				if err != nil {
					logger.Err(err)
					continue
				}
				if !ok {
					continue
				}
				// attempt liquidation if it was approved by decisionFunc
				outcome, err := executeLiquidation(ctx, intent)
				if err != nil {
					logger.Err(err)
					continue
				}
				logger.Info().Msgf(
					"LIQUIDATION SUCCESS: target: %s repaid %s reward%s",
					outcome.Addr.String(),
					outcome.Repay.String(),
					outcome.Reward.String(),
				)
			}
		}

		// Sleep at the end of each cycle
		time.Sleep(time.Minute)
	}
	return ctx.Err()
}

// return a list of potential liquidation targets, as
// well as their current borrowed and collateral balances in all token
// denominations. Collateral uTokens are be expressed as their equivalent
// value in base tokens.
func getLiquidationTargets(ctx context.Context) ([]LiquidationTarget, error) {
	// TODO: body
	return nil, nil
}

// selectLiquidationDenoms receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral. From there it should decide what borrow
// denominations the liquidator is interested in repaying, and what collateral
// rewards it wants to receive. For example, a liquidation target with three
// borrowed denominations and two collateral denominations will have six possible
// combinations of (repay denom, reward denom) that could be made into liquidation
// orders. The liquidator may not possess every repayment denom, or be interested
// in every collateral denom. Furthermore, if one liquidation brings the borrower
// back to health, then the remaining ones will no longer be available. The
// selectLiquidationDenoms function should choose the reward and repay denoms from
// available options. Also returns a boolean, which can be set to false if no
// workable denominations were discovered.
func selectLiquidationDenoms(ctx context.Context, target LiquidationTarget) (LiquidationOrder, bool, error) {
	// TODO: body
	return LiquidationOrder{}, false, nil
}

// estimateLiquidationOutcome simulates the result of a MsgLiquidate in selected
// denominations as closely as possible to how it would be executed by the leverage
// module on the umee chain.
func estimateLiquidationOutcome(ctx context.Context, intent LiquidationOrder) (LiquidationOrder, error) {
	// TODO: body
	return LiquidationOrder{}, nil
}

// decisionFunc decides whether an estimated liquidation outcome is worth it
func decisionFunc(ctx context.Context, estimate LiquidationOrder) (bool, error) {
	// TODO: body
	return false, nil
}

// executeLiquidation attempts to execute a chosen liquidation, and reports back
// the actual repaid and reward amounts if successful
func executeLiquidation(ctx context.Context, intent LiquidationOrder) (LiquidationOrder, error) {
	// TODO: body
	return LiquidationOrder{}, nil
}
