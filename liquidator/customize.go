package liquidator

import (
	types "github.com/umee-network/liquidator/types"
)

var (
	// Implementations can be replaced with mockups for testing

	getLiquidationTargets      types.TargetFunc   = baseTargetFunc
	selectLiquidationDenoms    types.SelectFunc   = baseSelectFunc
	estimateLiquidationOutcome types.EstimateFunc = baseEstimateFunc
	approveLiquidation         types.ApproveFunc  = baseApproveFunc
	executeLiquidation         types.ExecuteFunc  = baseExecuteFunc

	// validateConfigFuncs will be called in order upon reloading config file. Custom
	// liquidators which require new parameters to be in the config file should add
	// custom validate funcs.
	validateConfigFuncs []types.ValidateFunc = []types.ValidateFunc{baseValidateConfig}
)

func Customize(
	tf *types.TargetFunc,
	sf *types.SelectFunc,
	ef *types.EstimateFunc,
	af *types.ApproveFunc,
	xf *types.ExecuteFunc,
	vfs []types.ValidateFunc,
) {
	// add base config validate func
	vfs = append(vfs, baseValidateConfig)

	// nil inputs default to base implementations
	if tf == nil {
		tf = &baseTargetFunc
		vfs = append(vfs, validateBaseTargetConfig)
	}
	if sf == nil {
		sf = &baseSelectFunc
		vfs = append(vfs, validateBaseSelectConfig)
	}
	if ef == nil {
		ef = &baseEstimateFunc
		vfs = append(vfs, validateBaseEstimateConfig)
	}
	if af == nil {
		af = &baseApproveFunc
		vfs = append(vfs, validateBaseApproveConfig)
	}
	if xf == nil {
		xf = &baseExecuteFunc
		vfs = append(vfs, validateBaseExecuteConfig)
	}

	// wait until any current ticks or hot config reloads are done
	lock.Lock()
	defer lock.Unlock()

	// replace liquidation steps with chosen implementations
	getLiquidationTargets = *tf
	selectLiquidationDenoms = *sf
	estimateLiquidationOutcome = *ef
	approveLiquidation = *af
	executeLiquidation = *xf

	// clear any previous validate config funcs and add custom ones
	validateConfigFuncs = vfs
}
