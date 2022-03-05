package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/umee-network/liquidator/core"
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
	lock.Lock()
	launched = true // prevents further calls to Init
	lock.Unlock()

	configPath, err := cmd.Flags().GetString(flagConfigPath)
	if err != nil {
		return err
	}

	konfig, err := getConfig(configPath)
	if err != nil {
		return err
	}

	logger, err := getLogger(konfig)
	if err != nil {
		return err
	}

	password, err := getPass()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// listen for and trap any OS signal to gracefully shutdown and exit
	trapSignal(cancel, logger)

	g.Go(func() error {
		// returns on context cancelled
		return core.Start(ctx, logger, konfig, password, cancel)
	})

	// Block main process until all spawned goroutines have gracefully exited and
	// signal has been captured in the main process or if an error occurs.
	return g.Wait()
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
