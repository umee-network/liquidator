package core

import (
	"context"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
)

var (
	// created on Start()
	ctx      context.Context
	password string
	logger   *zerolog.Logger
	konfig   *koanf.Koanf
	cancel   context.CancelFunc = func() {}

	// allows functions to be replaced by Init() between cycles after Start()
	lock sync.Mutex

	// default functions used by liquidator bot - can be replaced using Init
	waitFunc       WaitFunc       = DefaultWaitFunc
	targetFunc     TargetFunc     = EmptyTargetFunc
	orderingFunc   OrderingFunc   = EmptyOrderingFunc
	estimationFunc EstimationFunc = EmptyEstimationFunc
	decisionFunc   DecisionFunc   = EmptyDecisionFunc
	executeFunc    ExecuteFunc    = EmptyExecuteFunc
)

// Init sets the Liquidator to use the provided operational functions once it starts. See the
// godocs on each operational function type for descriptions of their required behavior. Can
// be called to replace internal functions, even multiple times or before or after Start(),
// which is used by cmd.Execute().
func Init(
	wf WaitFunc,
	tf TargetFunc,
	of OrderingFunc,
	ef EstimationFunc,
	df DecisionFunc,
	xf ExecuteFunc,
) {
	// waits for main loop to cycle
	lock.Lock()
	defer lock.Unlock()

	waitFunc = wf
	targetFunc = tf
	orderingFunc = of
	estimationFunc = ef
	decisionFunc = df
	executeFunc = xf
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

// Stop cancels the liquidator's context, eventually halting its operation
func Stop() {
	cancel()
}

// Start causes the liquidator to continuously look for liquidation targets, decide on which
// borrowed and collateral denominations to attempt to liquidate, and attempt to execute any
// liquidations whose estimated outcomes are approved by its configured decisionmaking.
func Start(
	ctx context.Context,
	logger *zerolog.Logger,
	konfig *koanf.Koanf,
	keyringPassword string,
	cancelFunc context.CancelFunc,
) error {
	cancel = cancelFunc

	// loop as long as ctx is not cancelled
	for ctx.Err() == nil {
		// blocks any calls to Init during cycle
		lock.Lock()
		defer lock.Unlock()

		// get a list of eligible liquidation targets
		targets, err := targetFunc(konfig)
		if err != nil {
			logger.Err(err)
			continue
		}

		for _, target := range targets {
			if ctx.Err() == nil {
				// for each liquidation target, create an ordered list of liquidations to attempt
				orders, err := orderingFunc(konfig, target)
				if err != nil {
					logger.Err(err)
					continue
				}
				for _, order := range orders {
					if ctx.Err() == nil {
						// for each liquidation order, estimate actual liquidation outcome
						estOutcome, err := estimationFunc(konfig, order)
						if err != nil {
							logger.Err(err)
							continue
						}
						// decide whether to liquidate based on estimated outcomes
						ok, err := decisionFunc(konfig, estOutcome)
						if err != nil {
							logger.Err(err)
							continue
						}
						// attempt liquidation if it was approved by decisionFunc
						if ok {
							outcome, err := executeFunc(konfig, order)
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
				}
			}
		}

		// Wait at the end of each cycle
		if err := waitFunc(konfig); err != nil {
			// If the wait function encounters an error (which it might if the
			// function relies on outside queries like block number), then
			// a short sleep is triggered to prevent extremely fast looping.
			time.Sleep(time.Second)
			logger.Err(err)
			continue
		}
	}
	return ctx.Err()
}
