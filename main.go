package main

import (
	"fmt"
	"io"
	"os"
)

type lbits struct {
	file      io.Reader
	bits64    uint64
	bits64Len uint
}

func newLbits(file io.Reader) (*lbits, error) {
	return &lbits{file, 0, 0}, nil
}

func (l *lbits) fill() error {
	for l.bits64Len <= 56 {
		bits8 := [1]byte{}
		_, err := l.file.Read(bits8[:])
		if err != nil {
			return err
		}

		l.bits64 |= uint64(bits8[0]) << l.bits64Len
		l.bits64Len += 8
	}

	return nil
}

func (l *lbits) read(bits uint) (uint64, error) {
	err := l.fill()
	if err != io.EOF && err != nil {
		return 0, err
	}

	if l.bits64Len == 0 {
		return 0, io.EOF
	}
	if l.bits64Len < bits || 64 < bits {
		return 0, fmt.Errorf("too long")
	}

	ret := l.bits64 & ((1 << bits) - 1)
	l.bits64 >>= bits
	l.bits64Len -= bits

	return ret, nil
}

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
	xor8Offset := uint64(4*12 + 2)
	orgSizeOffset := uint64(4 * 1)
	sizeOffset := uint64(4 * 13)
	crc16Offset := uint64(4 * 2)

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

	// read bits from Y plane
	// 14[bits] * 4 = 56[bits] => 7[bytes]
	b := [5 * 4]byte{}
	n := uint64(0)
	xor8 := uint64(0)
	headerSize := uint64(4 * 15)
	orgSize := uint64(0)
	payloadSize := uint64(0)
	crc16 := uint16(0)
OuterLoop:
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

		// parse header, decrypt payload & dump
		for j := 0; j < 7; j++ {
			bits8 := bits64 & 0xff
			if orgSizeOffset <= n && n < orgSizeOffset+4 {
				orgSize |= bits8 << ((n - orgSizeOffset) * 8)
			}
			if crc16Offset <= n && n < crc16Offset+2 {
				crc16 |= uint16(bits8 << ((n - crc16Offset) * 8))
			}
			if n == xor8Offset {
				xor8 = bits8
			}
			if sizeOffset <= n && n < sizeOffset+4 {
				payloadSize |= bits8 << ((n - sizeOffset) * 8)
			}
			if n >= headerSize {
				bits8 ^= xor8
			}
			_, err := fenc.Write([]byte{byte(bits8)})
			if err != nil {
				panic(err)
			}
			bits64 >>= 8
			n++

			if n >= headerSize+payloadSize {
				break OuterLoop
			}
		}
	}
	err = fenc.Sync()
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "done read: header=%d, payload=%d, total=%d\n", headerSize, payloadSize, n)

	fdec, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer fdec.Close()

	_, err = fenc.Seek(int64(headerSize), io.SeekStart)
	if err != nil {
		panic(err)
	}
	bitreader, err := newLbits(fenc)
	if err != nil {
		panic(err)
	}

	// decode lzss
	bufSizeBits := uint(10)
	bufSize := uint64(1 << bufSizeBits)
	bufSizeMask := bufSize - 1
	lengthBits := uint(5)
	minLength := uint64(3)
	maxLength := uint64(1 << lengthBits)

	codeBuf := make([]byte, bufSize)
	codePos := uint64(0)

	decodedSize := uint64(0)
	for decodedSize < orgSize {
		// flag
		flag, err := bitreader.read(1)
		if err != nil {
			panic(err)
		}

		if flag == 0 {
			// 8bit data
			bits8, err := bitreader.read(8)
			if err != nil {
				panic(err)
			}

			byte1 := byte(bits8 & 0xff)
			codeBuf[codePos] = byte1
			codePos = (codePos + 1) & bufSizeMask

			//fmt.Fprintf(fdec, "0: %c\t(%02x)\n", byte1, byte1)
			_, err = fdec.Write([]byte{byte1})
			if err != nil {
				panic(err)
			}
			decodedSize++
		} else {
			// encoded data
			idx, err := bitreader.read(bufSizeBits)
			if err != nil {
				panic(err)
			}
			len, err := bitreader.read(lengthBits)
			if err != nil {
				panic(err)
			}
			len++

			if len < minLength || maxLength < len {
				panic(fmt.Errorf("too long code"))
			}

			i := ((bufSize - 1) - idx + codePos) & bufSizeMask
			for l := len; l > 0; l-- {
				byte1 := codeBuf[i]
				codeBuf[codePos] = byte1
				codePos = (codePos + 1) & bufSizeMask

				//fmt.Fprintf(fdec, "1: %c\t(%02x),\tidx=%d,\ti=%d\n", byte1, byte1, idx, i)
				_, err = fdec.Write([]byte{byte1})
				if err != nil {
					panic(err)
				}

				i = (i + 1) & bufSizeMask
			}
			decodedSize += len
		}
	}

	fmt.Fprintf(os.Stderr, "done dec: size=%d, decoded=%d\n", orgSize, decodedSize)

	// CRC16
	err = fdec.Sync()
	if err != nil {
		panic(err)
	}
	_, err = fdec.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}

	crc16decoded := uint16(0xffff)
	for {
		b := [1]byte{}
		_, err = fdec.Read(b[:])
		if err != nil {
			break
		}

		// calc
		b8 := uint16(b[0])
		temp := (crc16decoded ^ b8) & 0xff
		for i := 0; i < 8; i++ {
			if temp&1 == 1 {
				temp = 0xc7ed ^ (temp >> 1)
			} else {
				temp >>= 1
			}
		}
		crc16decoded = temp ^ (crc16decoded >> 8)
	}
	crc16decoded ^= 0xffff

	fmt.Fprintf(os.Stderr, "done crc16: org=%04x, decoded=%04x\n", crc16, crc16decoded)
}
