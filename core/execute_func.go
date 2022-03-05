package core

import (
	"fmt"

	"github.com/knadh/koanf"
)

// ExecuteFunc should attempt to create, sign, and execute a liquidation transaction using the
// liquidator's internal keyring and and a umee node's grpc endpoint. The input LiquidationOrder
// represents the liquidator's intent, which will be part of an umee MsgLiquidate, and the output
// LiquidationOrder (or error) represent the actual amounts repaid and rewarded, from an umee
// MsgLiquidateResponse. ExecuteFunc behavior can depend on the fields of the input configuration.
type ExecuteFunc func(*koanf.Koanf, LiquidationOrder) (LiquidationOrder, error)

// EmptyExecuteFunc is an ExecuteFunc that always discards liquidation orders and returns an error
func EmptyExecuteFunc(_ *koanf.Koanf, _ LiquidationOrder) (LiquidationOrder, error) {
	return LiquidationOrder{}, fmt.Errorf("not executing liquidations")
}
