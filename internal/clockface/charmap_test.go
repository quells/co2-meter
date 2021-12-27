package clockface

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCharmap_Render(t *testing.T) {
	msg := "Hello, world!"
	fb := NewFrameBuffer(128, 128)
	err := Elkgrove.Render(msg, 0xFFFF, 0x0000, 64-len(msg)*4, 60, fb)
	require.NoError(t, err)

	im := image.NewRGBA(image.Rect(0, 0, fb.w, fb.h))
	fg := color.RGBA{255, 255, 255, 255}
	bg := color.RGBA{0, 0, 0, 255}
	for j := 0; j < fb.h; j++ {
		for i := 0; i < fb.w; i++ {
			idx := i + j*fb.w
			px := fb.buf[idx]
			if px == 0 {
				im.Set(i, j, bg)
			} else {
				im.Set(i, j, fg)
			}
		}
	}

	f, err := os.Create("hello.png")
	require.NoError(t, err)
	defer f.Close()

	require.NoError(t, png.Encode(f, im))
}
