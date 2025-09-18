# Prometheus Awair Exporter
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/rtrox/prometheus-awair-exporter) ![Docker Image Version (latest semver)](https://img.shields.io/docker/v/rtrox/prometheus-awair-exporter?label=docker%20image) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rtrox/prometheus-awair-exporter) ![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/rtrox/prometheus-awair-exporter/test.yaml?label=tests) [![Coverage Status](https://coveralls.io/repos/github/rtrox/prometheus-awair-exporter/badge.svg?branch=main)](https://coveralls.io/github/rtrox/prometheus-awair-exporter?branch=main)

Prometheus Awair Exporter connects to an Awair Element device over the Local API, and exports metric via prometheus.

![Grafana Dashboard](.github/dashboard.png "Grafana Dashboard")
## Operating the Exporter

Prometheus-Awair-Exporter requires one environmental variable to be set - `AWAIR_HOSTNAME`, which defines the IP or hostname of the Awair device you wish to monitor. There are also additional flags which can be passed for debugging:
```bash
Usage of ./awair-exporter:
  -debug
        sets log level to debug
  -gocollector
        enables go stats exporter
  -processcollector
        enables process stats exporter
```

So normal usage would be:

```
AWAIR_HOSTNAME=192.168.1.2 ./awair-exporter
```


## Running via Docker

Docker images are available [on DockerHub](https://hub.docker.com/repository/docker/rtrox/prometheus-awair-exporter) and [GitHub Container Registry](https://github.com/users/rtrox/packages/container/package/prometheus-awair-exporter). Example usage:

```bash
docker run -d --name awair-exporter -p 8080:8080 ghcr.io/rtrox/prometheus-awair-exporter:latest
```

### Docker Image Tag Reference

| Tag Example                | Description                                                                 |
|----------------------------|-----------------------------------------------------------------------------|
| `v1.2.3`                   | Semantic version tag for a release (major.minor.patch)                       |
| `v1.2`                      | Major.minor tag for a release                                               |
| `v1`                        | Major tag for a release                                                     |
| `latest`                   | Always points to the latest versioned release                               |
| `sha-abcdef1`              | Image for a specific commit (short SHA)                                     |
| `main`                     | Moving tag for the main branch (latest commit on main, may be unreleased)    |

**Note:**

- Using the versioned
- Use `main` for testing the most recent changes on the main branch.

## Running via Docker Compose

A docker-compose file is provided in the root of this directory. Update your AWAIR_HOSTNAME value to the IP/Hostname of your awair device, and run:

```
docker compose up -d
```

## Running via Kubernetes


Kubernetes manifests are provided in the [`manifests`](kubernetes/manifests/) folder. Update the `AWAIR_HOSTNAME` environment variable in `deployment.yaml`, as well as `awair-exporter/device-name` (the ServiceMonitor will set this as an `awair_exporter_device_name` label on all metrics). From here:

```sh
kubectl apply -f manifests/
```

If you need to run multiple exporters, make sure you adjust the `metadata.name` and `metadata.labels.app.kubernetes.io/name` fields in each deployment to be unique.

No changes need to be made to `service.yaml` or `servicemonitor.yaml` files; they will automatically pick up new deployments for monitoring.

## Example Metric Output

```text
# HELP awair_absolute_humidity Absolute Humidity (g/m³)
# TYPE awair_absolute_humidity gauge
awair_absolute_humidity 7.71
# HELP awair_co2 Carbon Dioxide (ppm)
# TYPE awair_co2 gauge
awair_co2 530
# HELP awair_co2_est Estimated Carbon Dioxide (ppm - calculated by the TVOC sensor)
# TYPE awair_co2_est gauge
awair_co2_est 420
# HELP awair_co2_est_baseline A unitless value that represents the baseline from which the TVOC sensor partially derives its estimated (e)CO₂output.
# TYPE awair_co2_est_baseline gauge
awair_co2_est_baseline 35270
# HELP awair_device_info Info about the awair device
# TYPE awair_device_info gauge
awair_device_info{device_uuid="awair-element_1",firmware_version="1.2.8",voc_feature_set="34"} 1
# HELP awair_dew_point The temperature at which water will condense and form into dew (ºC)
# TYPE awair_dew_point gauge
awair_dew_point 7.58
# HELP awair_humidity Relative Humidity (%)
# TYPE awair_humidity gauge
awair_humidity 46.08
# HELP awair_pm10 Estimated particulate matter less than 10 microns in diameter (µg/m³ - calculated by the PM2.5 sensor)
# TYPE awair_pm10 gauge
awair_pm10 21
# HELP awair_pm25 Particulate matter less than 2.5 microns in diameter (µg/m³)
# TYPE awair_pm25 gauge
awair_pm25 20
# HELP awair_score Awair Score (0-100)
# TYPE awair_score gauge
awair_score 96
# HELP awair_temp Dry bulb temperature (ºC)
# TYPE awair_temp gauge
awair_temp 19.48
# HELP awair_voc Total Volatile Organic Compounds (ppb)
# TYPE awair_voc gauge
awair_voc 98
# HELP awair_voc_baseline A unitless value that represents the baseline from which the TVOC sensor partially derives its TVOC output.
# TYPE awair_voc_baseline gauge
awair_voc_baseline 37378
# HELP awair_voc_ethanol_raw A unitless value that represents the Ethanol gas signal from which the TVOC sensor partially derives its TVOC output.
# TYPE awair_voc_ethanol_raw gauge
awair_voc_ethanol_raw 36
# HELP awair_voc_h2_raw A unitless value that represents the Hydrogen gas signal from which the TVOC sensor partially derives its TVOC output.
# TYPE awair_voc_h2_raw gauge
awair_voc_h2_raw 25
# HELP exporter_info Info about this awair-exporter
# TYPE exporter_info gauge
exporter_info{app_name="awair-exporter",hostname="192.168.1.2",version="x.x.x"} 1
```
