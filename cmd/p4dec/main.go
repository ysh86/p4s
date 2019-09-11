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

	"private/p4s/pkg/lbits"
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

	bitreader, err := lbits.New(flzss, header.Xor8Bits())
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
	for decodedSize < uint64(header.OrgSize) {
		// flag
		flag, err := bitreader.Read(1)
		if err != nil {
			panic(err)
		}

		if flag == 0 {
			// 8bit data
			bits8, err := bitreader.Read(8)
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
			idx, err := bitreader.Read(bufSizeBits)
			if err != nil {
				panic(err)
			}
			len, err := bitreader.Read(lengthBits)
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

	fmt.Fprintf(os.Stderr, "done dec: size=%d, decoded=%d\n", header.OrgSize, decodedSize)

	// CRC16
	err = fdec.Sync()
	if err != nil {
		panic(err)
	}
	_, err = fdec.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	defer fdec.Close()

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
