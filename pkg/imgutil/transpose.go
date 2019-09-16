package imgutil

import (
	"io"
)

// Transpose transposes a image and write to a stream.
func Transpose(fdst io.WriteCloser, fsrc []uint8, stride, w, h int) (err error) {
	defer fdst.Close()
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			n := stride*y + x
			_, err = fdst.Write(fsrc[n : n+1])
			if err != nil {
				return
			}
		}
	}
	return
}
