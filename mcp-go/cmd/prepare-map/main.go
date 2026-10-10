// prepare-map resamples an owner-extracted Karvenia texture into the existing
// companion pixel frame. Input/output assets remain private, outside Git.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

func nativePixel(u, v float64) (float64, float64) {
	const a, b, c, d, e, f = 8.642874e-3, 3.269422e-5, 2050.02, -1.651886e-4, 8.595345e-3, 2069.67
	det := a*e - b*d
	x, y := (e*(u-c)-b*(v-f))/det, (-d*(u-c)+a*(v-f))/det
	const scale = 8192.0 / 514000
	return (x-500)*scale + 4096, y*scale + 4096
}

func run(input, output string) error {
	inAbs, _ := filepath.Abs(input)
	outAbs, _ := filepath.Abs(output)
	if input == "" || output == "" || inAbs == outAbs {
		return fmt.Errorf("distinct -input and -output paths required")
	}
	file, err := os.Open(input)
	if err != nil {
		return err
	}
	defer file.Close()
	config, err := png.DecodeConfig(io.LimitReader(file, 64<<20))
	if err != nil {
		return err
	}
	if config.Width != 8192 || config.Height != 8192 {
		return fmt.Errorf("expected 8192x8192 Karvenia PNG")
	}
	if _, err = file.Seek(0, 0); err != nil {
		return err
	}
	source, err := png.Decode(io.LimitReader(file, 64<<20))
	if err != nil {
		return err
	}
	target := image.NewNRGBA(image.Rect(0, 0, 4096, 4096))
	for v := 0; v < 4096; v++ {
		for u := 0; u < 4096; u++ {
			x, y := nativePixel(float64(u)+0.5, float64(v)+0.5)
			if x < 0 || y < 0 || x >= 8191 || y >= 8191 {
				target.SetNRGBA(u, v, color.NRGBA{R: 20, G: 19, B: 17, A: 255})
				continue
			}
			ix, iy := int(x), int(y)
			fx, fy := x-float64(ix), y-float64(iy)
			samples := [4]color.NRGBA{}
			samples[0] = color.NRGBAModel.Convert(source.At(ix, iy)).(color.NRGBA)
			samples[1] = color.NRGBAModel.Convert(source.At(ix+1, iy)).(color.NRGBA)
			samples[2] = color.NRGBAModel.Convert(source.At(ix, iy+1)).(color.NRGBA)
			samples[3] = color.NRGBAModel.Convert(source.At(ix+1, iy+1)).(color.NRGBA)
			blend := func(a, b, c, d uint8) uint8 {
				return uint8((float64(a)*(1-fx)+float64(b)*fx)*(1-fy) + (float64(c)*(1-fx)+float64(d)*fx)*fy + 0.5)
			}
			target.SetNRGBA(u, v, color.NRGBA{R: blend(samples[0].R, samples[1].R, samples[2].R, samples[3].R), G: blend(samples[0].G, samples[1].G, samples[2].G, samples[3].G), B: blend(samples[0].B, samples[1].B, samples[2].B, samples[3].B), A: blend(samples[0].A, samples[1].A, samples[2].A, samples[3].A)})
		}
	}
	if err = os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return err
	}
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	encodeErr := png.Encode(out, target)
	closeErr := out.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}
func main() {
	input := flag.String("input", "", "owner-extracted 8192px terrain PNG")
	output := flag.String("output", "", "private output PNG in companion frame")
	flag.Parse()
	if err := run(*input, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
