package clockface

import "github.com/quells/co2-meter/drivers/spi/ssd1351"

type FrameBuffer struct {
	w, h int
	buf  []uint16
}

func NewFrameBuffer(w, h int) *FrameBuffer {
	return &FrameBuffer{
		w:   w,
		h:   h,
		buf: make([]uint16, w*h),
	}
}

func (fb FrameBuffer) Draw(d *ssd1351.Driver, x, y int) error {
	return nil
}
