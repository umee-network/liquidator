package base

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/knadh/koanf"
)

// EstimationFunc should attempt to simulate the result of a MsgLiquidate contructred
// around the input LiquidationOrder as closely as possible to how it would be executed
// by the leverage module on the umee chain if liquidator balances in the repay denom
// were unlimited. The outcome of this simulated liquidation transaction depends on
// the chain's oracle module as well. The liquidation outcome's reward field should be
// represented by an amount of base tokens equivalent to the actual uToken rewards.
type EstimationFunc func(*koanf.Koanf, LiquidationOrder) (LiquidationOrder, error)

// EmptyEstimationFunc is an EstimationFunc that returns no-op liquidation outcome estimations
func EmptyEstimationFunc(_ *koanf.Koanf, order LiquidationOrder) (LiquidationOrder, error) {
	return LiquidationOrder{
		Addr:   order.Addr,
		Repay:  sdk.NewInt64Coin(order.Repay.Denom, 0),
		Reward: sdk.NewInt64Coin(order.Reward.Denom, 0),
	}, nil
}
