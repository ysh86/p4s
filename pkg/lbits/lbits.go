package lbits

import (
	"fmt"
	"io"
)

// LBits is a bit reader.
type LBits struct {
	file io.Reader
	xor8 uint8

	bits64    uint64
	bits64Len uint
}

// New allocates a new LBits struct.
func New(file io.Reader, xor8 uint8) (*LBits, error) {
	return &LBits{file, xor8, 0, 0}, nil
}

func (l *LBits) fill() error {
	for l.bits64Len <= 56 {
		bits8 := [1]byte{}
		_, err := l.file.Read(bits8[:])
		if err != nil {
			return err
		}

		// pack from LSB to MSB
		// after packing:
		//         MSB                                            LSB
		// bits64 = | ... empty ... | new 8bits | ... old bits ... |
		l.bits64 |= uint64(bits8[0]^l.xor8) << l.bits64Len
		l.bits64Len += 8
	}

	return nil
}

// Read reads some bits from LSB.
func (l *LBits) Read(bits uint) (uint64, error) {
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

	// read from LSB
	//         MSB                                           LSB
	// bits64 = | ... empty ... | ... old bits ... | ret bits |
	ret := l.bits64 & ((1 << bits) - 1)
	l.bits64 >>= bits
	l.bits64Len -= bits

	return ret, nil
}
