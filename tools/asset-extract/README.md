# Owner-local map extraction

This optional .NET 10 CLI reads a single package from your own authorized Bellwright archives using [CUE4Parse](https://github.com/FabianFG/CUE4Parse) (Apache-2.0, with its separate dependency notices). It replaces the old workspace-only helper. It does not attach to the game, dump mappings, supply keys, modify archives or download assets. Source was developed from the local prototype and CUE4Parse's documented archive/texture APIs; it does not assert independent authorship of archive decoding.

Only proceed where your game/Modkit license and permissions allow local extraction. Supply matching lawfully obtained mappings. No encryption-key handling or access-control bypass is implemented. Package restore downloads software dependencies, not game data. Extracted files, mappings, provenance manifests and diagnostics remain private and are not covered by this project's MIT license.

## Build with pinned dependencies

Install .NET SDK 10 and Go as documented in the main repository. From the checkout:

```powershell
$env:DOTNET_GENERATE_ASPNET_CERTIFICATE = 'false'
dotnet restore tools/asset-extract --locked-mode
dotnet build tools/asset-extract -c Release --no-restore
```

`packages.lock.json` pins package content hashes. The tool uses CUE4Parse `1.2.2.202610`, defaults to `GAME_UE5_7`, and has been checked against the owner's installed 0.0.52204 data. Set process `GAME` only when the build requires another supported engine version. A successful mount does not establish compatibility after a game update.

BC1/DXT1 PNG conversion uses the Go standard library without additional image-conversion dependencies. Other texture formats, virtual layers, unaligned tile borders and Crunch-compressed tiles fail closed.

## Texture and resource inputs

Use fresh output paths outside game directories, preferably outside Git. Existing outputs are refused. Example owner paths are placeholders:

```powershell
$paks = 'C:/your/authorized/Bellwright/Content/Paks'
$mappings = 'C:/your/private/Mappings.usmap'
$private = 'C:/your/private/steward-assets'
New-Item -ItemType Directory -Force "$private/exports" | Out-Null

dotnet tools/asset-extract/bin/Release/net10.0/AssetExtract.dll $paks $mappings Karvenia_08_map_overlay.uasset "$private/exports/Karvenia_08_map_overlay.bc1" texture
go -C mcp-go run ./cmd/prepare-texture "$private/exports/Karvenia_08_map_overlay.bc1" "$private/exports/Karvenia_08_map_overlay.png"
go -C mcp-go run ./cmd/prepare-map -input "$private/exports/Karvenia_08_map_overlay.png" -output "$private/exports/Karvenia_08_companion.png"

dotnet tools/asset-extract/bin/Release/net10.0/AssetExtract.dll $paks $mappings Karvenia_08.umap "$private/Karvenia_08.json" json
go -C mcp-go run ./cmd/prepare-resources "$private/Karvenia_08.json" "$private/resource-baseline.json"
```

Package suffixes must identify exactly one mounted file; include more path segments if ambiguous. Failures are generic; optional `STEWARD_EXTRACT_DIAGNOSTICS` points to a private local error log outside archives. Library logs may contain private package paths, so do not publish terminal logs. Each successful archive export creates a `.provenance.json` sidecar with engine, package identity, UTC extraction time, mappings hash and output hash. Record your game build and package hashes separately.

The private `.bc1` container is eight magic bytes `BWBC1\0\0\0`, little-endian uint32 width/height, then row-major BC1 blocks. The adapter reconstructs mip zero of the observed one-layer virtual map, strips aligned tile borders and supports RawGPU and bounded legacy zlib tiles. The Go decoder bounds dimensions to 8192 and requires exact input length. No orientation or color-space transformation is applied. Map projection still needs independent calibration.

`prepare-resources` streams up to a 1 GiB package export and keeps at most 100,000 instances from `MistFoliageDataComponent` classes under `FoliageHarvestable`. Missing/invalid translations, duplicate IDs, empty results and trailing input fail. Its output matches the static live-map baseline schema and includes input SHA-256 and explicit unknowns. It does not apply parent transforms/ground offsets, enumerate every streamed partition, include tree/ore attackables, measure depletion, identify ownership or prove harvestability/yield. Do not label packaged points as fresh runtime resources.

Follow [Docker setup](../../DOCKER-LOCAL.md) for calibration JSON and mount layout. Use `[]` for placement input until a separate evidenced placement analysis exists. Mappings acquisition and live-probe deployment remain owner prerequisites; this tool does not recreate them.
