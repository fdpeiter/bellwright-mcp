package livemap

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBackgroundIntegrityAndUnknownOrientation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "map.png")
	f, _ := os.Create(path)
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	f.Close()
	bytes, _ := os.ReadFile(path)
	hash := fmt.Sprintf("%x", sha256.Sum256(bytes))
	cal := filepath.Join(dir, "cal.json")
	matrix := [6]float64{1, 0, 0, 0, 1, 0}
	b := Background{Width: 2, Height: 2, SHA256: hash, SourcePackage: "SyntheticFixture", Evidence: []string{"synthetic test"}, ProjectionStatus: "orientation_pending", Matrix: &matrix}
	write := func() { raw, _ := json.Marshal(b); os.WriteFile(cal, raw, 0600) }
	write()
	s := Server{BackgroundPath: path, CalibrationPath: cal}
	got, e := s.Background()
	if e != nil || got.Matrix != nil {
		t.Fatal("unverified matrix leaked", got, e)
	}
	r := httptest.NewRequest("GET", "http://127.0.0.1/map-background.png", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 || w.Header().Get("Content-Type") != "image/png" {
		t.Fatal(w.Code)
	}
	b.SHA256 = "incorrect"
	write()
	if _, e = s.Background(); e == nil {
		t.Fatal("hash mismatch accepted")
	}
	b.SHA256 = hash
	b.ProjectionStatus = "verified"
	matrix = [6]float64{}
	write()
	if _, e = s.Background(); e == nil {
		t.Fatal("degenerate verified projection accepted")
	}
}
