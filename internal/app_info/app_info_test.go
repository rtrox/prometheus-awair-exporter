package app_info

import (
	"io/ioutil"
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

	cases := []struct {
		name           string
		hostname       string
		expectHostname bool
	}{
		{"with_hostname", "1.2.3.4", true},
		{"without_hostname", "", false},
	}
	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			info := AppInfoGaugeFunc(
				"asdf",
				"v1.1.1",
				cse.hostname,
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

			// Always expect the HELP line
			assert.Regexp(`(?m)^# HELP awair_exporter_info .*[a-zA-Z]+.*$`, string(buf))

			if cse.expectHostname {
				assert.Regexp(`(?m)^awair_exporter_info\{app_name="asdf",app_version="v1.1.1",device_hostname="1.2.3.4"\} 1$`, string(buf))
			} else {
				assert.Regexp(`(?m)^awair_exporter_info\{app_name="asdf",app_version="v1.1.1"\} 1$`, string(buf))
			}
		})
	}
}
