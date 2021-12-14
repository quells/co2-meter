package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var co2Gauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "co2_ppm",
	Help: "Carbon dioxide concentration in parts per million",
})

var tempGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "temp_c",
	Help: "Air temperature in degress Celsius",
})

var humidityGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "relative_humidity_percent",
	Help: "Relative humidity as percentage",
})

func Air(co2, temp, humidity float64) {
	co2Gauge.Set(co2)
	tempGauge.Set(temp)
	humidityGauge.Set(humidity)
}
