package main

import (
	"math"
	"testing"
)

func TestNativePixelPreservesWorldAlignment(t *testing.T) {
	for _, p := range [][2]float64{{500, 0}, {106135.7, -86690.29}, {-200000, 200000}} {
		u := 8.642874e-3*p[0] + 3.269422e-5*p[1] + 2050.02
		v := -1.651886e-4*p[0] + 8.595345e-3*p[1] + 2069.67
		gotX, gotY := nativePixel(u, v)
		wantX, wantY := (p[0]-500)*8192/514000+4096, p[1]*8192/514000+4096
		if math.Abs(gotX-wantX) > 1e-8 || math.Abs(gotY-wantY) > 1e-8 {
			t.Fatalf("world %v moved: (%f,%f) vs (%f,%f)", p, gotX, gotY, wantX, wantY)
		}
	}
}
