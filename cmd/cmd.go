package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	envVariablePass = "LIQUIDATOR_PASSWORD"
	flagConfigPath  = "config-file"
)

var rootCmd = &cobra.Command{
	Use:   "liquidator [" + flagConfigPath + "]",
	Args:  cobra.ExactArgs(1),
	Short: "liquidator runs a basic umee liquidator bot",
	Long:  `liquidator runs a basic umee liquidator bot.`,
	RunE:  liquidatorCmdHandler,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func liquidatorCmdHandler(cmd *cobra.Command, _ []string) error {

	configPath, err := cmd.Flags().GetString(flagConfigPath)
	if err != nil {
		return err
	}

	konfig, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	logger, err := getLogger(konfig)
	if err != nil {
		return err
	}

	password, err := getPassword()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(cancel, logger)

	g.Go(func() error {
		// returns on context cancelled
		return startLiquidator(ctx, konfig, logger, password, cancel)
	})

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process or if an error occurs.
	return g.Wait()
}

// getPassword reads the keyring password from an environment variable
func getPassword() (string, error) {
	pass := os.Getenv(envVariablePass)
	if pass == "" {
		return "", fmt.Errorf("empty keyring password")
	}
	return pass, nil
}

// loadConfig returns a koanf configuration loaded from a specified filepath
func loadConfig(path string) (*koanf.Koanf, error) {
	var k = koanf.New(".")

	// Load toml config from specified file path
	f := file.Provider(path)
	if err := k.Load(f, toml.Parser()); err != nil {
		return nil, err
	}

	return k, nil
}

// getLogger returns a zerolog logger configured from the "log.level" and "log.format"
// fields of a koanf config. Log format should be "json" or "text".
func getLogger(konfig *koanf.Koanf) (*zerolog.Logger, error) {

	logLvl, err := zerolog.ParseLevel(konfig.String("log.level"))
	if err != nil {
		return nil, err
	}

	logFormat := strings.ToLower(konfig.String("log.format"))

	var logWriter io.Writer
	if logFormat == "text" {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	switch logFormat {
	case "json":
		logWriter = os.Stderr
	case "text":
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	default:
		return nil, fmt.Errorf("invalid logging format: %s", logFormat)
	}

	l := zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()
	return &l, nil
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit. Uses the logger
// stored in the liquidator, which may be updated by config file changes
// during runtime.
func trapSignal(cancel context.CancelFunc, logger *zerolog.Logger) {
	sigCh := make(chan os.Signal)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		logger.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		cancel()
	}()
}
