package clockface

import (
	"fmt"
	"log"
	"math"
	"net"

	"github.com/quells/co2-meter/drivers/spi/ssd1351"
)

const (
	white  uint16 = 0xFFFF
	black  uint16 = 0x0000
	gray50 uint16 = 0b10000_100000_10000
	gray25 uint16 = 0b01000_010000_01000
	green  uint16 = 0b00000_101111_00000
	yellow uint16 = 0b10111_101111_00000
	orange uint16 = 0b10111_010111_00000
	red    uint16 = 0b10111_000000_00000
)

var jx = []int{
	0, 1, 2, 2, 2,
	1, 1, 0, 0, -1,
	-2, -2, -1, -1, -2,
	-2, -2, -1, -1, 0,
	0, 1, 2, 2, 1,
}

var jy = []int{
	0, 0, 0, -1, -2,
	-2, -1, -1, -2, -2,
	-2, -1, -1, 0, 0,
	1, 2, 2, 1, 1,
	2, 2, 2, 1, 1,
}

type Clockface struct {
	CharMap CharMap

	ip   *FrameBuffer
	dial *FrameBuffer
	temp *FrameBuffer
	co2  *FrameBuffer
	hum  *FrameBuffer

	jt, ji int
}

func New() *Clockface {
	cf := Clockface{
		CharMap: Elkgrove,

		ip:   NewFrameBuffer(120, 8),
		dial: NewFrameBuffer(100, 100),
		temp: NewFrameBuffer(32, 8),
		co2:  NewFrameBuffer(48, 8),
		hum:  NewFrameBuffer(32, 8),
	}

	_ = cf.CharMap.Render(getCurrentIP(), gray50, black, 0, 0, cf.ip)
	cf.dial.Circle(gray25, 50, 50, 48, 1)

	return &cf
}

func (cf *Clockface) UpdateReadings(co2, temp, hum float64) (err error) {
	{
		msg := fmt.Sprintf("%3dF", int(math.Round(temp*1.8+32)))
		if err = cf.CharMap.Render(msg, gray50, black, 0, 0, cf.temp); err != nil {
			return err
		}
	}

	{
		msg := fmt.Sprintf(" %4d ", int(math.Round(co2)))

		var bg uint16
		if co2 == 0 || math.IsNaN(co2) {
			bg = black
		} else if co2 < 800 {
			bg = green
		} else if co2 < 1000 {
			bg = yellow
		} else if co2 < 1200 {
			bg = orange
		} else {
			bg = red
		}

		if err = cf.CharMap.Render(msg, white, bg, 0, 0, cf.co2); err != nil {
			return err
		}
	}

	{
		msg := fmt.Sprintf("%3d%%", int(math.Round(hum)))
		if err = cf.CharMap.Render(msg, gray50, black, 0, 0, cf.hum); err != nil {
			return err
		}
	}

	return
}

func (cf *Clockface) draw(display *ssd1351.Driver, fb *FrameBuffer, x, y int) error {
	return display.Write(x+jx[cf.ji], y+jy[cf.ji], fb.w, fb.h, fb.buf)
}

func (cf *Clockface) DrawReadings(display *ssd1351.Driver) (err error) {
	cf.jt++
	if cf.jt >= 10 {
		cf.jt = 0

		cf.ji++
		if cf.ji >= 25 {
			cf.ji = 0
		}

		if err = display.FillScreen(black); err != nil {
			return err
		}
	}

	if err = cf.draw(display, cf.ip, 4, 2); err != nil {
		return err
	}

	if err = cf.draw(display, cf.temp, 2, 118); err != nil {
		return err
	}
	if err = cf.draw(display, cf.co2, 40, 118); err != nil {
		return err
	}
	if err = cf.draw(display, cf.hum, 94, 118); err != nil {
		return err
	}

	return
}

func (cf *Clockface) DrawClock(display *ssd1351.Driver) (err error) {
	return cf.draw(display, cf.dial, 14, 12)
}

func getCurrentIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Could not get network interfaces: %v", err)
		return ""
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Could not get network addresses for interface: %v", err)
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipv4 := ipnet.IP.To4(); ipv4 != nil {
					return fmt.Sprintf("%03d.%03d.%03d.%03d", ipv4[0], ipv4[1], ipv4[2], ipv4[3])
				}
			}
		}
	}

	return ""
}
