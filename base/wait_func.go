package base

import (
	"time"

	"github.com/knadh/koanf"
)

// WaitFunc is called between scans for eligible liquidation targets
// on chain, during which all other operations take place. It should
// return nil after waiting the desired amount of time, or an error
// if something stops it from computing said duration. Example
// behaviors include waiting a fixed time based on config file,
// or returning only after a set amount of blocks.
type WaitFunc func(*koanf.Koanf) error

// DefaultWaitFunc waits one second
func DefaultWaitFunc(konfig *koanf.Koanf) error {
	time.Sleep(time.Second)
	return nil
}
