package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/celestiaorg/blobstreamx-monitor/cmd/blobstreamx-monitor/version"
	"github.com/celestiaorg/blobstreamx-monitor/telemetry"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	blobstreamxwrapper "github.com/succinctlabs/blobstreamx/bindings"
	tmconfig "github.com/tendermint/tendermint/config"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
)

// Command the watcher start command.
func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "start <flags>",
		Short: "Starts the BlobstreamX monitor",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := parseStartFlags(cmd)
			if err != nil {
				return err
			}
			if err := config.ValidateBasics(); err != nil {
				return err
			}

			logger, err := GetLogger(config.LogLevel, config.LogFormat)
			if err != nil {
				return err
			}

			buildInfo := version.GetBuildInfo()
			logger.Info("initializing watcher", "version", buildInfo.SemanticVersion, "build_date", buildInfo.BuildTime)

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			stopFuncs := make([]func() error, 0)
			defer func() {
				for _, f := range stopFuncs {
					err := f()
					if err != nil {
						logger.Error(err.Error())
					}
				}
			}()

			// connecting to a BlobstreamX contract
			ethClient, err := ethclient.Dial(config.EVMRPC)
			if err != nil {
				return err
			}
			defer ethClient.Close()
			blobstreamWrapper, err := blobstreamxwrapper.NewBlobstreamXFilterer(ethcmn.HexToAddress(config.ContractAddress), ethClient)
			if err != nil {
				return err
			}

			meters, err := telemetry.InitMeters()
			if err != nil {
				return err
			}

			opts := []otlpmetrichttp.Option{
				otlpmetrichttp.WithEndpoint(config.MetricsConfig.Endpoint),
				otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
			}
			if !config.MetricsConfig.TLS {
				opts = append(opts, otlpmetrichttp.WithInsecure())
			}
			var shutdown func() error
			_, shutdown, err = telemetry.Start(ctx, logger, "blobstreamx-monitor", config.ContractAddress, opts)
			if shutdown != nil {
				stopFuncs = append(stopFuncs, shutdown)
			}
			if err != nil {
				return err
			}

			// Listen for and trap any OS signal to graceful shutdown and exit
			go TrapSignal(logger, cancel)

			eventsChan := make(chan *blobstreamxwrapper.BlobstreamXDataCommitmentStored, 100)
			subscription, err := blobstreamWrapper.WatchDataCommitmentStored(&bind.WatchOpts{}, eventsChan, nil, nil, nil)
			if err != nil {
				return err
			}
			defer subscription.Unsubscribe()

			logger.Info("starting watcher", "rpc", config.EVMRPC, "address", config.ContractAddress)
			for {
				logger.Debug("waiting for new transactions...")
				select {
				case <-ctx.Done():
					return ctx.Err()
				case err := <-subscription.Err():
					logger.Error("subscription failed", "err", err)
					// TODO(@rach-id): recover from failure
					return err
				case event := <-eventsChan:
					logger.Info(
						"received new data root tuple root event",
						"nonce",
						event.ProofNonce.Uint64(),
						"data_commitment",
						ethcmn.Bytes2Hex(event.DataCommitment[:]),
						"start_block",
						event.StartBlock,
						"end_block",
						event.EndBlock,
					)
					meters.ProcessedNonces.Add(ctx, 1)
					logger.Debug("incrementing metric 'blobstreamx_monitor_submitted_nonces_counter'")
				}
			}
		},
	}
	return addStartFlags(command)
}

// GetLogger creates a new logger and returns
func GetLogger(level string, format string) (tmlog.Logger, error) {
	logLvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level (%s): %w", level, err)
	}
	var logWriter io.Writer
	if strings.ToLower(format) == tmconfig.LogFormatPlain {
		logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
	} else {
		logWriter = os.Stderr
	}

	return server.ZeroLogWrapper{Logger: zerolog.New(logWriter).Level(logLvl).With().Timestamp().Logger()}, nil
}

// TrapSignal will listen for any OS signal and cancel the context to gracefully exit.
func TrapSignal(logger tmlog.Logger, cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGTERM)
	signal.Notify(sigCh, syscall.SIGINT)

	sig := <-sigCh
	logger.Info("caught signal; shutting down...", "signal", sig.String())
	cancel()
}
