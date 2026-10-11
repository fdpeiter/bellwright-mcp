// Decode the bounded BC1 mip exported by AssetExtract using only the Go standard library.
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
)

func decode(r io.Reader) (*image.NRGBA, error) {
	header := make([]byte, 16)
	if _, err := io.ReadFull(r, header); err != nil || string(header[:8]) != "BWBC1\x00\x00\x00" {
		return nil, fmt.Errorf("invalid BC1 header")
	}
	width, height := int(binary.LittleEndian.Uint32(header[8:12])), int(binary.LittleEndian.Uint32(header[12:16]))
	if width < 1 || height < 1 || width > 8192 || height > 8192 {
		return nil, fmt.Errorf("unsupported bounded dimensions")
	}
	blocksX, blocksY := (width+3)/4, (height+3)/4
	pixels := image.NewNRGBA(image.Rect(0, 0, width, height))
	block := make([]byte, 8)
	for by := 0; by < blocksY; by++ {
		for bx := 0; bx < blocksX; bx++ {
			if _, err := io.ReadFull(r, block); err != nil {
				return nil, fmt.Errorf("incomplete BC1 mip")
			}
			a, b := binary.LittleEndian.Uint16(block[:2]), binary.LittleEndian.Uint16(block[2:4])
			palette := [4]color.NRGBA{rgb565(a), rgb565(b), {}, {}}
			if a > b {
				palette[2] = mix(palette[0], palette[1], 2, 1, 3)
				palette[3] = mix(palette[0], palette[1], 1, 2, 3)
			} else {
				palette[2] = mix(palette[0], palette[1], 1, 1, 2)
				palette[3] = color.NRGBA{}
			}
			indices := binary.LittleEndian.Uint32(block[4:8])
			for py := 0; py < 4; py++ {
				for px := 0; px < 4; px++ {
					x, y := bx*4+px, by*4+py
					c := palette[indices&3]
					indices >>= 2
					if x < width && y < height {
						pixels.SetNRGBA(x, y, c)
					}
				}
			}
		}
	}
	var extra [1]byte
	if n, err := r.Read(extra[:]); n != 0 || err != io.EOF {
		return nil, fmt.Errorf("trailing mip data")
	}
	return pixels, nil
}
func rgb565(v uint16) color.NRGBA {
	r, g, b := (v>>11)&31, (v>>5)&63, v&31
	return color.NRGBA{uint8((r << 3) | (r >> 2)), uint8((g << 2) | (g >> 4)), uint8((b << 3) | (b >> 2)), 255}
}
func mix(a, b color.NRGBA, x, y, d int) color.NRGBA {
	return color.NRGBA{
		uint8((int(a.R)*x + int(b.R)*y) / d), uint8((int(a.G)*x + int(b.G)*y) / d), uint8((int(a.B)*x + int(b.B)*y) / d), 255}
}
func run(input, output string) error {
	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("input unavailable")
	}
	defer f.Close()
	pixels, err := decode(f)
	if err != nil {
		return err
	}
	target, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("output unavailable or already exists")
	}
	err = png.Encode(target, pixels)
	closeErr := target.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: prepare-texture PRIVATE.bc1 OUTPUT.png")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Prepared PNG from packaged BC1 mip; no projection or runtime claims.")
}
