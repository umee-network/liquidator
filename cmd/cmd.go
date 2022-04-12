package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/umee-network/liquidator/liquidator"
)

const (
	envVariablePass = "KEYRING_PASSPHRASE" // nolint: gosec
	flagConfigPath  = "config"
	flagLogLevel    = "log-level"
	flagLogFormat   = "log-format"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "umeeliqd",
		Short: "umeeliqd runs a basic umee liquidator bot",
		Long: `umeeliquidd runs a basic umee liquidator bot. Reads environment var
KEYRING_PASSPHRASE on start as well as requiring a toml config file.`,
		RunE: liquidatorCmdHandler,
	}

	cmd.PersistentFlags().String(flagConfigPath, "path/to/umeeliqd.toml", "config file path")
	cmd.PersistentFlags().String(flagLogLevel, "debug", "log level")
	cmd.PersistentFlags().String(flagLogFormat, "text", "log format (text|json)")

	cmd.AddCommand(getVersionCmd())

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func liquidatorCmdHandler(cmd *cobra.Command, _ []string) error {
	configPath, err := cmd.Flags().GetString(flagConfigPath)
	if err != nil {
		return err
	}
	logLevel, err := cmd.Flags().GetString(flagLogLevel)
	if err != nil {
		return err
	}
	logFormat, err := cmd.Flags().GetString(flagLogFormat)
	if err != nil {
		return err
	}
	logger, err := getLogger(logLevel, logFormat)
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
	trapSignal(cancel, &logger)

	g.Go(func() error {
		// returns on context canceled
		return liquidator.Start(ctx, logger, configPath, password)
	})

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process or if an error occurs.
	return g.Wait()
}

// getPassword reads the keyring password from an environment variable.
func getPassword() (string, error) {
	pass := os.Getenv(envVariablePass)
	if pass == "" {
		return "", fmt.Errorf("empty keyring password")
	}
	return pass, nil
}

// getLogger returns a zerolog logger with the given level and format. Log format
// should be "json" or "text".
func getLogger(logLevel, logFormat string) (zerolog.Logger, error) {
	logLvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return zerolog.Nop(), err
	}

	logFmt := strings.ToLower(logFormat)

	var logWriter io.Writer
	switch logFmt {
	case "json":
		logWriter = os.Stderr
	case "text":
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	default:
		return zerolog.Nop(), fmt.Errorf("invalid logging format: %s", logFmt)
	}

	return zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger(), nil
}

// trapSignal will listen for any OS signal and invoke Done on the main
// WaitGroup allowing the main process to gracefully exit.
func trapSignal(cancel context.CancelFunc, logger *zerolog.Logger) {
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		logger.Info().Str("signal", sig.String()).Msg("caught signal; shutting down...")
		cancel()
	}()
}
