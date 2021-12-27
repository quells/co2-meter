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

func (fb *FrameBuffer) Circle(stroke uint16, x, y, r, t int) {
	ir := r * r
	or := ir + 2*r*t + t*t
	for j := 0; j < fb.h; j++ {
		jx := j - x
		j2 := jx*jx + x - j
		for i := 0; i < fb.w; i++ {
			iy := i - y
			i2 := iy*iy + y - i
			sq := i2 + j2

			if ir <= sq && sq <= or {
				idx := i + j*fb.w
				fb.buf[idx] = stroke
			}
		}
	}
}
