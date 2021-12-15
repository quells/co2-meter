package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quells/co2-meter/drivers/i2c/scd30"
	prom "github.com/quells/co2-meter/internal/metrics"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	promAddr = flag.String("prom", ":9100", "Address on which to serve Prometheus metrics")
)

func main() {
	flag.Parse()

	go servePrometheus(*promAddr)

	adaptor := raspi.NewAdaptor()
	co2Sensor := scd30.NewDriver(adaptor)

	work := func() {
		gobot.Every(2*time.Second, func() {
			if reading, err := co2Sensor.GetLevels(); err != nil {
				log.Printf("Failed to check levels: %v", err)
			} else {
				prom.Air(reading.CO2, reading.Temp, reading.Hum)
			}
		})
	}

	robot := gobot.NewRobot(
		"co2-meter",
		[]gobot.Connection{adaptor},
		[]gobot.Device{co2Sensor},
		work,
	)

	robot.Start()
}

func servePrometheus(addr string) {
	log.Printf("Serving Prometheus at %q", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Serving Prometheus metrics: %v", err)
	}
}
