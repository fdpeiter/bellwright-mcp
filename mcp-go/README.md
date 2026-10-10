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
- read_resource_baseline: bounded packaged node locations with category and spatial filters.
- read_foraging_candidates: bounded provisional placements with target and spatial filters.
- read_gear_plan: bounded craft shortages, re-equipping candidates and compact presets.
- read_item_history: exact item class history with nullable observations and lineage.
- search_saved_villagers: recruited name/GUID, skill, injury, morale, housing and priority filters.
- locate_saved_inventory: exact saved item locations, separate stored/carried/equipped scopes.
- inspect_saved_squads: saved membership, equipment and unresolved GUIDs.
- read_saved_village_progress: saved trust, prosperity and liberation, including missing/empty states.
- read_connection_health: independent save, companion telemetry and live capture health.

retrieved_at is when the MCP fetched the endpoint, NOT when the game state was
captured. Respect the backend snapshot timestamps and live capture ages.
The server can read game data; it cannot construct, recruit or change priorities.

## Bounded resource and placement queries

Both tools accept `limit` (integer 1–100, default 25), an optional `center`
object with required `x`/`y` and positive `radius`, and/or `bounds` with required
`min_x`, `min_y`, `max_x`, `max_y`. Coordinates and radius use engine units;
conversion to meters is unverified. Radius uses horizontal XY distance and ignores
Z. Bounds and radius are inclusive and intersect when supplied together. Supply
center and radius together; inverted bounds, unknown parameters, null values and
invalid limits produce MCP errors. Coordinates are finite and within ±1e8;
radius is at most 1e8. Empty arguments remain supported.

`read_resource_baseline` also accepts `category`, an exact `baseline_category`
match (up to 128 UTF-8 bytes). `read_foraging_candidates` accepts
`target_runtime_actor_id` (up to 256 UTF-8 bytes); its spatial filters apply to
`practical_candidate.position`, while both current and proposed site evidence
and original assumptions remain in each returned item. Omitted or empty string
filters match all. Results retain backend order; a radius query does not sort
nearest first. The server reads the fixed endpoint before filtering; its 32MiB
backend response cap still applies. It never infers the query center from stale
live or save positions.

Example tool arguments:

```json
{"category":"berries","center":{"x":10000,"y":20000},"radius":5000,"limit":10}
```

```json
{"bounds":{"min_x":0,"min_y":0,"max_x":50000,"max_y":50000},"limit":5}
```

Structured results expose typed `items`, `total` loaded items, `matched` before
the limit, `returned`, `truncated` (omitted matching items), and effective `limit`.
`total` and `matched` describe loaded evidence, not complete world coverage.
An empty match is not proof that resources do not exist. Each result includes
source, backend schema, retrieval time, unknowns, and `availability: "unknown"`.
Static resource positions never establish depletion, regrowth, seasonal
availability or measured yield; placement node counts never establish production,
buildability or a hauling route. `retrieved_at` is retrieval time, not a fresh
game capture. Missing/malformed evidence fails rather than becoming zero positions.

`bounded_queries_test.go` initializes an MCP SDK client/server transport,
discovers tools, calls both tools with synthetic loopback fixtures, and validates
their structured results against the actual advertised output schemas. It covers
default/hard limits, counts, filters, invalid arguments and missing position evidence.

The MCP implementation reports the name bellwright-steward-go. Existing client server keys, tool names, executable names and CLI flags remain compatible.

## Expanded save and health queries

```json
{"name":"read_gear_plan","arguments":{"section":"craft","limit":10}}
{"name":"read_item_history","arguments":{"item_id":"Iron_C","limit":20}}
{"name":"search_saved_villagers","arguments":{"skill":"craft","min_level":3,"limit":10}}
{"name":"locate_saved_inventory","arguments":{"item":"Iron_C","limit":10}}
{"name":"inspect_saved_squads","arguments":{"limit":5,"member_limit":10}}
{"name":"read_saved_village_progress","arguments":{"name":"","limit":10}}
{"name":"read_connection_health","arguments":{}}
```

Query limits default to 25 and allow 1–100. Gear bounds each selected section
independently; squads additionally bound members per group. Inventory faction
and container filters affect container matches; its stored/carried/equipped
aggregates retain their world-wide saved scope. Missing aggregate/equipment
observations remain null. Villager total counts only NPCs explicitly marked
`is_player_npc: true` in the save. Missing optional squad/village evidence is
different from an observed empty collection.

Gear uses saved preset choices and an inventory heuristic (`stored > 0` or worn
by two villagers), falling back to the ideal preset choice. This does not prove
research unlocks. Shortages include satisfied and outstanding allocations;
manual overrides and higher-tier worn gear are retained. Re-equipping candidates
can compete for the same stored item. Private configuration paths are omitted.

Item history returns the latest limited points from the backend's sampled
series, in chronological order. The backend targets 240 points and also retains
the last snapshot, so it can return 241. Sampling metadata is explicit; missing
observed quantity/playtime remains null. Legacy backends cannot provide these
nullable observations and lineage. Check save/character/region/map identity
before inferring trends; this tool does not compute production or consumption.

Health reports available, no_data or unavailable independently for each source.
Companion telemetry age is receipt age and cannot establish source freshness.
Live freshness uses capture status/age, never a save fallback. Missing/unknown
fields remain null; sanitized health output omits backend errors and local paths.

## Synthetic verification

Run `go test ./...`, `go vet ./...`, build the binary, then run
`node capability-smoke.mjs <absolute-binary-path>` with Node 24+. The synthetic
fixture exercises all 14 tools over actual stdio, invalid limits, stale/null
evidence, independent backend outages and clean EOF. Go integration tests
validate structured results against all advertised output schemas. No game or
private files are required. The Docker smoke uses `SMOKE_ITEM_ID` for the required
inventory/history item (default `Iron_C`); success does not prove fresh gameplay.

Expanded tools appear after rebuilding the MCP and restarting/reconnecting its
client process. See [the capability backlog](../docs/MCP-CAPABILITY-BACKLOG.md)
for all 32 ideas; investigative ideas are not exposed as implemented tools.
