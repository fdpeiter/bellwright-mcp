# MCP planning verification

Six new read-only capabilities address #27, #28, #29, #32, #34 and #42, building on PR #46. No screens or overlays changed.

Worker ranking uses exact observed skills and preserves ties; job audits distinguish missing, malformed and empty priorities. Housing keeps individual observations separate from aggregate capacity and withholds exact comparisons when ownership is incomplete. Stock targets are per-call and use one explicit inventory scope, never summed scopes. Item search includes observed positive quantities and occupied slots; formatted identifiers are not verified names. Resource coverage describes loaded packaged nodes, not current availability, complete world coverage or yield.

Lists default to 25, maximum 100, and expose truncation. Partial and null evidence remains unknown. Save ingestion, retrieval and packaged resource evidence remain distinct.

Validation passed: go test ./..., go vet ./..., go build, and capability-smoke.mjs. Integration tests validate actual responses against output schemas for all 20 tools. Real stdio smoke checks every tool, invalid limits, null/stale evidence, independent outages and EOF. It also asserts ranking score/ties, housing unknown observations, stock deficits, identifier provenance and resource bounds/truncation. Module regressions cover malformed/missing evidence, partial capacity and incomplete population/scope data.

Verification uses synthetic fixtures rather than private game saves. No save-write path, private data or new game assets are introduced. PR #46 was merged upstream before this feature PR was published; this batch targets main.
