package liquidator

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
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
func StartLiquidator(
	ctx context.Context,
	cancelFunc context.CancelFunc,
	konfig *koanf.Koanf,
	logger *zerolog.Logger,
	password string,
) error {
	if err := validateConfig(konfig); err != nil {
		return err
	}

	ticker := time.NewTicker(konfig.Duration("liquidator.interval"))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return err
			}

		case <-ticker.C:
			sweepLiquidations(ctx, cancelFunc, konfig, logger, password)
		}
	}
}

func sweepLiquidations(
	ctx context.Context,
	cancelFunc context.CancelFunc,
	konfig *koanf.Koanf,
	logger *zerolog.Logger,
	password string,
) {
	// get a list of eligible liquidation targets
	targets, err := getLiquidationTargets(ctx)
	if err != nil {
		logger.Err(err).Msg("get targets func failed")
		return
	}

	// iterate through  eligible liquidation targets
	for _, target := range targets {
		// select one reward denom and one repay denom to consider on target address
		intent, ok, err := selectLiquidationDenoms(ctx, target)
		if err != nil {
			logger.Err(err).Str(
				"target-address",
				target.Addr.String(),
			).Str(
				"target-borrowed",
				target.Borrowed.String(),
			).Str(
				"target-collateral",
				target.Collateral.String(),
			).Msg("selection func failed")
			continue
		}
		if !ok {
			continue
		}

		// estimate actual liquidation outcome if chosen denoms were to be liquidated
		estimate, err := estimateLiquidationOutcome(ctx, intent)
		if err != nil {
			logger.Err(err).Str(
				"target-address",
				intent.Addr.String(),
			).Str(
				"intended-repay",
				intent.Repay.String(),
			).Str(
				"intended-reward",
				intent.Reward.String(),
			).Msg("estimate func failed")
			continue
		}

		// decide whether to liquidate based on estimated outcome
		ok, err = decisionFunc(ctx, estimate)
		if err != nil {
			logger.Err(err).Str(
				"target-address",
				estimate.Addr.String(),
			).Str(
				"estimated-repay",
				estimate.Repay.String(),
			).Str(
				"estimated-reward",
				estimate.Reward.String(),
			).Msg("decision func failed")
			continue
		}
		if !ok {
			continue
		}

		// attempt liquidation if it was approved by decisionFunc
		outcome, err := executeLiquidation(ctx, intent)
		if err != nil {
			logger.Err(err).Str(
				"target-addres",
				intent.Addr.String(),
			).Str(
				"intended-repay",
				intent.Repay.String(),
			).Str(
				"intended-reward",
				intent.Reward.String(),
			).Msg("liquidation func failed")
			continue
		}

		logger.Info().Str(
			"target-addres",
			outcome.Addr.String(),
		).Str(
			"repaid",
			outcome.Repay.String(),
		).Str(
			"reward",
			outcome.Reward.String(),
		).Msg("liquidation success")
	}
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
