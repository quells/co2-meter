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
