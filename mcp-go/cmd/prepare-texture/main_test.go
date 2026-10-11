package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func fixture(a, b uint16, indices uint32) []byte {
	data := make([]byte, 24)
	copy(data, "BWBC1\x00\x00\x00")
	binary.LittleEndian.PutUint32(data[8:12], 4)
	binary.LittleEndian.PutUint32(data[12:16], 4)
	binary.LittleEndian.PutUint16(data[16:18], a)
	binary.LittleEndian.PutUint16(data[18:20], b)
	binary.LittleEndian.PutUint32(data[20:24], indices)
	return data
}
func TestBC1PaletteAndAlpha(t *testing.T) {
	opaque, err := decode(bytes.NewReader(fixture(0xf800, 0x001f, 0xe4)))
	if err != nil {
		t.Fatal(err)
	}
	want := [][4]uint8{{255, 0, 0, 255}, {0, 0, 255, 255}, {170, 0, 85, 255}, {85, 0, 170, 255}}
	for x, c := range want {
		got := opaque.NRGBAAt(x, 0)
		if [4]uint8{got.R, got.G, got.B, got.A} != c {
			t.Fatalf("palette %d: %+v", x, got)
		}
	}
	alpha, err := decode(bytes.NewReader(fixture(0x001f, 0xf800, 3)))
	if err != nil || alpha.NRGBAAt(0, 0).A != 0 {
		t.Fatal("three-color transparency failed")
	}
}
func TestInvalidMipRejected(t *testing.T) {
	valid := fixture(0xf800, 0x001f, 0)
	huge := append([]byte{}, valid...)
	binary.LittleEndian.PutUint32(huge[8:12], 8193)
	for _, bad := range [][]byte{nil, valid[:23], append(append([]byte{}, valid...), 0), huge} {
		if _, err := decode(bytes.NewReader(bad)); err == nil {
			t.Fatal("invalid mip accepted")
		}
	}
}
