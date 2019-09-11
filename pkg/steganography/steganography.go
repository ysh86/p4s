package steganography

import (
	"fmt"
	"io"

	"private/p4s/pkg/lzss"
)

func level7(pix uint8) uint64 {
	switch {
	case 0 <= pix && pix < 22:
		return 0
	case 22 <= pix && pix < 64:
		return 1
	case 64 <= pix && pix < 107:
		return 2
	case 107 <= pix && pix < 149:
		return 3
	case 149 <= pix && pix < 192:
		return 4
	case 192 <= pix && pix < 234:
		return 5
	case 234 <= pix && pix <= 255:
		return 6
	}

	panic(fmt.Errorf("pix level: out of range"))
}

// Decode decodes data embedded into Y plane(steganography).
func Decode(fdst io.WriteCloser, fsrc io.Reader, w, h int) (err error) {
	defer fdst.Close()

	// validate
	if w*h%(5*4) != 0 {
		return fmt.Errorf("w*h %% (5*4) must be 0")
	}

	// read bits from Y plane
	// 14[bits] * 4 = 56[bits] => 7[bytes]
	b := [5 * 4]byte{}
	n := uint64(0)
	payloadSize := uint64(0)
OuterLoop:
	for i := 0; i < w*h/5/4; i++ {
		_, err = fsrc.Read(b[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		// dec steganography
		var bits64 uint64
		for j := uint(0); j < 4; j++ {
			bits14 := level7(b[5*j+0]) +
				level7(b[5*j+1])*7 +
				level7(b[5*j+2])*7*7 +
				level7(b[5*j+3])*7*7*7 +
				level7(b[5*j+4])*7*7*7*7
			bits14 &= 0x3fff // fail safe
			bits64 |= bits14 << (14 * j)
		}

		// parse header & dump
		for j := 0; j < 7; j++ {
			bits8 := bits64 & 0xff
			if lzss.PayloadSizeOffset <= n && n < lzss.PayloadSizeOffset+4 {
				payloadSize |= bits8 << ((n - lzss.PayloadSizeOffset) * 8)
			}
			_, err = fdst.Write([]byte{byte(bits8)})
			if err != nil {
				break OuterLoop
			}
			bits64 >>= 8
			n++

			if n >= lzss.HeaderSize+payloadSize {
				break OuterLoop
			}
		}
	}

	return err
}
