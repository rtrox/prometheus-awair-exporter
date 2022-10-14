package exporter

import (
	"fmt"
	"io"

	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tj/assert"
)

func init() {
	log.Logger = zerolog.New(io.Discard)
}
func getTestServer() *httptest.Server {
	// Note: the docs at https://support.getawair.com/hc/en-us/articles/360049221014-Awair-Element-Local-API-Feature
	// are incorrect about the type of `voc_feature_set`, the actual return is an int as here.
	configData := `{
		"device_uuid": "awair-element_1",
		"wifi_mac": "70:88:6B:00:00:00",
		"ssid": "Your_AP_Name_Here",
		"ip": "192.168.1.2",
		"netmask": "255.255.255.0",
		"gateway": "192.168.1.1",
		"fw_version": "1.1.4",
		"timezone": "America/Los_Angeles",
		"display": "score",
		"led": {
		  "mode": "sleep",
		  "brightness": 179
		},
		"voc_feature_set": 32
	  }`
	valuesData := `{
		"timestamp": "",
		"score": 89,
		"dew_point": 8.95,
		"temp": 21.13,
		"humid": 45.7,
		"abs_humid": 8.41,
		"co2": 625,
		"co2_est": 563,
		"co2_est_baseline": 35252,
		"voc": 60,
		"voc_baseline": 36539,
		"voc_h2_raw": 25,
		"voc_ethanol_raw": 36,
		"pm25": 40,
		"pm10_est": 42
	  }`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/settings/config/data":
			fmt.Fprint(w, configData)
		case "/air-data/latest":
			fmt.Fprint(w, valuesData)
		default:
			fmt.Printf("Unrecognized Path: %s", r.URL.Path)
			fmt.Fprint(w, "Broken")
		}
	}))
	return srv
}

func exporterFromTestServer(s *httptest.Server) (*AwairExporter, error) {
	e, err := NewAwairExporter(strings.Replace(s.URL, "http://", "", -1))
	if err != nil {
		return nil, err
	}
	return e, nil
}

func TestNewAwairExporter_fail(t *testing.T) {
	_, err := NewAwairExporter("not_a_real_host.not_a_host")
	assert.NotNil(t, err)
}

func TestGetMetrics(t *testing.T) {
	assert := assert.New(t)
	expected := &AwairValues{
		Score:          89,
		DewPoint:       8.95,
		Temp:           21.13,
		Humidity:       45.7,
		AbsHumidity:    8.41,
		CO2:            625,
		CO2Est:         563,
		CO2EstBaseline: 35252,
		Voc:            60,
		VocBaseline:    36539,
		VocH2Raw:       25,
		VocEthanolRaw:  36,
		PM25:           40,
		PM10Est:        42,
	}
	srv := getTestServer()
	defer srv.Close()
	e, err := exporterFromTestServer(srv)
	assert.Nil(err)
	metrics, err := e.GetMetrics()
	assert.Nil(err)
	assert.Equal(expected, metrics, "Metrics don't match!")
}

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	expected := &ConfigResponse{
		DeviceUUID:      "awair-element_1",
		WifiMAC:         "70:88:6B:00:00:00",
		SSID:            "Your_AP_Name_Here",
		IP:              "192.168.1.2",
		Netmask:         "255.255.255.0",
		Gateway:         "192.168.1.1",
		FirmwareVersion: "1.1.4",
		Timezone:        "America/Los_Angeles",
		Display:         "score",
		LED: LEDSettings{
			Mode:       "sleep",
			Brightness: 179,
		},
		VocFeatureSet: 32,
	}
	srv := getTestServer()
	defer srv.Close()
	e, err := exporterFromTestServer(srv)
	assert.Nil(err)
	config, err := e.GetConfig()
	assert.Nil(err)
	assert.Equal(expected, config, "Config doesn't match!")
}

func TestDescribe(t *testing.T) {
	assert := assert.New(t)
	srv := getTestServer()
	defer srv.Close()
	e, err := exporterFromTestServer(srv)
	assert.Nil(err)

	ch := make(chan *prometheus.Desc)
	received := 0
	go func() {
		assert.NotPanics(func() {
			e.Describe(ch)
		})
		close(ch)
	}()
	for elem := range ch {
		assert.NotEqual(&prometheus.Desc{}, elem)
		received++
	}
	assert.GreaterOrEqual(received, 15)
}

func TestCollect(t *testing.T) {
	assert := assert.New(t)
	srv := getTestServer()
	defer srv.Close()
	e, err := exporterFromTestServer(srv)
	assert.Nil(err)

	ch := make(chan prometheus.Metric)
	received := 0
	go func() {
		assert.NotPanics(func() {
			e.Collect(ch)
		})
		close(ch)
	}()

	for elem := range ch {
		metric := &dto.Metric{}
		elem.Write(metric)

		assert.NotEqual(0, metric.GetGauge().GetValue())
		received++
	}
	assert.GreaterOrEqual(received, 15)
}
