package liquidator

import (
	types "github.com/umee-network/liquidator/types"
)

var (
	// Implementations can be replaced with custom functions or mockups for testing

	getLiquidationTargets      types.TargetFunc   = defaultTargetFunc
	selectLiquidationDenoms    types.SelectFunc   = defaultSelectFunc
	estimateLiquidationOutcome types.EstimateFunc = defaultEstimateFunc
	approveLiquidation         types.ApproveFunc  = defaultApproveFunc
	executeLiquidation         types.ExecuteFunc  = defaultExecuteFunc

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

	// nil inputs specify default implementations
	if tf == nil {
		tf = &defaultTargetFunc
		vfs = append(vfs, validateDefaultTargetConfig)
	}
	if sf == nil {
		sf = &defaultSelectFunc
		vfs = append(vfs, validateDefaultSelectConfig)
	}
	if ef == nil {
		ef = &defaultEstimateFunc
		vfs = append(vfs, validateDefaultEstimateConfig)
	}
	if af == nil {
		af = &defaultApproveFunc
		vfs = append(vfs, validateDefaultApproveConfig)
	}
	if xf == nil {
		xf = &defaultExecuteFunc
		vfs = append(vfs, validateDefaultExecuteConfig)
	}

	// after Customize releases its lock, Reconfigure will run the new validateConfigFuncs
	defer func() {
		if err := Reconfigure(konfig); err != nil {
			logger.Err(err).Msg("config validate on Customize")
		}
	}()

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
