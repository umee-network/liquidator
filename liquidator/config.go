package liquidator

import (
	"fmt"
	"time"

	"github.com/knadh/koanf"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
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

// loadConfig returns a koanf configuration loaded from a specified filepath
func loadConfig(path string) (*koanf.Koanf, error) {
	var k = koanf.New(".")

	// Load toml config from specified file path
	f := file.Provider(path)
	if err := k.Load(f, toml.Parser()); err != nil {
		return nil, err
	}

	// Only return non-nil if valid
	if err := validateConfig(k); err != nil {
		return nil, err
	}

	return k, nil
}
