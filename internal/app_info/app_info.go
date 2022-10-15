package app_info

import (
	"github.com/prometheus/client_golang/prometheus"
)

var ()

func AppInfoGaugeFunc(appName string, appVersion string, deviceHostname string) prometheus.GaugeFunc {
	infoMetricOpts := prometheus.GaugeOpts{
		Namespace: "awair_exporter",
		Name:      "info",
		Help:      "Info about this awair-exporter",
		ConstLabels: prometheus.Labels{
			"app_name":        appName,
			"app_version":     appVersion,
			"device_hostname": deviceHostname,
		},
	}
	return prometheus.NewGaugeFunc(
		infoMetricOpts,
		func() float64 { return 1 },
	)
}
