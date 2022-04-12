package liquidator

import (
	"context"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
	"github.com/tendermint/tendermint/libs/sync"
)

var (
	// lock used by sweepLiquidations, reconfigure, and Customize
	lock sync.Mutex

	Cancel     func()
	konfig     *koanf.Koanf
	configFile *file.File
	logger     zerolog.Logger = zerolog.Nop()
	ticker     *time.Ticker

	// password string
)

// Start causes the liquidator to continuously look for liquidation targets, decide on which
// borrowed and collateral denominations to attempt to liquidate, and attempt to execute any liquidations
// whose estimated outcomes it approves of. Returns only when context is canceled.
func Start(
	ctx context.Context,
	log zerolog.Logger,
	configPath string,
	keyringPassword string,
) error {
	ctx, Cancel = context.WithCancel(ctx)

	logger = log
	// password = keyringPassword

	// load config file, then watch to reload on changes
	configFile = file.Provider(configPath)
	loadConfig()
	err := configFile.Watch(reloadConfig)
	if err != nil {
		return err
	}

	// TODO: Create and start clients here. We need:
	//	- umee node query client (e.g. QueryEligibleLiquidationTargets)
	//	- umee node client with keyring (sign and submit MsgLiquidate)

	ticker = time.NewTicker(time.Second)
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
	// prevents config file hot-reload during tick
	lock.Lock()
	defer lock.Unlock()

	// on empty config, skip tick
	if konfig == nil {
		logger.Info().Msg("empty config")
		return
	}

	// get a list of eligible liquidation targets
	targets, err := getLiquidationTargets(ctx, konfig)
	if err != nil {
		logger.Err(err).Msg("get targets func failed")
		return
	}

	// iterate through eligible liquidation targets
	for _, target := range targets {
		// select one reward denom and one repay denom to consider on target address
		intent, ok, err := selectLiquidationDenoms(ctx, konfig, target)
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
		estimate, err := estimateLiquidationOutcome(ctx, konfig, intent)
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
		ok, err = approveLiquidation(ctx, konfig, estimate)
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
		outcome, err := executeLiquidation(ctx, konfig, intent)
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
