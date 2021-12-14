package scd30

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

type Reading struct {
	CO2  float64 // Carbon dioxide concentration, ppm
	Temp float64 // Temperature, Celsius
	Hum  float64 // Humidity, relative %
}

const (
	defaultAddress = 0x61
	defaultChipID  = 0x60

	resetDelay     = 30 * time.Millisecond
	writeReadDelay = 4 * time.Millisecond
)

type command uint16

const (
	cmdReadMeasurement           command = 0x0300
	cmdContinuousMeasurement     command = 0x0010
	cmdStopMeasurements          command = 0x0104
	cmdSetMeasurementInterval    command = 0x4600
	cmdGetDataReady              command = 0x0202
	cmdAutomaticSelfCalibration  command = 0x5306
	cmdSetForcedRecalibrationRef command = 0x5204
	cmdSetTemperatureOffset      command = 0x5403
	cmdSetAltitudeCompensation   command = 0x5102
	cmdSoftReset                 command = 0xD304
	cmdReadRevision              command = 0xD100
)

// A Driver for the SCD-30 CO2, Temperature, and Humidity sensor.
type Driver struct {
	name       string
	connector  i2c.Connector
	connection i2c.Connection

	i2c.Config
}

func NewDriver(c i2c.Connector, options ...func(i2c.Config)) *Driver {
	d := &Driver{
		name:      gobot.DefaultName("SCD-30"),
		connector: c,
		Config:    i2c.NewConfig(),
	}

	for _, option := range options {
		option(d)
	}

	return d
}

func (d *Driver) Name() string {
	return d.name
}

func (d *Driver) SetName(n string) {
	d.name = n
}

func (d *Driver) Connection() gobot.Connection {
	return d.connection.(gobot.Connection)
}

func (d *Driver) Start() (err error) {
	bus := d.GetBusOrDefault(d.connector.GetDefaultBus())
	address := d.GetAddressOrDefault(defaultAddress)

	if d.connection, err = d.connector.GetConnection(address, bus); err != nil {
		return err
	}

	return d.initialize()
}

func (d *Driver) Halt() (err error) {
	return d.sendCommand(cmdSoftReset)
}

func (d *Driver) Reset() (err error) {
	if err = d.sendCommand(cmdSoftReset); err != nil {
		return err
	}

	time.Sleep(resetDelay)
	return nil
}

func (d *Driver) initialize() (err error) {
	if err = d.Reset(); err != nil {
		return err
	}

	var firmware uint16
	firmware, err = d.GetFirmwareVersion()
	if err != nil {
		firmware, err = d.GetFirmwareVersion()
		if err != nil {
			// Try twice; first I2C transfer after reset usually fails
			return err
		}
	}

	log.Printf("Found firmware version %d for SCD-30", firmware)

	if err = d.StartContinuousMeasurement(); err != nil {
		return err
	}

	if err = d.SetMeasurementInterval(2 * time.Second); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetFirmwareVersion() (uint16, error) {
	return d.readRegister(uint16(cmdReadRevision))
}

func (d *Driver) StartContinuousMeasurement() error {
	return d.sendComandWithArg(cmdContinuousMeasurement, 0)
}

func (d *Driver) SetMeasurementInterval(interval time.Duration) error {
	secs := int(math.Round(interval.Seconds()))
	if secs < 2 || 1800 < secs {
		return fmt.Errorf("invalid measurement interval - must be between 2 seconds and 30 minutes")
	}

	return d.sendComandWithArg(cmdSetMeasurementInterval, uint16(secs))
}

func (d *Driver) UpdateLevels() (reading Reading, err error) {
	var buf []byte
	buf, err = d.read(uint16(cmdReadMeasurement), 18)
	if err != nil {
		return
	}

	// check crc [ ... data data crc8 data data crc8 ... ]
	for i := 0; i < 18; i += 3 {
		if crc8(buf[i:i+2]) != buf[i+2] {
			err = fmt.Errorf("failed checksum")
			return
		}
	}

	co2Data := append(buf[0:2], buf[3:5]...)
	tempData := append(buf[6:8], buf[9:11]...)
	humData := append(buf[12:14], buf[15:17]...)

	co2 := math.Float32frombits(binary.BigEndian.Uint32(co2Data))
	temp := math.Float32frombits(binary.BigEndian.Uint32(tempData))
	hum := math.Float32frombits(binary.BigEndian.Uint32(humData))

	if math.IsNaN(float64(co2)) || math.IsNaN(float64(temp)) || math.IsNaN(float64(hum)) {
		err = fmt.Errorf("invalid reading")
		return
	}

	reading.CO2 = float64(co2)
	reading.Temp = float64(temp)
	reading.Hum = float64(hum)

	return
}

func (d *Driver) sendCommand(cmd command) error {
	buf := make([]byte, 2)
	buf[0] = byte(uint16(cmd) >> 8)
	buf[1] = byte(uint16(cmd) & 0xFF)
	_, err := d.connection.Write(buf)
	return err
}

func (d *Driver) sendComandWithArg(cmd command, arg uint16) error {
	buf := make([]byte, 5)
	buf[0] = byte(uint16(cmd) >> 8)
	buf[1] = byte(uint16(cmd) & 0xFF)
	buf[2] = byte(arg >> 8)
	buf[3] = byte(arg & 0xFF)
	buf[4] = crc8(buf[2:3])
	_, err := d.connection.Write(buf)
	return err
}

func (d *Driver) read(addr uint16, n int) ([]byte, error) {
	{
		buf := make([]byte, 2)
		buf[0] = byte(addr >> 8)
		buf[1] = byte(addr & 0xFF)
		// TODO: stop before reading?
		if _, err := d.connection.Write(buf); err != nil {
			return nil, err
		}
	}

	time.Sleep(writeReadDelay)

	{
		buf := make([]byte, n)
		bytesRead, err := d.connection.Read(buf)
		if err != nil {
			return nil, err
		}
		if bytesRead != n {
			return nil, fmt.Errorf("expected %d bytes, got %d", n, bytesRead)
		}

		return buf, nil
	}
}

func (d *Driver) readRegister(addr uint16) (uint16, error) {
	buf, err := d.read(addr, 2)
	if err != nil {
		return 0, err
	}

	hi := uint16(buf[0]) << 8
	lo := uint16(buf[1])
	return hi | lo, nil
}
