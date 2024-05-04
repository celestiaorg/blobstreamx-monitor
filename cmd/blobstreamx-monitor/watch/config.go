package watch

import (
	"errors"
	"fmt"

	"github.com/celestiaorg/blobstreamx-monitor/telemetry"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

const (
	FlagEVMRPC             = "evm.rpc"
	FlagEVMContractAddress = "evm.contract-address"

	FlagMetricsEndpoint = "metrics.endpoint"
	FlagMetricsTLS      = "metrics.tls"
	FlagMetricsLabel    = "metrics.label"

	FlagLogLevel  = "log.level"
	FlagLogFormat = "log.format"
)

func addStartFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().String(FlagEVMRPC, "http://localhost:8545", "Specify the ethereum rpc address")
	cmd.Flags().String(FlagEVMContractAddress, "", "Specify the contract at which the BlobstreamX is deployed")
	cmd.Flags().String(FlagMetricsLabel, "blobstream x", "Custom label for the metric")
	cmd.Flags().String(
		FlagMetricsEndpoint,
		"localhost:4318",
		"Sets HTTP endpoint for OTLP metrics to be exported to",
	)
	cmd.Flags().Bool(
		FlagMetricsTLS,
		false,
		"Enable TLS connection to OTLP metric backend",
	)
	cmd.Flags().String(
		FlagLogLevel,
		"info",
		"The logging level (trace|debug|info|warn|error|fatal|panic)",
	)
	cmd.Flags().String(
		FlagLogFormat,
		"plain",
		"The logging format (json|plain)",
	)
	return cmd
}

type StartConfig struct {
	EVMRPC          string
	ContractAddress string
	MetricsConfig   telemetry.Config
	LogLevel        string
	LogFormat       string
}

func (cfg StartConfig) ValidateBasics() error {
	if err := ValidateEVMAddress(cfg.ContractAddress); err != nil {
		return fmt.Errorf("%s: flag --%s", err.Error(), FlagEVMContractAddress)
	}
	return nil
}

func ValidateEVMAddress(addr string) error {
	if addr == "" {
		return fmt.Errorf("the EVM address cannot be empty")
	}
	if !ethcmn.IsHexAddress(addr) {
		return errors.New("valid EVM address is required")
	}
	return nil
}

func parseStartFlags(cmd *cobra.Command) (StartConfig, error) {
	contractAddress, err := cmd.Flags().GetString(FlagEVMContractAddress)
	if err != nil {
		return StartConfig{}, err
	}

	evmRPC, err := cmd.Flags().GetString(FlagEVMRPC)
	if err != nil {
		return StartConfig{}, err
	}

	endpoint, err := cmd.Flags().GetString(FlagMetricsEndpoint)
	if err != nil {
		return StartConfig{}, err
	}

	label, err := cmd.Flags().GetString(FlagMetricsLabel)
	if err != nil {
		return StartConfig{}, err
	}

	tls, err := cmd.Flags().GetBool(FlagMetricsTLS)
	if err != nil {
		return StartConfig{}, err
	}

	logLevel, err := cmd.Flags().GetString(FlagLogLevel)
	if err != nil {
		return StartConfig{}, err
	}

	logFormat, err := cmd.Flags().GetString(FlagLogFormat)
	if err != nil {
		return StartConfig{}, err
	}

	return StartConfig{
		EVMRPC:          evmRPC,
		ContractAddress: contractAddress,
		MetricsConfig: telemetry.Config{
			Endpoint: endpoint,
			TLS:      tls,
			Label:    label,
		},
		LogLevel:  logLevel,
		LogFormat: logFormat,
	}, nil
}
