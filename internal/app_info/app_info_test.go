package app_info

import (
	"io/ioutil"
	"regexp"
	"time"

	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"

	"github.com/tj/assert"
)

func TestAllMetricsPopulated(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	info := AppInfoGaugeFunc(
		"asdf",
		"v1.1.1",
		"1.2.3.4",
	)
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(info)
	srv := httptest.NewServer(promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	defer srv.Close()

	c := &http.Client{Timeout: 300 * time.Millisecond}
	resp, err := c.Get(srv.URL)
	require.Nil(err)
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)

	/*
		# HELP awair_exporter_info Info about this awair-exporter
		# TYPE awair_exporter_info gauge
		awair_exporter_info{app_name="<appName>",app_version="<appVersion>",device_hostname="<deviceHostname>"} 1
	*/
	tests := []struct {
		desc  string
		match *regexp.Regexp
	}{
		{"device_info_desc", regexp.MustCompile(`(?m)^# HELP awair_exporter_info .*[a-zA-Z]+.*$`)},
		{"device_info", regexp.MustCompile(`(?m)^awair_exporter_info{app_name="asdf",app_version="v1.1.1",device_hostname="1.2.3.4"} 1$`)},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.True(tt.match.Match(buf), "Regex %s didn't match a line!", tt.match.String())
		})
	}
}
