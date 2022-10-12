package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	score = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "score"),
		"Awair Score (0-100)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	dew_point = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "dew_point"),
		"The temperature at which water will condense and form into dew (ºC)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	temp = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "temp"),
		"Dry bulb temperature (ºC)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	humidity = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "humidity"),
		"Relative Humidity (%)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	abs_humidity = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "absolute_humidity"),
		"Absolute Humidity (g/m³)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	co2 = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "c02"),
		"Carbon Dioxide (ppm)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	co2_estimated = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "c02_est"),
		"Estimated Carbon Dioxide (ppm - calculated by the TVOC sensor)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	co2_estimate_baseline = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "c02_est_baseline"),
		"A unitless value that represents the baseline from which the TVOC sensor partially derives its estimated (e)CO₂output.",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	voc = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc"),
		"Total Volatile Organic Compounds (ppb)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	voc_baseline = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc_baseline"),
		"A unitless value that represents the baseline from which the TVOC sensor partially derives its TVOC output.",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	voc_h2_raw = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc_h2_raw"),
		"A unitless value that represents the Hydrogen gas signal from which the TVOC sensor partially derives its TVOC output.",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	voc_ethanol_raw = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "voc_ethanol_raw"),
		"A unitless value that represents the Ethanol gas signal from which the TVOC sensor partially derives its TVOC output.",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	pm25 = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "pm25"),
		"Particulate matter less than 2.5 microns in diameter (µg/m³)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
	pm10 = prometheus.NewDesc(
		prometheus.BuildFQName("awair", "", "pm10"),
		"Estimated particulate matter less than 10 microns in diameter (µg/m³ - calculated by the PM2.5 sensor)",
		[]string{"hostname", "uuid", "name"},
		nil,
	)
)

type AwairValues struct {
	Score          float64 `json:"score"`
	DewPoint       float64 `json:"dew_point"`
	Temp           float64 `json:"temp"`
	Humidity       float64 `json:"humid"`
	AbsHumdity     float64 `json:"abs_humid"`
	CO2            float64 `json:"co2"`
	CO2Est         float64 `json:"co2_est"`
	CO2EstBaseline float64 `json:"co2_est_baseline"`
	Voc            float64 `json:"voc"`
	VocBaseline    float64 `json:"voc_baseline"`
	VocH2Raw       float64 `json:"voc_h2_raw"`
	VocEthanolRaw  float64 `json:"voc_ethanol_raw"`
	PM25           float64 `json:"pm25"`
	PM10Est        float64 `json:"pm10_est"`
}

type LEDSettings struct {
	Mode       string
	Brightness int
}

type ConfigResponse struct {
	DeviceUUID      string      `json:"device_uuid"`
	WifiMAC         string      `json:"wifi_mac"`
	SSID            string      `json:"ssid"`
	IP              string      `json:"ip"`
	Netmask         string      `json:"netmask"`
	Gateway         string      `json:"gateway"`
	FirmwareVersion string      `json:"fw_version"`
	Timezone        string      `json:"timezone"`
	Display         string      `json:"display"`
	LED             LEDSettings `json:"led"`
	VocFeatureSet   int         `json:"voc_feature_set"`
}

type AwairExporter struct {
	hostname string
	uuid     string
	name     string
}

func NewAwairExporter(name string, hostname string) (*AwairExporter, error) {
	uri := fmt.Sprintf("http://%s/settings/config/data", hostname)
	log.Info().
		Str("uri", uri).
		Msg("Attempting to connect to Awair device.")

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	config := ConfigResponse{}
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	log.Info().
		Str("device_uuid", config.DeviceUUID).
		Str("IP", config.IP).
		Str("WifiMAC", config.WifiMAC).
		Str("FirmwareVersion", config.FirmwareVersion).
		Msg("Successfully connected to Awair device.")

	return &AwairExporter{
		hostname: hostname,
		uuid:     config.DeviceUUID,
		name:     name,
	}, nil
}

func (e *AwairExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- score
	ch <- dew_point
	ch <- temp
	ch <- humidity
	ch <- abs_humidity
	ch <- co2
	ch <- co2_estimated
	ch <- co2_estimate_baseline
	ch <- voc
	ch <- voc_baseline
	ch <- voc_h2_raw
	ch <- voc_ethanol_raw
	ch <- pm25
	ch <- pm10
}

func (e *AwairExporter) Collect(ch chan<- prometheus.Metric) {
	uri := fmt.Sprintf("http://%s/air-data/latest", e.hostname)
	log.Debug().
		Str("uri", uri).
		Msg("Attempting to retrieve metrics from Awair device.")

	resp, err := http.Get(uri)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error retrieving metrics from Awair device.")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error reading response body.")
		return
	}
	values := AwairValues{}
	err = json.Unmarshal(body, &values)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error unmarshalling JSON response.")
		return
	}
	ch <- prometheus.MustNewConstMetric(
		score, prometheus.GaugeValue, values.Score, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		dew_point, prometheus.GaugeValue, values.DewPoint, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		temp, prometheus.GaugeValue, values.Temp, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		humidity, prometheus.GaugeValue, values.Humidity, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		abs_humidity, prometheus.GaugeValue, values.AbsHumdity, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		co2, prometheus.GaugeValue, values.CO2, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		co2_estimated, prometheus.GaugeValue, values.CO2Est, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		co2_estimate_baseline, prometheus.GaugeValue, values.CO2EstBaseline, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		voc, prometheus.GaugeValue, values.Voc, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		voc_baseline, prometheus.GaugeValue, values.VocBaseline, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		voc_h2_raw, prometheus.GaugeValue, values.VocH2Raw, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		voc_ethanol_raw, prometheus.GaugeValue, values.VocEthanolRaw, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		pm25, prometheus.GaugeValue, values.PM25, e.hostname, e.uuid, e.name,
	)
	ch <- prometheus.MustNewConstMetric(
		pm10, prometheus.GaugeValue, values.PM10Est, e.hostname, e.uuid, e.name,
	)
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	hostname := flag.String("hostname", "", "hostname of Awair device to poll")
	name := flag.String("name", "", "a human-readable name to label metrics")
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
		Str("name", app_name).
		Str("version", version).
		Msg("Exporter Started.")

	// TODO: retries
	exporter, err := NewAwairExporter(*name, *hostname)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed to connect to Awair device.")
	}

	prometheus.MustRegister(exporter)

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	srv.Addr = ":8080"
	srv.Handler = router
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start HTTP Server")
	}
	<-idleConnsClosed
}
