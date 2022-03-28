package liquidator

import (
	"context"

	"github.com/knadh/koanf"

	types "github.com/umee-network/liquidator/types"
)

const (
	keyDisableTargeting = "targeting.disable"
)

// baseTargetFunc queries the chain for all eligible liquidation
// targets and their total borrowed and collateral, converting collateral uTokens
// to equivalent base tokens.
var baseTargetFunc types.TargetFunc = func(
	ctx context.Context, k *koanf.Koanf,
) ([]types.LiquidationTarget, error) {
	if k.Bool(keyDisableTargeting) {
		return nil, nil
	}

	// TODO: body
	return nil, nil
}

// validateBaseTargetConfig is the config file validator associated with BaseTargetFunc
func validateBaseTargetConfig(k *koanf.Koanf) error {
	_ = k.Bool(keyDisableTargeting)
	return nil
}

// baseSelectFunc receives a LiquidationTarget indicating a single borrower's
// address, borrows, and collateral. From there it should decide what borrow
// denominations the liquidator is interested in repaying, and what collateral
// rewards it wants to receive. For example, a liquidation target with three
// borrowed denominations and two collateral denominations will have six possible
// combinations of (repay denom, reward denom) that could be made into liquidation
// orders. The liquidator may not possess every repayment denom, or be interested
// in every collateral denom. Furthermore, if one liquidation brings the borrower
// back to health, then the remaining ones will no longer be available. The
// defaultSelectFunc function should choose the reward and repay denoms from
// available options. Also returns a boolean, which can be set to false if no
// workable denominations were discovered.
var baseSelectFunc types.SelectFunc = func(ctx context.Context, k *koanf.Koanf, target types.LiquidationTarget,
) (types.LiquidationOrder, bool, error) {
	// TODO: body
	return types.LiquidationOrder{}, false, nil
}

// validateBaseSelectConfig is the config file validator associated with BaseSelectFunc
func validateBaseSelectConfig(k *koanf.Koanf) error {
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

// validateBaseEstimateConfig is the config file validator associated with BaseEstimateFunc
func validateBaseEstimateConfig(k *koanf.Koanf) error {
	return nil
}

// baseApproveFunc decides whether an estimated liquidation outcome is worth it
var baseApproveFunc types.ApproveFunc = func(ctx context.Context, k *koanf.Koanf, estimate types.LiquidationOrder,
) (bool, error) {
	// TODO: body
	return false, nil
}

// validateBaseApproveConfig is the config file validator associated with BaseEstimateFunc
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

// validateBaseExecuteConfig is the config file validator associated with BaseEstimateFunc
func validateBaseExecuteConfig(k *koanf.Koanf) error {
	return nil
}
