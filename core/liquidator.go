package base

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
)

// NewLiquidator creates a Liquidator from the provided context (for cancelation), logger, konfig,
// and keyring password as well as operational functions to use once it is started. See the godocs
// on operational functions types for descriptions of their required behavior.
func NewLiquidator(
	ctx context.Context,
	logger *zerolog.Logger,
	konfig *koanf.Koanf,
	keyringPassword string,
	wf WaitFunc,
	tf TargetFunc,
	of OrderingFunc,
	ef EstimationFunc,
	df DecisionFunc,
	xf ExecuteFunc,
) Liquidator {
	if logger == nil || konfig == nil {
		panic("logger or konfig was nil")
	}
	return Liquidator{
		ctx:            ctx,
		logger:         logger,
		konfig:         konfig,
		password:       keyringPassword,
		waitFunc:       wf,
		targetFunc:     tf,
		orderingFunc:   of,
		estimationFunc: ef,
		decisionFunc:   df,
		executeFunc:    xf,
	}
}

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

// Liquidator is a configurable structure with a cancelable context, which can query a
// umee node to look for eligible liquidation targets, and use its keyring to sign and
// execute liquidation transactions.
type Liquidator struct {
	ctx context.Context

	password string
	logger   *zerolog.Logger
	konfig   *koanf.Koanf

	waitFunc       func(*koanf.Koanf) error
	targetFunc     func(*koanf.Koanf) ([]LiquidationTarget, error)
	orderingFunc   func(*koanf.Koanf, LiquidationTarget) ([]LiquidationOrder, error)
	estimationFunc func(*koanf.Koanf, LiquidationOrder) (LiquidationOrder, error)
	decisionFunc   func(*koanf.Koanf, LiquidationOrder) (bool, error)
	executeFunc    func(*koanf.Koanf, LiquidationOrder) (LiquidationOrder, error)
}

// Start causes the liquidator to continuously look for liquidation targets, decide on which
// borrowed and collateral denominations to attempt to liquidate, and attempt to execute any
// liquidations whose estimated outcomes are approved by its configured decisionmaking.
func (liq *Liquidator) Start() error {
	// loop as long as ctx is not cancelled
	for liq.ctx.Err() == nil {
		// get a list of eligible liquidation targets
		targets, err := liq.targetFunc(liq.konfig)
		if err != nil {
			liq.logger.Err(err)
			continue
		}

		for _, target := range targets {
			if liq.ctx.Err() == nil {
				// for each liquidation target, create an ordered list of liquidations to attempt
				orders, err := liq.orderingFunc(liq.konfig, target)
				if err != nil {
					liq.logger.Err(err)
					continue
				}
				for _, order := range orders {
					if liq.ctx.Err() == nil {
						// for each liquidation order, estimate actual liquidation outcome
						estOutcome, err := liq.estimationFunc(liq.konfig, order)
						if err != nil {
							liq.logger.Err(err)
							continue
						}
						// decide whether to liquidate based on estimated outcomes
						ok, err := liq.decisionFunc(liq.konfig, estOutcome)
						if err != nil {
							liq.logger.Err(err)
							continue
						}
						// attempt liquidation if it was approved by decisionFunc
						if ok {
							outcome, err := liq.executeFunc(liq.konfig, order)
							if err != nil {
								liq.logger.Err(err)
								continue
							}
							liq.logger.Debug().Msgf(
								"LIQUIDATION: target: %s repaid %s reward%s",
								outcome.Addr.String(),
								outcome.Repay.String(),
								outcome.Reward.String(),
							)
						}
					}
				}
			}
		}

		// Wait at the end of each cycle
		if err := liq.waitFunc(liq.konfig); err != nil {
			// If the wait function encounters an error (which it might if the
			// function relies on outside queries like block number), then
			// a short sleep is triggered to prevent extremely fast looping.
			time.Sleep(time.Second)
			liq.logger.Err(err)
			continue
		}
	}
	return liq.ctx.Err()
}
