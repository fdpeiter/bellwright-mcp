# Bellwright Steward MCP (Go)

Read-only stdio MCP server using the official Go SDK. Steward save data and
UE4SS live map data are separate tools with explicit provenance. No save writes,
game commands, uploads or arbitrary URL tool arguments.

Build: `go build -o bellwright-mcp.exe .`
Test: `go test ./...` and `go vet ./...`.

Start the Steward dashboard following its root README (default port 8710).
Its publisher continues shipping saves to `/api/ingest`; this MCP does not replace
the parser or watcher. Start the UE4SS live map helper separately (port 18126).
Either backend may be absent: its tools report errors instead of substituting data.

Launch: `bellwright-mcp.exe --companion-url http://127.0.0.1:8710 --live-url http://127.0.0.1:18126`

Configure your MCP client with the absolute binary path and those arguments.
Transport: stdio. stdout is reserved for MCP; diagnostics go to stderr.
Only HTTP loopback origins are accepted; redirects are rejected, reads time out
at ten seconds and responses are capped at 32MiB. Backend endpoints are fixed.

Tools:
- read_saved_world: latest parsed save (villagers, buildings, inventory, etc.).
- list_save_snapshots: ingestion history.
- read_save_trends: save-derived changes and rates.
- read_companion_telemetry: companion telemetry with age_ms.
- read_live_map: Unreal actor positions with staleness and session metadata.
- read_resource_baseline: packaged node locations, not current availability.
- read_foraging_candidates: provisional node coverage and assumptions.

retrieved_at is when the MCP fetched the endpoint, NOT when the game state was
captured. Respect the backend snapshot timestamps and live capture ages.
Resource output can be large; use a client with adequate tool-result capacity.
The server can read game data; it cannot construct, recruit or change priorities.

The MCP implementation reports the name bellwright-steward-go. Existing client server keys, tool names, executable names and CLI flags remain compatible.
