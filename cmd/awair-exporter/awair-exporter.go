package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prometheus-awair-exporter/internal/exporter"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// TODO: Remove once automated
	// go build -ldflags="-X \"main.version=${VERSION}\"" ./cmd/awair-exporter/awair-exporter.go
	app_name = "awair-exporter"
	version  = "x.x.x"
)

var (
	infoMetricOpts = prometheus.GaugeOpts{
		Namespace: "exporter",
		Name:      "info",
		Help:      "Info about this awair-exporter",
	}
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	hostname := flag.String("hostname", "", "hostname of Awair device to poll")
	flag.Parse()

	if *hostname == "" {
		log.Fatal().Msg("hostname flag must be set.")
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	var srv http.Server

	idleConnsClosed := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)

		signal.Notify(sigchan, os.Interrupt)
		signal.Notify(sigchan, syscall.SIGTERM)
		sig := <-sigchan
		log.Info().
			Str("signal", sig.String()).
			Msg("Stopping in response to signal")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Failed to gracefully close http server")
		}
		close(idleConnsClosed)
	}()

	log.Info().
		Str("app_name", app_name).
		Str("version", version).
		Msg("Exporter Started.")

	// TODO: retries
	ex, err := exporter.NewAwairExporter(*hostname)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to connect to Awair device.")
	}

	prometheus.MustRegister(ex)
	infoMetricOpts.ConstLabels = prometheus.Labels{
		"app_name": app_name,
		"version":  version,
		"hostname": *hostname,
	}
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		infoMetricOpts,
		func() float64 { return 1 },
	))

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	srv.Addr = ":8080"
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}
	<-idleConnsClosed
}
