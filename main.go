package main

import (
	"fmt"
	"io"
	"os"

	"github.com/icza/bitio"
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

func main() {
	w := 720
	h := 720
	xor8Offset := 4*12 + 2
	headerSize := 4 * 15

	fsrc, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer fsrc.Close()

	fenc, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer fenc.Close()

	// 14[bits] * 4 = 56[bits] => 7[bytes]
	b := [5 * 4]byte{}
	writer := bitio.NewWriter(fenc)
	n := 0
	var xor8 uint64
	for i := 0; i < w*h/5/4; i++ {
		_, err := fsrc.Read(b[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		// steganography
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

		// decrypt & dump
		for j := 0; j < 7; j++ {
			bits8 := bits64 & 0xff
			if n == xor8Offset {
				xor8 = bits8
			}
			if n >= headerSize {
				bits8 ^= xor8
			}
			err := writer.WriteBits(bits8, 8)
			if err != nil {
				panic(err)
			}
			bits64 >>= 8
			n++
		}
	}
	writer.Close()

	//fenc.Seek(0, os.SEEK_SET)
	//reader := bitio.NewReader(fsrc)

	fmt.Println("done")
}
