package main

import (
	"github.com/umee-network/liquidator/cmd"
	"github.com/umee-network/liquidator/core"
)

func main() {

	// If a program needs to extend or replace the core functionality of
	// the example liquidator bot, it can do so by importing the cmd and
	// core packages, then passing custom functions to cmd.Init and/or
	// core.Init before calling cmd.Execute.

	// Forking the repository is not required unless modifying the
	// contents of core/liquidator.go or cmd/cmd.go.

	// Both Init calls can be omitted if using defaults.

	// cmd.Init must be called before cmd.Execute
	cmd.Init(
		cmd.DefaultGetPassword,
		cmd.DefaultLoadConfig,
		cmd.DefaultGetLogger,
	)

	// core.Init supports hot-swapping after cmd.Execute as well
	core.Init(
		core.DefaultWaitFunc,
		core.EmptyTargetFunc,
		core.EmptyOrderingFunc,
		core.EmptyEstimationFunc,
		core.EmptyDecisionFunc,
		core.EmptyExecuteFunc,
	)

	cmd.Execute()
}
