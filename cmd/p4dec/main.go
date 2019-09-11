// ------------------------------------------------
// Usage:
// $ go run main.go data/lzss_720x1280_30fps.yuv data/dst.lzss data/dst_utf16le.txt
// $ iconv -f UTF-16LE data/dst_utf16le.txt > data/dst_utf8.txt
// or
// add BOM UTF-16 LE: 0xFF, 0xFE
// ------------------------------------------------

package main

import (
	"fmt"
	"os"

	"private/p4s/pkg/lzss"
	"private/p4s/pkg/steganography"
)

func main() {
	fsrc, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer fsrc.Close()

	///////////////////////////////////////////////
	// steganography: data is embedded into Y plane
	///////////////////////////////////////////////
	w := 720
	h := 720
	flzss, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	err = steganography.Decode(flzss, fsrc, w, h)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "done read from Y plane\n")

	///////////////////////////////////////////////
	// decompress lzss
	///////////////////////////////////////////////
	flzss, err = os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer flzss.Close()

	info, err := flzss.Stat()
	if err != nil {
		panic(err)
	}

	// read lzss header
	header, err := lzss.ParseHeader(flzss)
	if err != nil {
		panic(err)
	}
	if int64(lzss.HeaderSize)+int64(header.PayloadSize) != info.Size() {
		panic(fmt.Errorf("invalid size of lzss file"))
	}
	fmt.Fprintf(os.Stderr, "done parse header: header=%d, payload=%d, total=%d\n", lzss.HeaderSize, header.PayloadSize, info.Size())
	//fmt.Fprintf(os.Stderr, "%#v\n", header)

	// decode lzss
	fdec, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	err = lzss.Decode(fdec, flzss, header)
	if err != nil {
		panic(err)
	}

	///////////////////////////////////////////////
	// CRC16
	///////////////////////////////////////////////
	fdec, err = os.Open(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer fdec.Close()

	info, err = fdec.Stat()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "done dec: size=%d, decoded=%d\n", header.OrgSize, info.Size())

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

	fmt.Fprintf(os.Stderr, "done crc16: org=%04x, decoded=%04x\n", header.CRC16, crc16decoded)
}
