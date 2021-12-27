package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quells/co2-meter/drivers/i2c/scd30"
	"github.com/quells/co2-meter/drivers/spi/ssd1351"
	"github.com/quells/co2-meter/internal/clockface"
	prom "github.com/quells/co2-meter/internal/metrics"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
	"gobot.io/x/gobot/platforms/raspi"
)

var (
	promAddr = flag.String("prom", ":9100", "Address on which to serve Prometheus metrics")
	rotation = flag.Uint("rotation", 0, "Display rotation")
)

func main() {
	flag.Parse()

	go servePrometheus(*promAddr)

	adaptor := raspi.NewAdaptor()
	co2Sensor := scd30.NewDriver(adaptor)
	displayDC := gpio.NewDirectPinDriver(adaptor, "18")    // Pin 18 aka GPIO 24
	displayReset := gpio.NewDirectPinDriver(adaptor, "22") // Pin 22 aka GPIO 25
	display := ssd1351.NewDriver(adaptor, displayDC, displayReset, 128, 128, spi.WithSpeed(4000000))

	cf := clockface.New()

	work := func() {
		if err := setupDisplay(display); err != nil {
			log.Printf("Failed to setup display: %v", err)
		}

		gobot.Every(2*time.Second, func() {
			if reading, err := co2Sensor.GetLevels(); err != nil {
				log.Printf("Failed to check levels: %v", err)
			} else {
				prom.UpdateAir(reading.CO2, reading.Temp, reading.Hum)
			}
		})

		gobot.Every(5*time.Second, func() {
			if err := cf.UpdateReadings(prom.GetLatestAir()); err != nil {
				log.Printf("Failed to update values for display: %v", err)
			} else if err = cf.DrawReadings(display); err != nil {
				log.Printf("Failed to update display: %v", err)
			}
		})

		gobot.Every(1*time.Second, func() {
			if err := cf.DrawClock(display); err != nil {
				log.Printf("Failed to update clock: %v", err)
			}
		})
	}

	robot := gobot.NewRobot(
		"co2-meter",
		[]gobot.Connection{adaptor},
		[]gobot.Device{co2Sensor, displayDC, displayReset, display},
		work,
	)

	robot.Start()
}

func setupDisplay(d *ssd1351.Driver) (err error) {
	if err = d.SetRotation(ssd1351.Rotation(*rotation)); err != nil {
		return err
	}

	return nil
}

func servePrometheus(addr string) {
	log.Printf("Serving Prometheus at %q", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Serving Prometheus metrics: %v", err)
	}
}
