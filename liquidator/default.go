package liquidator

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/knadh/koanf"

	types "github.com/umee-network/liquidator/types"
)

const (
	ConfigKeySelectRepayDenoms  = "liquidator.select.repay_denoms"
	ConfigKeySelectRewardDenoms = "liquidator.select.reward_denoms"
)

func errInvalidConfig(k *koanf.Koanf, key string) error {
	val := k.String(key)
	return fmt.Errorf("invalid %s: %s", key, val)
}

// defaultTargetFunc queries the chain for all eligible liquidation
// targets and their total borrowed and collateral, converting collateral uTokens
// to equivalent base tokens.
var defaultTargetFunc types.TargetFunc = func(
	ctx context.Context, k *koanf.Koanf,
) ([]types.LiquidationTarget, error) {
	// TODO: body
	// - init/reconnect query client if failed last time
	// - use QueryEligibleLiquidationTargets to get a list of addresses
	// - use GetBorrowed and GetCollateral on each address
	// - return the info as []LiquidationTarget
	// - (error on any query failing)
	return nil, nil
}

// validateDefaultTargetConfig is the config file validator associated with defaultTargetFunc
func validateDefaultTargetConfig(k *koanf.Koanf) error {
	return nil
}

// defaultSelectFunc receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral, and returns preferred reward and repay denoms.
// It chooses via simple order or priority using slices of denoms set in the config
// file, Then sets reward amount to zero, which opts out of a user-enforced minimum ratio
// of reward:repay, and trusts the module's oracle.
var defaultSelectFunc types.SelectFunc = func(ctx context.Context, k *koanf.Koanf, target types.LiquidationTarget,
) (types.LiquidationOrder, bool, error) {
	order := types.LiquidationOrder{Addr: target.Addr}

repay:
	for _, r := range k.Strings(ConfigKeySelectRepayDenoms) {
		for _, b := range target.Borrowed {
			if b.Denom == r {
				order.Repay = b
				break repay
			}
		}
	}

reward:
	for _, r := range k.Strings(ConfigKeySelectRewardDenoms) {
		for _, c := range target.Collateral {
			if c.Denom == r {
				order.Reward = c
				order.Reward.Amount = sdk.ZeroInt()
				break reward
			}
		}
	}

	if order.Repay.Denom == "" || order.Reward.Denom == "" {
		return types.LiquidationOrder{}, false, nil
	}

	return order, true, nil
}

// validateDefaultSelectConfig is the config file validator associated with defaultSelectFunc.
func validateDefaultSelectConfig(k *koanf.Koanf) error {
	repays := k.Strings(ConfigKeySelectRepayDenoms)
	if len(repays) == 0 {
		return errInvalidConfig(k, ConfigKeySelectRepayDenoms)
	}
	rewards := k.Strings(ConfigKeySelectRewardDenoms)
	if len(rewards) == 0 {
		return errInvalidConfig(k, ConfigKeySelectRewardDenoms)
	}
	return nil
}

// defaultEstimateFunc simulates the result of a MsgLiquidate in selected
// denominations as closely as possible to how it would be executed by the leverage
// module on the umee chain.
var defaultEstimateFunc types.EstimateFunc = func(ctx context.Context, k *koanf.Koanf, intent types.LiquidationOrder,
) (types.LiquidationOrder, error) {
	// TODO: body
	// - use oracle query client, query required exchange rates
	// - use as-of-yet nonexistent estimate liquidation outcome function from x/leverage/types
	//   (that, or create the function here and avoid moving it to types in the actual module)
	return types.LiquidationOrder{}, nil
}

// validateDefaultEstimateConfig is the config file validator associated with defaultEstimateFunc
func validateDefaultEstimateConfig(k *koanf.Koanf) error {
	return nil
}

// defaultApproveFunc approves all nonzero estimated liquidation outcomes.
var defaultApproveFunc types.ApproveFunc = func(ctx context.Context, k *koanf.Koanf, estimate types.LiquidationOrder,
) (bool, error) {
	if estimate.Addr.Empty() {
		return false, fmt.Errorf("empty address")
	}
	if err := estimate.Repay.Validate(); err != nil {
		return false, err
	}
	if err := estimate.Reward.Validate(); err != nil {
		return false, err
	}
	if estimate.Reward.IsPositive() {
		return true, nil
	}
	return false, nil
}

// validateDefaultApproveConfig is the config file validator associated with defaultApproveFunc
func validateDefaultApproveConfig(k *koanf.Koanf) error {
	return nil
}

// defaultExecuteFunc attempts to execute a chosen liquidation, and reports back
// the actual repaid and reward amounts if successful
var defaultExecuteFunc types.ExecuteFunc = func(ctx context.Context, k *koanf.Koanf, intent types.LiquidationOrder,
) (types.LiquidationOrder, error) {
	// TODO: body
	// - use keyring-enabled client to send a MsgLiquidate with fields from input "intent"
	// - get the MsgLiquidateResponse and return a LiquidationOrder struct with reward and repaid amounts
	return types.LiquidationOrder{}, nil
}

// validateDefaultExecuteConfig is the config file validator associated with defaultExecuteFunc
func validateDefaultExecuteConfig(k *koanf.Koanf) error {
	return nil
}
