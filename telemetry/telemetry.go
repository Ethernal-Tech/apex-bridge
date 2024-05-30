package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/armon/go-metrics"
	prometheusMetrics "github.com/armon/go-metrics/prometheus"
	"github.com/hashicorp/go-hclog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

type TelemetryConfig struct {
	PrometheusAddr string `json:"prometheusAddr"` // empty means disabled otherwise something like 0.0.0.0:5001
	DataDogAddr    string `json:"dataDogAddr"`    // empty means disabled otherwise something like localhost:8126
}

// Telemetry holds the config details for metric services
type Telemetry struct {
	prometheusServer *http.Server
	config           TelemetryConfig
	logger           hclog.Logger
}

func NewTelemetry(config TelemetryConfig, logger hclog.Logger) (*Telemetry, error) {
	if err := setupDataDog(config.DataDogAddr != ""); err != nil {
		return nil, err
	}

	return &Telemetry{
		config:           config,
		prometheusServer: setupPrometheus(config.PrometheusAddr),
		logger:           logger,
	}, nil
}

func (t *Telemetry) Start() error {
	go t.startPrometheus()

	return t.startDataDogProfiler()
}

func (t *Telemetry) Close(ctx context.Context) error {
	if t.prometheusServer != nil {
		t.logger.Info("Prometheus server stopping", "addr", t.prometheusServer.Addr)

		if err := t.prometheusServer.Shutdown(ctx); err != nil {
			return err
		}
	}

	if t.config.DataDogAddr != "" {
		profiler.Stop()
		tracer.Stop()
	}

	return nil
}

func (t *Telemetry) startPrometheus() {
	if t.prometheusServer == nil {
		return
	}

	t.logger.Info("Prometheus server started", "addr", t.config.PrometheusAddr)

	if err := t.prometheusServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			t.logger.Error("Prometheus server ListenAndServe error", "err", err)
		}
	}
}

func (t *Telemetry) startDataDogProfiler() error {
	if t.config.DataDogAddr == "" {
		return nil
	}

	err := profiler.Start(
		// enable all profiles
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
			profiler.BlockProfile,
			profiler.MutexProfile,
			profiler.GoroutineProfile,
			profiler.MetricsProfile,
		),
		profiler.WithAgentAddr(t.config.DataDogAddr),
	)
	if err != nil {
		return fmt.Errorf("could not start datadog profiler: %w", err)
	}

	tracer.Start() // start the tracer

	t.logger.Info("DataDog profiler started", "addr", t.config.DataDogAddr)

	return nil
}

func setupDataDog(enabled bool) error {
	if !enabled {
		return nil
	}

	inm := metrics.NewInmemSink(10*time.Second, time.Minute)
	metrics.DefaultInmemSignal(inm)

	promSink, err := prometheusMetrics.NewPrometheusSinkFrom(prometheusMetrics.PrometheusOpts{
		Name:       "apex_bridge_prometheus_sink",
		Expiration: 0,
	})
	if err != nil {
		return err
	}

	metricsConf := metrics.DefaultConfig("apex_bridge")
	metricsConf.EnableHostname = false

	_, err = metrics.NewGlobal(metricsConf, metrics.FanoutSink{
		inm, promSink,
	})

	return err
}

func setupPrometheus(prometheusAddr string) *http.Server {
	if prometheusAddr == "" {
		return nil
	}

	return &http.Server{
		Addr: prometheusAddr,
		Handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{},
			),
		),
		ReadHeaderTimeout: 60 * time.Second,
	}
}
