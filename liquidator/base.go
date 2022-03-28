package liquidator

import (
	"context"
	"fmt"

	"github.com/knadh/koanf"

	types "github.com/umee-network/liquidator/types"
)

const (
	keySelectRepayDenoms  = "select.repay_denoms"
	keySelectRewardDenoms = "select.reward_denoms"
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
	return nil, nil
}

// validateBaseTargetConfig is the config file validator associated with baseTargetFunc
func validateBaseTargetConfig(k *koanf.Koanf) error {
	return nil
}

// baseSelectFunc receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral, and returns preferred reward and repay denoms.
// chooses via simple order or priority using slices of denoms set in the config
// file.
var baseSelectFunc types.SelectFunc = func(ctx context.Context, k *koanf.Koanf, target types.LiquidationTarget,
) (types.LiquidationOrder, bool, error) {
	order := types.LiquidationOrder{Addr: target.Addr}

repay:
	for _, r := range k.Strings(keySelectRepayDenoms) {
		for _, b := range target.Borrowed {
			if b.Denom == r {
				order.Repay = b
				break repay
			}
		}
	}

reward:
	for _, r := range k.Strings(keySelectRewardDenoms) {
		for _, c := range target.Collateral {
			if c.Denom == r {
				order.Reward = c
				break reward
			}
		}
	}

	if order.Repay.IsZero() || order.Reward.IsZero() {
		return types.LiquidationOrder{}, false, nil
	}

	return order, true, nil
}

// validateBaseSelectConfig is the config file validator associated with baseSelectFunc
func validateBaseSelectConfig(k *koanf.Koanf) error {
	repays := k.Strings(keySelectRepayDenoms)
	if len(repays) == 0 {
		return errInvalidConfig(k, keySelectRepayDenoms)
	}
	rewards := k.Strings(keySelectRewardDenoms)
	if len(rewards) == 0 {
		return errInvalidConfig(k, keySelectRewardDenoms)
	}
	return nil
}

// baseEstimateFunc simulates the result of a MsgLiquidate in selected
// denominations as closely as possible to how it would be executed by the leverage
// module on the umee chain.
var baseEstimateFunc types.EstimateFunc = func(ctx context.Context, k *koanf.Koanf, intent types.LiquidationOrder,
) (types.LiquidationOrder, error) {
	// TODO: body
	return types.LiquidationOrder{}, nil
}

// validateBaseEstimateConfig is the config file validator associated with baseEstimateFunc
func validateBaseEstimateConfig(k *koanf.Koanf) error {
	return nil
}

// baseApproveFunc approves all nonzero estimated liquidation outcomes.
var baseApproveFunc types.ApproveFunc = func(ctx context.Context, k *koanf.Koanf, estimate types.LiquidationOrder,
) (bool, error) {
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
	return types.LiquidationOrder{}, nil
}

// validateBaseExecuteConfig is the config file validator associated with baseExecuteFunc
func validateBaseExecuteConfig(k *koanf.Koanf) error {
	return nil
}
