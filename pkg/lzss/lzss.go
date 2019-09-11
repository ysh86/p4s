package lzss

import (
	"encoding/binary"
	"io"
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
