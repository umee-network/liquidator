package liquidator

import (
	"fmt"
	"time"

	"github.com/knadh/koanf"
)

const (
	keyWait = "liquidator.wait"

	// TODO: Many other liquidator configuration values, like allowed repayment denoms
)

func validateConfig(konfig *koanf.Koanf) error {
	interval := konfig.Duration(keyWait)
	if interval < time.Second {
		return fmt.Errorf("%s must be a nonzero duration longer than 1s", keyWait)
	}
	return nil
}
