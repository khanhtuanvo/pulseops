package observability

import (
	"context"
	"os"

	"github.com/tuankhanhvo/pulseops/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
)

func InitTracing(ctx context.Context, cfg config.Config, logger *zap.Logger) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if cfg.OTLPEndpoint == "" {
		logger.Info("opentelemetry tracing disabled", zap.String("reason", "OTEL_EXPORTER_OTLP_ENDPOINT not set"))
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	logger.Info("opentelemetry tracing enabled",
		zap.String("serviceName", cfg.ServiceName),
		zap.String("endpoint", cfg.OTLPEndpoint),
		zap.Bool("headersConfigured", os.Getenv("OTEL_EXPORTER_OTLP_HEADERS") != ""),
	)

	return provider.Shutdown, nil
}
