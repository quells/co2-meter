package ssd1351

import (
	"log"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/spi"
)

type command byte

const (
	cmdSetColumn      command = 0x15
	cmdSetRow         command = 0x75
	cmdWriteRAM       command = 0x5C
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

	w, h int
	fb   []uint16

	spi.Config
}

func NewDriver(c spi.Connector, w, h int, options ...func(spi.Config)) *Driver {
	d := &Driver{}

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
	return nil
	// bus := d.GetBusOrDefault(d.connector.GetSpiDefaultBus())

	// d.connector.GetSpiConnection()
}

func (d *Driver) Halt() (err error) {
	return d.sendCommand(cmdDisplayOff)
}

func (d *Driver) Reset() (err error) {
	if err = d.sendCommand(cmdCommandLock, 0x12); err != nil {
		return err
	}
	if err = d.sendCommand(cmdCommandLock, 0xB1); err != nil {
		return err
	}
	if err = d.sendCommand(cmdDisplayOff); err != nil {
		return err
	}
	// 7:4 Oscillator Freq, 3:0 Clock Divider Ratio
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
	var cmd command
	if inverted {
		cmd = cmdInvertDisplay
	} else {
		cmd = cmdNormalDisplay
	}

	return d.sendCommand(cmd)
}

func (d *Driver) Enable() (err error) {
	return d.sendCommand(cmdDisplayOn)
}

func (d *Driver) Disable() (err error) {
	return d.sendCommand(cmdDisplayOff)
}

func (d *Driver) sendCommand(cmd command, args ...byte) (err error) {
	log.Printf("%x %v\n", cmd, args)
	return nil
}
