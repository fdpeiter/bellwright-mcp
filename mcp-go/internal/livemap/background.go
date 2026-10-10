package livemap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"math"
	"os"
)

type Background struct {
	ImageURL         string      `json:"image_url"`
	Width            int         `json:"width"`
	Height           int         `json:"height"`
	SHA256           string      `json:"sha256"`
	SourcePackage    string      `json:"source_package"`
	ProjectionStatus string      `json:"projection_status"`
	Matrix           *[6]float64 `json:"matrix,omitempty"`
	Evidence         []string    `json:"evidence"`
}

func (s Server) Background() (Background, error) {
	var b Background
	if s.BackgroundPath == "" || s.CalibrationPath == "" {
		return b, fmt.Errorf("background not configured")
	}
	cf, e := os.Open(s.CalibrationPath)
	if e != nil {
		return b, e
	}
	defer cf.Close()
	raw, e := io.ReadAll(io.LimitReader(cf, 65537))
	if e != nil {
		return b, e
	}
	if len(raw) > 65536 {
		return b, fmt.Errorf("oversized calibration")
	}
	if e = json.Unmarshal(raw, &b); e != nil {
		return b, e
	}
	if b.SourcePackage == "" || len(b.Evidence) == 0 {
		return b, fmt.Errorf("background provenance required")
	}
	f, e := os.Open(s.BackgroundPath)
	if e != nil {
		return b, e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() > 64*1024*1024 {
		return b, fmt.Errorf("invalid bounded background file")
	}
	cfg, format, e := image.DecodeConfig(f)
	if e != nil || format != "png" || cfg.Width != b.Width || cfg.Height != b.Height {
		return b, fmt.Errorf("background dimensions mismatch")
	}
	if _, e = f.Seek(0, 0); e != nil {
		return b, e
	}
	hash := sha256.New()
	if _, e = io.Copy(hash, f); e != nil {
		return b, e
	}
	if hex.EncodeToString(hash.Sum(nil)) != b.SHA256 {
		return b, fmt.Errorf("background hash mismatch")
	}
	if b.ProjectionStatus == "verified" {
		if b.Matrix == nil {
			return b, fmt.Errorf("verified projection requires matrix")
		}
		for _, v := range b.Matrix {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return b, fmt.Errorf("nonfinite calibration")
			}
		}
		if math.Abs(b.Matrix[0]*b.Matrix[4]-b.Matrix[1]*b.Matrix[3]) < 1e-15 {
			return b, fmt.Errorf("degenerate projection")
		}
	} else {
		b.Matrix = nil
	}
	b.ImageURL = "/map-background.png"
	return b, nil
}
