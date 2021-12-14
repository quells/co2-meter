package scd30

// crc8 checksum.
// Described on page 14 of SHT spec PDF.
// Test Data 0xBEEF should yield 0x92.
func crc8(data []byte) (crc byte) {
	const poly = 0x31
	crc = 0xFF

	for _, d := range data {
		crc ^= d
		for i := 0; i < 8; i++ {
			if crc&0x80 == 0 {
				crc <<= 1
			} else {
				crc = (crc << 1) ^ poly
			}
		}
	}

	return
}
