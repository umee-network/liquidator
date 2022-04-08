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

// baseTargetFunc queries the chain for all eligible liquidation
// targets and their total borrowed and collateral, converting collateral uTokens
// to equivalent base tokens.
var baseTargetFunc types.TargetFunc = func(
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

// validateBaseTargetConfig is the config file validator associated with baseTargetFunc
func validateBaseTargetConfig(k *koanf.Koanf) error {
	return nil
}

// baseSelectFunc receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral, and returns preferred reward and repay denoms.
// chooses via simple order or priority using slices of denoms set in the config
// file. sets reward amount to zero, which opts out of a user-enforced minimum ratio
// of reward:repay and trusts the module's oracle.
var baseSelectFunc types.SelectFunc = func(ctx context.Context, k *koanf.Koanf, target types.LiquidationTarget,
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

// validateBaseSelectConfig is the config file validator associated with baseSelectFunc
func validateBaseSelectConfig(k *koanf.Koanf) error {
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

// baseEstimateFunc simulates the result of a MsgLiquidate in selected
// denominations as closely as possible to how it would be executed by the leverage
// module on the umee chain.
var baseEstimateFunc types.EstimateFunc = func(ctx context.Context, k *koanf.Koanf, intent types.LiquidationOrder,
) (types.LiquidationOrder, error) {
	// TODO: body
	// - use oracle query client, query required exchange rates
	// - use as-of-yet nonexistent estimate liquidation outcome function from x/leverage/types
	//   (that, or create the function here and avoid moving it to types in the actual module)
	return types.LiquidationOrder{}, nil
}

// validateBaseEstimateConfig is the config file validator associated with baseEstimateFunc
func validateBaseEstimateConfig(k *koanf.Koanf) error {
	return nil
}

// baseApproveFunc approves all nonzero estimated liquidation outcomes.
var baseApproveFunc types.ApproveFunc = func(ctx context.Context, k *koanf.Koanf, estimate types.LiquidationOrder,
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

// validateBaseApproveConfig is the config file validator associated with baseApproveFunc
func validateBaseApproveConfig(k *koanf.Koanf) error {
	return nil
}

// baseExecuteFunc attempts to execute a chosen liquidation, and reports back
// the actual repaid and reward amounts if successful
var baseExecuteFunc types.ExecuteFunc = func(ctx context.Context, k *koanf.Koanf, intent types.LiquidationOrder,
) (types.LiquidationOrder, error) {
	// TODO: body
	// - use keyring-enabled client to send a MsgLiquidate with fields from input "intent"
	// - get the MsgLiquidateResponse and return a LiquidationOrder struct with reward and repaid amounts
	return types.LiquidationOrder{}, nil
}

// validateBaseExecuteConfig is the config file validator associated with baseExecuteFunc
func validateBaseExecuteConfig(k *koanf.Koanf) error {
	return nil
}
