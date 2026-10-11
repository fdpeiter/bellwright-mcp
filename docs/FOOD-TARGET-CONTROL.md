# Live food MCP extension

This first implementation adds `read_live_food_state`, an isolated stdio MCP
extension backed only by `StewardFood` runtime observations. It does not write
game settings or saves. The inherited Bellwright Companion code and license
remain attributed; Bellwright game content belongs to Donkey Crew. No UI is changed.

## Build and connect

From `mcp-go`, run `go build -o food-control ./cmd/food-control`. Start the
result with `-log <private UE4SS.log path>`. The caller configures this path;
tools accept no arbitrary URLs, filenames, function names or properties.
Keep the log and any game type mappings/debug symbols outside Git.

Install `mod/StewardFood/Scripts/main.lua`, `sample.lua` and `storage.lua` as a complete batch
in an enabled `StewardFood/Scripts` mod directory. Initially reload Lua mods
with Ctrl+R. The stable loader owns one game-thread timer and checks its fixed
`sample.lua` file every ten seconds. Valid reader edits automatically replace
the sampling function; no additional timers are registered. Compile/load errors
produce an unavailable result rather than presenting the old reader as updated.
Changes to the loader itself still require a normal UE4SS reload.

UE4SS's installed build registers filesystem-watch directories at program
startup. A mod added while the game is running can be loaded by Ctrl+R but
remain absent from that existing watcher. The module loader avoids repeatedly
asking for global reloads while developing the food reader. See the
[official implementation](https://github.com/UE4SS-RE/RE-UE4SS/blob/5e627997/UE4SS/src/UE4SSProgram.cpp).

## Evidence and bounds

The probe samples seven explicit food-related building classes within 50,000
engine units of the player, treating 500m as unverified. Each class discovery
is capped at 64 objects; orders and storage rules at 32 entries. It reads
settlements referenced by those structures and keeps observed residence distinct
from recruitment and faction references distinct from verified ownership.

Building container slots are selected through actual structure-settlement
references, separately from carried inventories. Container discovery examines
at most 2,048 loaded components and exports at most 32 containers with 64 slots
each. These are bounded observations, not a complete settlement stock census.
Storage Min/Max rules are distinct from production orders. A failed map iteration
is labelled `partial_map_read`; its observed entries are not a complete target list.

Records are limited to 60,000 bytes. Budget trimming explicitly reports omission
and per-domain truncation. The MCP reader scans only the last 8MiB of the private
log, rejects invalid provenance and future captures, and marks observations over
20 seconds old unavailable. Menu observations replace the previous world data.
The MCP always forces `write_available=false`, even if a log claims otherwise.

## Runtime validation on 10 October 2026

The isolated MCP stdio smoke returned fresh, available observations from the
loaded game: 13 food-related structures, one linked settlement, seven NPC
resident records, and building storage slots containing actual food quantities.
This is live evidence, independent of saved snapshots. Full runtime identifiers,
coordinates, inventory contents and raw logs are kept private.

Automatic reader reload was verified by a harmless comment edit: the loader
reported `reader_loaded`, the reader session changed, and subsequent MCP
observations remained fresh without another keypress or game restart.

Go tests and vet, populated MCP output-schema checks, Lua reader tests, Lua reload
tests and actual stdio MCP smoke are required. Tests cover stale/menu transitions,
invalid provenance/future timestamps, missing fields, bounded nearby selection,
automatic replacement without extra timers and invalid replacement source.

## Write stage remains incomplete

The game's shipped symbols identify these real reflected broker RPCs:

- `ServerSetHarvestingOrderAmount(building GUID, order GUID, integer amount)`
- `ServerSetCraftingOrderAmount(building GUID, order GUID, integer amount)`
- `ServerSetInfiniteCraftingAmount(building GUID, item template, integer amount)`

Symbol names alone do not validate the running object's identity or authorize
guessing parameter values. The current reflected harvesting orders expose item
references but not their native GUID/target fields. Cooking-map iteration also
fails partway in the installed Lua runtime. The reflection iterator diagnostics
were removed after demonstrating incompatibility; they do not run in this probe.

Before exposing a writer, obtain verified current building/order GUIDs, complete
target mode/amount evidence and the player's ownership/authority relation. Use
the real RPC with a food-only item allowlist, bounded caps, current session and
capture checks, and an explicit expected old value. Apply one change, then
independently reread it and retain the previous value for rollback. No guessed
memory offsets, save editing, arbitrary reflection commands or unverified setter
calls are included. No food target has been changed by this implementation.

Tracked in issue #60. This extension is a separately reviewable first step while
the active shared MCP/probe has other uncommitted work; it must not overwrite
those files when integrated.
# Storage transfer investigation (issue #62)

The separate `read_live_storage_state` tool returns storage observations under
`data.storage`, together with the same live provenance, session and freshness
checks. It discovers nearby loaded stockpiles and chests independently of food
workstations, retaining faction and settlement references, empty slots, access
flags and exact item custom values. Nearby storage may belong to another faction
or have unknown ownership; discovery does not authorize transfers.

Bounds: 2,048 components examined, 16 storage buildings returned, 128 slots per
container, and up to eight existing-item capacity checks per container. Native
`IsFull`, `GetEmptySlotCount` and `HasSpaceForItems(item, 1)` calls were observed
working in the running game. An empty slot and `InitialCapacity` are insufficient
to infer free volume in dynamically sized stockpiles. Capacity checks preserve
false and zero values and report unavailable methods separately. Output budget
trimming is explicit.

The shipped symbols identify reflected `ServerMove(source, target, operationID)`
and mapped inventory-operation structs, but this does not prove the installed
Lua bridge can marshal those structs safely. A separately approved, single-use
one-item transfer test on 2026-10-11 exited Bellwright with an access violation
and repeated UE4SS stack frames before returning from the RPC. No resulting
inventory change or conservation could be verified. The test hook and arm file
were removed and the installed sampler restored to read-only behavior. The
experiment is not included in this repository or any MCP tool.

**Automatic transfers remain unavailable.** A compatible, verified struct-call
bridge and successful independent source/destination conservation checks are
required before exposing writes. No raw game logs, symbols, private coordinates,
saves or extracted mappings are committed.

## Offline diagnosis of the storage crash

The installed revision is `5e627997`. In its
[struct handler](https://github.com/UE4SS-RE/RE-UE4SS/blob/5e627997/UE4SS/src/LuaType/LuaUObject.cpp#L753),
`push_structproperty` checks userdata at `params.stored_at_index` but retrieves
`UScriptStruct` with no index argument. The
[wrapper defaults](https://github.com/UE4SS-RE/RE-UE4SS/blob/5e627997/deps/first/LuaMadeSimple/include/LuaMadeSimple/LuaMadeSimple.hpp#L749)
that retrieval to stack position 1. Nested fields use position -1.

The test source was a Lua table containing a native `MistItem` userdata in its
`Item` field. During nested conversion, the correct item is on top of the stack
while position 1 still contains the outer source table. This is a concrete
index mismatch, not a guessed memory offset. The wrapper also obtains a userdata
tag without checking that it is present: a failed integer conversion becomes
zero, which matches `LuaOwnedStackObject`, before a pointer is dereferenced.
[Lua's conversion rules](https://www.lua.org/manual/5.4/manual.html#lua_tointegerx)
explain why this can turn a type mismatch into a native null-pointer access.
This path is a strong crash-cause candidate; the unsymbolized crash stack does
not conclusively identify the exact instruction. Lua `pcall` is not protection
against native access violations.

The mapped source and target structs were independently rechecked against the
installed mapping. Source has Type, Actor, Container, SlotIndex, NumItems,
OriginCount and nested MistItem. Target has Type, Actor, Container and SlotIndex.
Mapping establishes field types, not behavioral semantics: container mode,
OriginCount, slot indices, access checks and operation IDs still need evidence
from a normal game-generated transfer. The same omitted index was also present
in upstream main when inspected on 2026-10-10; an update alone is not a verified
remedy.

`node mod/StewardFood/audit_bridge.mjs <source-directory>` reproduces the source
audit without contacting the game. Supply the official LuaUObject.cpp,
LuaMadeSimple.hpp and LuaMadeSimple.cpp from the exact runtime revision. Its
output explicitly does not certify native runtime compatibility.

The next validation sequence is:

1. Observe one ordinary transfer made through the game's UI using a passive
   native-function hook, with bounded field reads and no parameter writes,
   overrides, retained parameter pointers or replay. Verify the hook's struct
   getter path in source before installing it. This establishes real operation
   semantics and completion behavior.
2. Correct and test the native struct handler outside the game, including
   nested userdata, wrong-type rejection, stack balance and missing tags.
   The candidate handler correction passes `params.stored_at_index` to
   `get_userdata<UScriptStruct>`; robust userdata-tag validation is also needed.
   A compatible typed native bridge is an alternative. Neither is installed.
3. Validate a harmless non-mutating call with the same operation-struct shapes
   and inspect actual function metadata. Existing item capacity calls exercise
   a simple direct struct argument, not the failed nested struct path.
4. Only after those checks, run a single non-food transfer against explicit
   current ownership, capacity and count expectations. Verify source decrease,
   destination increase, preserved item/custom values and conserved totals.
   Stop on stale capture, rejected RPC, lost session or missing receipt; never
   automatically retry an uncertain operation.

Representing the nested item as a fully explicit Lua table might avoid the
identified userdata branch, but has not been tested and is not a safe bypass.
Direct Count assignments, spawning/removing items and save editing are not
acceptable storage-transfer implementations.
