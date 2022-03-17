package liquidator

import (
	"fmt"
	"time"

	"github.com/knadh/koanf"
)

const (
	key_wait = "liquidator.wait"

	// TODO: Many other liquidator configuration values, like allowed repayment denoms
)

func validateConfig(konfig *koanf.Koanf) error {
	interval := konfig.Duration(key_wait)
	if interval < time.Second {
		return fmt.Errorf("%s must be a nonzero duration longer than 1s", key_wait)
	}
	return nil
}
