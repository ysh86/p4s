package crc16

import "io"

// Calc calculates a CRC16 check sum
func Calc(fsrc io.Reader) (crc16 uint16, err error) {
	b := [1]byte{}
	crc16 = 0xffff
	for {
		_, err = io.ReadFull(fsrc, b[:])
		if err != nil {
			break
		}

		b8 := uint16(b[0])
		temp := (crc16 ^ b8) & 0xff
		for i := 0; i < 8; i++ {
			if temp&1 == 1 {
				temp = 0xc7ed ^ (temp >> 1)
			} else {
				temp >>= 1
			}
		}
		crc16 = temp ^ (crc16 >> 8)
	}
	crc16 ^= 0xffff

	return
}
