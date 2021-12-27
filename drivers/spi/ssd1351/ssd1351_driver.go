package ssd1351

import (
	"fmt"
	"sync"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/spi"
)

type command byte

const (
	cmdSetColumn      command = 0x15 // left right
	cmdSetRow         command = 0x75 // top bottom
	cmdWriteRAM       command = 0x5C // data...
	cmdReadRAM        command = 0x5D // Not used
	cmdSetRemap       command = 0xA0
	cmdSetStartLine   command = 0xA1
	cmdDisplayOffset  command = 0xA2
	cmdDisplayAllOff  command = 0xA4 // Not used
	cmdDisplayAllow   command = 0xA5 // Not used
	cmdNormalDisplay  command = 0xA6
	cmdInvertDisplay  command = 0xA7
	cmdFunctionSelect command = 0xAB
	cmdDisplayOff     command = 0xAE
	cmdDisplayOn      command = 0xAF
	cmdPrecharge      command = 0xB1
	cmdDisplayEnhance command = 0xB2 // Not used
	cmdClockDivider   command = 0xB3
	cmdSetVSL         command = 0xB4
	cmdSetGPIO        command = 0xB5
	cmdPrecharge2     command = 0xB6
	cmdSetGray        command = 0xB8 // Not used
	cmdUseLUT         command = 0xB9 // Not used
	cmdPrechargeLevel command = 0xBB // Not used
	cmdVCOMH          command = 0xBE
	cmdContrastABC    command = 0xC1
	cmdContrastMaster command = 0xC7
	cmdMuxRatio       command = 0xCA
	cmdCommandLock    command = 0xFD

	cmdHorizontalScroll command = 0x96 // Not used
	cmdStopScroll       command = 0x9E // Not used
	cmdStartScroll      command = 0x9F // Not used
)

type Driver struct {
	name       string
	connector  spi.Connector
	connection spi.Connection
	lock       *sync.Mutex

	dcPin    *gpio.DirectPinDriver
	resetPin *gpio.DirectPinDriver

	w, h int
	r    Rotation

	spi.Config
}

func NewDriver(c spi.Connector, dcPin, resetPin *gpio.DirectPinDriver, w, h int, options ...func(spi.Config)) *Driver {
	d := &Driver{
		name:      gobot.DefaultName("SSD1351"),
		connector: c,
		lock:      new(sync.Mutex),

		dcPin:    dcPin,
		resetPin: resetPin,

		w: w,
		h: h,
		r: RotateUp,

		Config: spi.NewConfig(),
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
	bus := d.GetBusOrDefault(d.connector.GetSpiDefaultBus())
	chip := d.GetChipOrDefault(d.connector.GetSpiDefaultChip())
	mode := d.GetModeOrDefault(d.connector.GetSpiDefaultMode())
	bits := d.GetModeOrDefault(d.connector.GetSpiDefaultBits())
	speed := d.GetSpeedOrDefault(d.connector.GetSpiDefaultMaxSpeed())

	d.connection, err = d.connector.GetSpiConnection(bus, chip, mode, bits, speed)
	if err != nil {
		return err
	}

	if err = d.initialize(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Halt() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if err = d.resetPin.Off(); err != nil {
		return nil
	}
	if err = d.dcPin.Off(); err != nil {
		return nil
	}
	if err = d.sendCommand(cmdDisplayOff); err != nil {
		return nil
	}

	return nil
}

func (d *Driver) initialize() (err error) {
	if err = d.Reset(); err != nil {
		return err
	}
	if err = d.FillScreen(0); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Reset() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if err = d.dcPin.On(); err != nil {
		return err
	}

	if err = d.resetPin.On(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err = d.resetPin.Off(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err = d.resetPin.On(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	if err = d.sendCommand(cmdCommandLock, 0x12); err != nil {
		return err
	}
	if err = d.sendCommand(cmdCommandLock, 0xB1); err != nil {
		return err
	}
	if err = d.sendCommand(cmdDisplayOff); err != nil {
		return err
	}
	// 7:4 Oscillator Freq = 15, 3:0 Clock Divider Ratio = 1
	if err = d.sendCommand(cmdClockDivider, 0xF1); err != nil {
		return err
	}
	if err = d.sendCommand(cmdMuxRatio, 127); err != nil {
		return err
	}
	if err = d.sendCommand(cmdDisplayOffset, 0); err != nil {
		return err
	}
	if err = d.sendCommand(cmdSetGPIO, 0); err != nil {
		return err
	}
	// Internal, diode drop
	if err = d.sendCommand(cmdFunctionSelect, 0x01); err != nil {
		return err
	}
	if err = d.sendCommand(cmdPrecharge, 0x32); err != nil {
		return err
	}
	if err = d.sendCommand(cmdVCOMH, 0x05); err != nil {
		return err
	}
	if err = d.sendCommand(cmdNormalDisplay); err != nil {
		return err
	}
	if err = d.sendCommand(cmdContrastABC, 0xC8, 0x80, 0xC8); err != nil {
		return err
	}
	if err = d.sendCommand(cmdContrastMaster, 0x0F); err != nil {
		return err
	}
	if err = d.sendCommand(cmdSetVSL, 0xA0, 0xB5, 0x55); err != nil {
		return err
	}
	if err = d.sendCommand(cmdPrecharge2, 0x01); err != nil {
		return err
	}
	if err = d.sendCommand(cmdDisplayOn); err != nil {
		return err
	}

	return nil
}

type Rotation int

const (
	// These are probably in the wrong order
	RotateUp Rotation = iota
	RotateRight
	RotateDown
	RotateLeft
)

// SetRotation of display. The image may change immediately, so clearing the
// screen before changing rotation is recommended.
func (d *Driver) SetRotation(r Rotation) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.r = r

	/*
		madctl bitmask
		7,6 color depth (01 = 64k 565 RGB)
		5   odd/even split COM (0 disable, 1 enable)
		4   scan direction (0 top-down, 1 bottom-up)
		3   reserved
		2   color map (0 ABC, 1 CBA)
		1   column map (0 ltr, 1 rtl)
		0   address increment (0 horizontal, 1 vertical)
	*/
	var madctl byte = 0b01100100 // 64k color, enable split, CBA

	switch r {
	case RotateUp:
		madctl |= 0b00010000 // scan bottom-up
	case RotateRight:
		madctl |= 0b00010011 // scan bottom-up, column map rtl, vertical
	case RotateDown:
		madctl |= 0b00000010 // column map rtl
	case RotateLeft:
		madctl |= 0b00000001 // vertical
	default:
		return fmt.Errorf("invalid rotation")
	}

	if err = d.sendCommand(cmdSetRemap, madctl); err != nil {
		return err
	}

	var startLine byte
	switch r {
	case RotateUp, RotateRight:
		startLine = byte(d.h)
	}

	if err = d.sendCommand(cmdSetStartLine, startLine); err != nil {
		return err
	}

	return nil
}

func (d *Driver) SetInverted(inverted bool) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	var cmd command
	if inverted {
		cmd = cmdInvertDisplay
	} else {
		cmd = cmdNormalDisplay
	}

	return d.sendCommand(cmd)
}

// FillScreen with a single color. Default color space is 565 RGB.
func (d *Driver) FillScreen(color uint16) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if err = d.sendCommand(cmdSetColumn, 0, byte(d.w-1)); err != nil {
		return err
	}
	if err = d.sendCommand(cmdSetRow, 0, byte(d.h-1)); err != nil {
		return err
	}

	hi := byte((color & 0xFF00) >> 8)
	lo := byte(color & 0x00FF)
	buf := make([]byte, 2*d.w*d.h)
	for i := 0; i < d.w*d.h; i++ {
		buf[2*i] = hi
		buf[2*i+1] = lo
	}
	return d.sendCommand(cmdWriteRAM, buf...)
}

func (d *Driver) FillHalfScreen(color uint16) (err error) {
	pixels := make([]uint16, d.w*d.h/2)
	for i := range pixels {
		pixels[i] = color
	}
	return d.Write(0, 0, d.w, d.h/2, pixels)
}

func (d *Driver) Write(x, y, w, h int, pixels []uint16) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if len(pixels) != w*h {
		return fmt.Errorf("invalid pixel buffer length")
	}

	var cols, rows int
	if d.r == RotateUp || d.r == RotateDown {
		cols = w - 1
		rows = h - 1
	} else {
		cols = h - 1
		rows = w - 1
		x, y = y, x
	}

	if err = d.sendCommand(cmdSetColumn, byte(x), byte(x+cols)); err != nil {
		return err
	}
	if err = d.sendCommand(cmdSetRow, byte(y), byte(y+rows)); err != nil {
		return err
	}

	buf := make([]byte, len(pixels)*2)
	for i, px := range pixels {
		buf[2*i] = byte((px & 0xFF00) >> 8)
		buf[2*i+1] = byte(px & 0x00FF)
	}
	return d.sendCommand(cmdWriteRAM, buf...)
}

func (d *Driver) Enable() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.sendCommand(cmdDisplayOn)
}

func (d *Driver) Disable() (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.sendCommand(cmdDisplayOff)
}

func (d *Driver) sendCommand(cmd command, args ...byte) (err error) {
	if err = d.dcPin.Off(); err != nil {
		return err
	}
	if err = d.connection.Tx([]byte{byte(cmd)}, nil); err != nil {
		return err
	}
	if err = d.dcPin.On(); err != nil {
		return err
	}

	// SPI library has maximum buffer size
	chunkSize := 4096
	for i := 0; i < len(args); i += chunkSize {
		j := i + chunkSize
		if j > len(args) {
			j = len(args)
		}

		if err = d.connection.Tx(args[i:j], nil); err != nil {
			return err
		}
	}

	return nil
}
