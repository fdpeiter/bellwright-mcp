package main

import (
	"flag"
	"fmt"
	"github.com/claygorman/bellwright-companion/mcp-go/internal/livemap"
	"net/http"
	"os"
	"time"
)

func main() {
	log := flag.String("log", "", "configured UE4SS log path")
	port := flag.Int("port", 18126, "loopback port")
	background := flag.String("background", "", "fixed extracted map PNG path")
	calibration := flag.String("calibration", "", "fixed map provenance/calibration JSON")
	resources := flag.String("resources", "", "fixed packaged resource baseline JSON, availability unknown")
	placements := flag.String("placements", "", "fixed provisional foraging placement JSON")
	flag.Parse()
	if *port < 1 || *port > 65535 {
		fmt.Fprintln(os.Stderr, "invalid port")
		os.Exit(1)
	}
	if *log == "" {
		fmt.Fprintln(os.Stderr, "-log continuous UE4SS log required; no save fallback")
		os.Exit(1)
	}
	s := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%d", *port), Handler: (livemap.Server{LogPath: *log, BackgroundPath: *background, CalibrationPath: *calibration, ResourcePath: *resources, PlacementPath: *placements}).Handler(), ReadHeaderTimeout: 5 * time.Second}
	fmt.Fprintf(os.Stderr, "Coordinate map: http://%s\n", s.Addr)
	if e := s.ListenAndServe(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
