package lzss

import (
	"encoding/binary"
	"fmt"
	"io"

	"private/p4s/pkg/lbits"
)

// Header is the header of lzss file.
type Header struct {
	magic       uint32
	OrgSize     uint32
	CRC16       uint32
	name        [4 * 9]byte
	flags       uint32 // (MSB) comp or not, xor8bits, img seq, # of imgs (LSB)
	PayloadSize uint32
	Width       uint16
	Height      uint16
}

const (
	// HeaderSize is the size of lzss header
	HeaderSize = uint64(4 * 15)

	// PayloadSizeOffset is the file offset of the PayloadSize field
	PayloadSizeOffset = uint64(4 * 13)
)

// ParseHeader reads lzss header from io.Reader and returns the parsed header.
func ParseHeader(flzss io.Reader) (header *Header, err error) {
	var magic uint32
	if err = binary.Read(flzss, binary.LittleEndian, &magic); err != nil {
		return nil, err
	}

	var orgSize uint32
	if err = binary.Read(flzss, binary.LittleEndian, &orgSize); err != nil {
		return nil, err
	}

	var crc16 uint32
	if err = binary.Read(flzss, binary.LittleEndian, &crc16); err != nil {
		return nil, err
	}

	var name [4 * 9]byte
	if _, err = flzss.Read(name[:]); err != nil {
		return nil, err
	}

	var flags uint32
	if err = binary.Read(flzss, binary.LittleEndian, &flags); err != nil {
		return nil, err
	}

	var payloadSize uint32
	if err = binary.Read(flzss, binary.LittleEndian, &payloadSize); err != nil {
		return nil, err
	}

	var w uint16
	if err = binary.Read(flzss, binary.LittleEndian, &w); err != nil {
		return nil, err
	}

	var h uint16
	if err = binary.Read(flzss, binary.LittleEndian, &h); err != nil {
		return nil, err
	}

	header = &Header{
		magic,
		orgSize,
		crc16,
		name,
		flags,
		payloadSize,
		w, h,
	}
	return header, err
}

// Xor8Bits gets xor mask from header
func (h *Header) Xor8Bits() uint8 {
	return uint8((h.flags >> 16) & 0xff)
}

// Decode decompresses lzss
func Decode(fdst io.WriteCloser, fsrc io.Reader, header *Header) error {
	defer fdst.Close()

	bitreader, err := lbits.New(fsrc, header.Xor8Bits())
	if err != nil {
		return err
	}

	bufSizeBits := uint(10)
	bufSize := uint64(1 << bufSizeBits)
	bufSizeMask := bufSize - 1
	lengthBits := uint(5)
	minLength := uint64(3)
	maxLength := uint64(1 << lengthBits)

	codeBuf := make([]byte, bufSize)
	codePos := uint64(0)

	decodedSize := uint64(0)
	for decodedSize < uint64(header.OrgSize) {
		// flag
		flag, err := bitreader.Read(1)
		if err != nil {
			return err
		}

		if flag == 0 {
			// 8bit data
			bits8, err := bitreader.Read(8)
			if err != nil {
				return err
			}

			byte1 := byte(bits8 & 0xff)
			codeBuf[codePos] = byte1
			codePos = (codePos + 1) & bufSizeMask

			//fmt.Fprintf(fdst, "0: %c\t(%02x)\n", byte1, byte1)
			_, err = fdst.Write([]byte{byte1})
			if err != nil {
				return err
			}
			decodedSize++
		} else {
			// encoded data
			idx, err := bitreader.Read(bufSizeBits)
			if err != nil {
				return err
			}
			len, err := bitreader.Read(lengthBits)
			if err != nil {
				return err
			}
			len++

			if len < minLength || maxLength < len {
				return fmt.Errorf("too long code")
			}

			i := ((bufSize - 1) - idx + codePos) & bufSizeMask
			for l := len; l > 0; l-- {
				byte1 := codeBuf[i]
				codeBuf[codePos] = byte1
				codePos = (codePos + 1) & bufSizeMask

				//fmt.Fprintf(fdst, "1: %c\t(%02x),\tidx=%d,\ti=%d\n", byte1, byte1, idx, i)
				_, err = fdst.Write([]byte{byte1})
				if err != nil {
					return err
				}

				i = (i + 1) & bufSizeMask
			}
			decodedSize += len
		}
	}

	return nil
}
