# Private terrain background

The Docker companion map serves `/api/map/background` from the fixed `MAP_BACKGROUND_FILE` path. Compose mounts `MAP_ASSETS_DIR` read-only. The source is the owner-extracted `/Game/Mist/Maps/Karvenia/Karvenia_08/Karvenia_08_map_overlay` texture, not a screenshot. Neither the source nor prepared image belongs in Git.

Prepare the image once, then rebuild the companion:

```powershell
cd mcp-go
go run ./cmd/prepare-map -input "$env:MAP_ASSETS_DIR/exports/Karvenia_08_map_overlay.png" -output "$env:MAP_ASSETS_DIR/exports/Karvenia_08_companion.png"
cd ..
docker compose up -d --build companion
```

Set `MAP_ASSETS_DIR` in the local `.env` (PowerShell does not automatically load that file). The command needs an 8192×8192 PNG and writes a 4096×4096 PNG. Source geometry: map bounds span 514000 engine units; center X=500, Y=0. X-right/Y-down orientation was visually confirmed by the user against the in-game player marker. Native UV orientation has not been independently read from the constructor.

The preparation command resamples native terrain **into the existing companion affine frame**. Save pins, village labels, and community wiki resource pixels are unchanged. The existing wiki affine calibration remains approximate (upstream reports ~0.7% residual); this fix does not upgrade its accuracy or establish resource availability.

The endpoint validates size and PNG dimensions, returns a generic 503 if absent/invalid, and supports private browser caching and ETags. It accepts no client-supplied file paths. The existing UI placeholder remains the fallback.

Related: issue #13, UI #5, private-input setup #7. The map-tab component is intentionally untouched because another agent owns that UI.
