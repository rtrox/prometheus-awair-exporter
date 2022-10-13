package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prometheus-awair-exporter/internal/exporter"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func newHealthCheckHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "OK")
	})
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	err := godotenv.Load(".env")
	if err != nil {
		// Typical use will be via direct env in kubernetes,
		// don't fail here.
		log.Warn().Err(err).Msg("No .env file loaded")
	}

	hostname := os.Getenv("AWAIR_HOSTNAME")
	if hostname == "" {
		log.Fatal().
			Msg("AWAIR_HOSTNAME must be set to the hostname of the awair device")
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
	ex, err := exporter.NewAwairExporter(hostname)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to connect to Awair device.")
	}

	prometheus.MustRegister(ex)
	infoMetricOpts.ConstLabels = prometheus.Labels{
		"app_name": app_name,
		"version":  version,
		"hostname": hostname,
	}
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		infoMetricOpts,
		func() float64 { return 1 },
	))

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/healthz", newHealthCheckHandler())
	srv.Addr = ":8080"
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}
	<-idleConnsClosed
}
