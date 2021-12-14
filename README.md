# Carbon Dioxide Meter

Measure CO2 concentration, temperature, and relative humidity
and record these values as Prometheus metrics.

Uses the [SCD-30](https://www.adafruit.com/product/4867) sensor.
Ports the I2C protocol to [Gobot](https://gobot.io/) from the
[Adafruit Arduino library](https://github.com/adafruit/Adafruit_SCD30).

This is meant to be deployed on a Raspberry Pi Zero W.

## Build and Deploy

```
$ make
$ scp co2-meter pi@zero:/path/to/executable
```

## Update gobot

```
$ go get -d -u gobot.io/x/gobot/...@dev
```

This gets around an incompatibility with gobuffalo's UUID package
when the Gobot version is not specified.
