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
	"io"
	"os"

	"github.com/ysh86/p4s/pkg/crc16"
	"github.com/ysh86/p4s/pkg/lzss"
	"github.com/ysh86/p4s/pkg/steganography"
)

func main() {
	fsrc, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer fsrc.Close()

	rlzss, wlzss := io.Pipe()
	defer rlzss.Close()

	///////////////////////////////////////////////
	// steganography: data is embedded into Y plane
	///////////////////////////////////////////////
	w := 720
	h := 720
	go steganography.Decode(wlzss, fsrc, w, h)

	///////////////////////////////////////////////
	// decompress lzss
	///////////////////////////////////////////////
	// read lzss header
	header, err := lzss.ParseHeader(rlzss)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "done read header: header=%d, payload=%d\n", lzss.HeaderSize, header.PayloadSize)
	//fmt.Fprintf(os.Stderr, "%#v\n", header)

	// decode lzss
	fdec, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}

	rcrc, wcrc := io.Pipe()
	defer rcrc.Close()

	go lzss.Decode(fdec, wcrc, rlzss, header)

	///////////////////////////////////////////////
	// CRC16
	///////////////////////////////////////////////
	crc16decoded, err := crc16.Calc(rcrc)
	if err != io.EOF {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "done dec: size=%d\n", header.OrgSize)
	fmt.Fprintf(os.Stderr, "done crc16: org=%04x, decoded=%04x\n", header.CRC16, crc16decoded)
}
