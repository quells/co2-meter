package clockface

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"fmt"
	"io/ioutil"
)

//go:embed elkgrove.gz
var elkgroveGZ []byte

// Elkgrove is an 8x8 character map.
var Elkgrove CharMap

func init() {
	r, err := gzip.NewReader(bytes.NewReader(elkgroveGZ))
	if err != nil {
		panic(err)
	}

	Elkgrove.data, err = ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	Elkgrove.w = 8
	Elkgrove.h = 8
}

type CharMap struct {
	data []byte
	w, h int
}

func (cm CharMap) Render(msg string, fg, bg uint16, x, y int, fb *FrameBuffer) error {
	if x+len(msg)*cm.w > fb.w {
		return fmt.Errorf("message is too wide")
	}
	if y+cm.h > fb.h {
		return fmt.Errorf("message is too low")
	}

	for cj := 0; cj < cm.h; cj++ {
		fbOffset := x + (y+cj)*fb.w
		for msgIdx := 0; msgIdx < len(msg); msgIdx++ {
			char := msg[msgIdx]
			cmOffset := int(char)*cm.h + cj
			charLine := cm.data[cmOffset]
			for ci := 0; ci < cm.w; ci++ {
				cb := charLine & (1 << ci)
				fbIdx := fbOffset + ci
				if cb == 0 {
					fb.buf[fbIdx] = bg
				} else {
					fb.buf[fbIdx] = fg
				}
			}
			fbOffset += cm.w
		}
	}

	return nil
}
