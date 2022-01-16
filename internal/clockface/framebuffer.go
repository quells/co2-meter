package clockface

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

func (fb *FrameBuffer) Reset() {
	for i := range fb.buf {
		fb.buf[i] = 0
	}
}

func ternary(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}

func (fb *FrameBuffer) Line(stroke uint16, x0, y0, x1, y1 int) {
	// https://circuitcellar.com/resources/bresenhams-algorithm/
	// Code Fragment 3
	dx := ternary(x1 >= x0, x1-x0, x0-x1)
	dy := ternary(y1 >= y0, y0-y1, y1-y0)
	sx := ternary(x0 < x1, 1, -1)
	sy := ternary(y0 < y1, 1, -1)
	err := dx + dy
	x, y := x0, y0

	yoff := y * fb.w
	for {
		idx := x + yoff
		fb.buf[idx] = stroke

		if x == x1 && y == y1 {
			break
		}

		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
			yoff = y * fb.w
		}
	}
}

func (fb *FrameBuffer) Circle(stroke uint16, x0, y0, r int) {
	// https://circuitcellar.com/resources/bresenhams-algorithm/
	// Code Fragment 4
	x := -r
	y := 0
	err := 2 - (2 * r)

	offA := (y0 + y) * fb.w
	offB := (y0 - x) * fb.w
	offC := (y0 - y) * fb.w
	offD := (y0 + x) * fb.w
	for {
		fb.buf[x0-x+offA] = stroke
		fb.buf[x0-y+offB] = stroke
		fb.buf[x0+x+offC] = stroke
		fb.buf[x0+y+offD] = stroke

		prevErr := err
		if prevErr > x {
			x++
			err += x*2 + 1
			offB = (y0 - x) * fb.w
			offD = (y0 + x) * fb.w
		}
		if prevErr <= y {
			y++
			err += y*2 + 1
			offA = (y0 + y) * fb.w
			offC = (y0 - y) * fb.w
		}

		if x >= 0 {
			break
		}
	}
}
