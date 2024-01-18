package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.11.0"
)

const (
	defaultMetricsCollectInterval = 10 * time.Second
	globalMetricsNamespace        = "blobstreamx-watcher"
)

// Config defines the configuration options for blobstreamx-monitor telemetry.
type Config struct {
	Endpoint string
	TLS      bool
}

var meter = otel.Meter(globalMetricsNamespace)

type Meters struct {
	ProcessedNonces metric.Int64Counter
}

func InitMeters() (*Meters, error) {
	processedNonces, err := meter.Int64Counter("blobstreamx_monitor_submitted_nonces_counter",
		metric.WithDescription("the count of the nonces that have been successfully submitted to blobstreamx contract"))
	if err != nil {
		return nil, err
	}

	return &Meters{
		ProcessedNonces: processedNonces,
	}, nil
}

func Start(
	ctx context.Context,
	logger tmlog.Logger,
	serviceName string,
	instanceID string,
	opts []otlpmetrichttp.Option,
) (*prometheus.Registry, func() error, error) {
	exp, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	provider := sdk.NewMeterProvider(
		sdk.WithReader(
			sdk.NewPeriodicReader(exp,
				sdk.WithTimeout(defaultMetricsCollectInterval),
				sdk.WithInterval(defaultMetricsCollectInterval))),
		sdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNamespaceKey.String(globalMetricsNamespace),
				semconv.ServiceNameKey.String(serviceName),
				// ServiceInstanceIDKey will be exported with key: "instance"
				semconv.ServiceInstanceIDKey.String(instanceID),
			),
		),
	)

	otel.SetMeterProvider(provider)
	logger.Info("global meter setup", "namespace", globalMetricsNamespace, "service_name_key", serviceName, "service_instance_id_key", instanceID)

	err = runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(defaultMetricsCollectInterval),
		runtime.WithMeterProvider(provider))
	if err != nil {
		return nil, nil, fmt.Errorf("start runtime metrics: %w", err)
	}

	prometheusRegistry := prometheus.NewRegistry()

	return prometheusRegistry, func() error {
		return provider.Shutdown(ctx)
	}, err
}
