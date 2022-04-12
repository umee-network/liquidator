package liquidator

import (
	"fmt"
	"time"

	"github.com/knadh/koanf"

	types "github.com/umee-network/liquidator/types"
)

const (
	ConfigKeyWait = "liquidator.wait"
)

// Reconfigure validates an incoming config file, updating liquidator's internal config if it is valid,
// setting it to nil otherwise. Used once on startup, and again every time config file is hot-reloaded.
func Reconfigure(k *koanf.Koanf) error {
	// Wait for liquidator to finish current sweep
	lock.Lock()
	defer lock.Unlock()

	// Collect all validate functions that need to be called. Automatically add baseValidateConfig
	// and validate funcs for default liquidation step implementations, so that ValidateConfigFuncs
	// explicitly added only need to cover steps that have been replaced with custom implementations.
	validateFuncs := []types.ValidateFunc{baseValidateConfig}
	validateFuncs = append(validateFuncs, validateConfigFuncs...)

	// Execute all config validation functions.
	for _, v := range validateFuncs {
		if err := v(k); err != nil {
			// Setting konfig to nil on invalid input ensures that the most recent valid  config
			// will not continue to be used as the user attempts modifications during runtime
			konfig = nil
			return err
		}
	}

	// Update config file and reset the ticker controlling the main loop
	konfig = k
	if ticker != nil {
		ticker.Reset(konfig.Duration("liquidator.interval"))
	}
	return nil
}

// baseValidateConfig validates config fields that are not associated with swappable steps in the
// liquidation sweep
func baseValidateConfig(k *koanf.Koanf) error {
	interval := k.Duration(ConfigKeyWait)
	if interval < time.Second {
		return fmt.Errorf("%s must be a nonzero duration longer than 1s", ConfigKeyWait)
	}
	return nil
}
