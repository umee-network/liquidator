package liquidator

import (
	"context"
	"time"

	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
)

var (
	// Stored on start
	konfig *koanf.Koanf
	logger *zerolog.Logger
	// password string

	// Implementations can be replaced with mockups for testing

	GetLiquidationTargets      TargetFunc   = defaultGetLiquidationTargets
	SelectLiquidationDenoms    SelectFunc   = defaultSelectLiquidationDenoms
	EstimateLiquidationOutcome EstimateFunc = defaultEstimateLiquidationOutcome
	ApproveLiquidation         ApproveFunc  = defaultApproveLiquidation
	ExecuteLiquidation         ExecuteFunc  = defaultExecuteLiquidation
)

// startLiquidator causes the liquidator to continuously look for liquidation targets, decide on which
// borrowed and collateral denominations to attempt to liquidate, and attempt to execute any liquidations
// whose estimated outcomes it approves of. Returns only when context is canceled.
func StartLiquidator(
	ctx context.Context,
	log *zerolog.Logger,
	config *koanf.Koanf,
	keyringPassword string,
) error {
	logger = log
	// password = keyringPassword
	konfig = config

	if err := validateConfig(config); err != nil {
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
			sweepLiquidations(ctx)
		}
	}
}

func sweepLiquidations(
	ctx context.Context,
) {
	// get a list of eligible liquidation targets
	targets, err := GetLiquidationTargets(ctx)
	if err != nil {
		logger.Err(err).Msg("get targets func failed")
		return
	}

	// iterate through  eligible liquidation targets
	for _, target := range targets {
		// select one reward denom and one repay denom to consider on target address
		intent, ok, err := SelectLiquidationDenoms(ctx, target)
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
		estimate, err := EstimateLiquidationOutcome(ctx, intent)
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
		ok, err = ApproveLiquidation(ctx, estimate)
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
		outcome, err := ExecuteLiquidation(ctx, intent)
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
			).Msg("liquidation func failed")
			continue
		}

		logger.Info().Str(
			"target-address",
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
