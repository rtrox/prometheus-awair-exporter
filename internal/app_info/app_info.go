package app_info

import (
	"github.com/prometheus/client_golang/prometheus"
)

var ()

func AppInfoGaugeFunc(appName string, appVersion string, deviceHostname string) prometheus.GaugeFunc {
	labels := prometheus.Labels{
		"app_name":    appName,
		"app_version": appVersion,
	}
	if deviceHostname != "" {
		labels["device_hostname"] = deviceHostname
	}
	infoMetricOpts := prometheus.GaugeOpts{
		Namespace:   "awair_exporter",
		Name:        "info",
		Help:        "Info about this awair-exporter",
		ConstLabels: labels,
	}
	return prometheus.NewGaugeFunc(
		infoMetricOpts,
		func() float64 { return 1 },
	)
}
