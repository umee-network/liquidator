package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
)

const (
	envVariablePass = "LIQUIDATOR_PASSWORD"
	flagConfigPath  = "config-file"
	logLevelJSON    = "json"
	logLevelText    = "text"
)

// GetPasswordFunc is used to retrieve the user's keyring password on startup
type GetPasswordFunc = func() (string, error)

// GetConfigFunc is used to load config on startup. It receives cmd's first argument (config-file)
type GetConfigFunc = func(string) (*koanf.Koanf, error)

// GetLoggerFunc is used to create a logger on startup. It receives the output of GetConfigFunc
type GetLoggerFunc = func(*koanf.Koanf) (*zerolog.Logger, error)

var (
	lock     sync.Mutex
	launched bool

	getPass   GetPasswordFunc = DefaultGetPassword
	getConfig GetConfigFunc   = DefaultLoadConfig
	getLogger GetLoggerFunc   = DefaultGetLogger
)

// Init can be used to alter the functions cmd uses to get keyring password,
// load config, and create its logger. It can only be used before cmd.Execute
func Init(pf GetPasswordFunc, cf GetConfigFunc, lf GetLoggerFunc) {
	lock.Lock()
	defer lock.Unlock()

	if launched {
		panic("already launched")
	}

	getPass = pf
	getConfig = cf
	getLogger = lf
}

// DefaultGetPassword reads the keyring password from an environment variable
func DefaultGetPassword() (string, error) {
	pass := os.Getenv(envVariablePass)
	if pass == "" {
		return "", fmt.Errorf("empty keyring password")
	}
	return pass, nil
}

// DefaultLoadConfig returns a koanf configuration loaded from a specified filepath
func DefaultLoadConfig(path string) (*koanf.Koanf, error) {
	var k = koanf.New(".")

	// Load toml config from specified file path
	f := file.Provider(path)
	if err := k.Load(f, toml.Parser()); err != nil {
		return nil, err
	}

	return k, nil
}

// DefaultGetLogger returns a zerolog logger configured from the "log.level" and "log.format"
// fields of a koanf config. Log format should be "json" or "text".
func DefaultGetLogger(konfig *koanf.Koanf) (*zerolog.Logger, error) {

	logLvl, err := zerolog.ParseLevel(konfig.String("log.level"))
	if err != nil {
		return nil, err
	}

	logFormat := strings.ToLower(konfig.String("log.format"))

	var logWriter io.Writer
	if strings.ToLower(logFormat) == logLevelText {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	switch logFormat {
	case logLevelJSON:
		logWriter = os.Stderr
	case logLevelText:
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	default:
		return nil, fmt.Errorf("invalid logging format: %s", logFormat)
	}

	logger := zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()
	return &logger, nil
}
