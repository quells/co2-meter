package scd30

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_crc8(t *testing.T) {
	tests := []struct {
		input []byte
		want  byte
	}{
		{[]byte{0xBE, 0xEF}, 0x92},
	}
	for _, tt := range tests {
		t.Run(hex.EncodeToString(tt.input), func(t *testing.T) {
			got := crc8(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
