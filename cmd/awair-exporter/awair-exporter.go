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

	"prometheus-awair-exporter/internal/app_info"
	"prometheus-awair-exporter/internal/exporter"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	app_name = "awair-exporter"
	version  = "x.x.x"
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

func newProbeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Missing 'target' query parameter", http.StatusBadRequest)
			return
		}
		ex, err := exporter.NewAwairExporter(target)
		if err != nil {
			http.Error(w, "Failed to connect to target: "+err.Error(), http.StatusBadGateway)
			return
		}
		reg := prometheus.NewPedanticRegistry()
		reg.MustRegister(ex)
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}

func newMetricsHandler(hostname string, goCollector, processCollector bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reg := prometheus.NewPedanticRegistry()
		appFunc := app_info.AppInfoGaugeFunc(app_name, version, hostname)
		reg.MustRegister(appFunc)

		if hostname != "" {
			// Backward compatible: exporter self-metrics + target metrics
			ex, err := exporter.NewAwairExporter(hostname)
			if err != nil {
				http.Error(w, "Failed to connect to Awair device: "+err.Error(), http.StatusBadGateway)
				return
			}
			reg.MustRegister(ex)
		}
		if goCollector {
			reg.MustRegister(collectors.NewGoCollector())
		}
		if processCollector {
			reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	goCollector := flag.Bool("gocollector", false, "enables go stats exporter")
	processCollector := flag.Bool("processcollector", false, "enables process stats exporter")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Warn().Err(err).Msg("No .env file loaded")
	}

	hostname := os.Getenv("AWAIR_HOSTNAME")
	if hostname != "" {
		log.Warn().Msg(
			"[DEPRECATION] AWAIR_HOSTNAME is set. In a future " +
				"version, only multi-target mode via /probe will be " +
				"supported. Please update your Prometheus config to " +
				"use /probe?target=<ip>.",
		)
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

	router := http.NewServeMux()
	router.Handle("/healthz", newHealthCheckHandler())
	router.Handle("/probe", newProbeHandler())
	router.Handle("/metrics", newMetricsHandler(hostname, *goCollector, *processCollector))

	srv.Addr = ":8080"
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}
	<-idleConnsClosed
}
