# Automatic extractor observations — proposed MCP contract

Status: proposed contract, not implemented service functionality. User requirement: measurements must be recorded automatically through MCP; no manual quantities, worker-hours, activity durations, losses or deliveries. Existing read-only tools cannot produce an attributed extractor productivity window. Keep the UI unsupported until capability and identity gates pass.

## Boundary and capabilities

Game/save access stays read-only. Starting/stopping an observation changes app-owned measurement records, so those control tools must not be advertised as read-only app operations. Store records outside saves and private data outside Git. Browser forwards only allowlisted MCP operations, never process commands or arbitrary origins.

A versioned manifest should report per-building/resource counters and their statuses, supported controls, persistence semantics, units, sample frequency, build/map/session, identity method, complete actor scope and expiry/freshness. Availability is field-specific. Tool presence alone does not prove counter support. Coordinate #44; identity evidence #39; worker inspection #45.

## Proposed operations

Names below are proposals, not existing tools:

- begin_extractor_observation: stable building evidence reference, selected resource IDs, requested duration in simulation time and idempotency token. Returns persistent window ID, immutable conditions, capability status and lifecycle. Reject absent attribution/labor capabilities before recording.
- read_extractor_observation: window ID; current status, capture sequence, source/game-time span, conditions and per-resource results. Includes partial evidence and explicit missing terms.
- stop_extractor_observation: window ID and reason; idempotent stop; never makes incomplete records complete.
- list_extractor_observations: bounded filters and stable pagination, building/session scope, summary only.

Lifecycle: ready, starting, recording, completed, interrupted, failed; unsupported is capability state before begin. Automatic capture resumes only under an explicitly supported session/identity contract. Reload, pause, worker reassignment, node depletion, target fulfillment, session change or missing samples must invalidate affected comparisons or be recorded as condition changes. Duration cannot advance merely because wall-clock did.

## Required evidence

Stable save/runtime building and worker linkage confirmed independently, not suffix/proximity guess. Separate gathered output, output buffer, workers carrying attributable output, completed outgoing deliveries and transfers in. Pending hauling requests are not completed deliveries. Count each resource extraction once; transfer between camp and barn is not second extraction. Assignments establish neither actual work nor worker-hours; capture activity state and simulation-time spans. Distinguish harvesting labor, hauling labor and scheduled/blocked intervals.

For each resource record opening/closing inventory, extracted, transfers in/out, consumption, losses, delivered units, attributable worker game-hours and evidence coverage. Every field is either supported known value or explicit unknown/unsupported; absent never becomes zero. Record counter baseline/reset/overflow handling. Complete reconciliation requires all terms.

Productivity = delivered units / attributable worker game-hours, separately by resource. Zero labor gives undefined productivity. Partial windows with missing labor/deliveries cannot produce a rate. A stock delta never substitutes for delivered output. Efficiency percentage needs a reference under comparable building/resource/staff/tools/targets/availability conditions. Static node counts remain location coverage hypotheses.

## Persistence and validation

Record immutable build/map/session/window IDs, source capture wall-clock, receipt wall-clock, simulation timestamp and counter sequence separately. Persist raw evidence privately and derived provenance. Reconnect must retain previous records without treating cached events as fresh. Bound retention/export; no private coordinates or IDs in public examples.

Validate with synthetic cases for duplicate events, out-of-order capture, skipped sample, counter reset, pause, session switch, repeated begin/stop, no workers, target fulfillment, transfer double count and partial resource support. Controlled real-game evidence is required before measured-output claims. This contract does not authorize undocumented runtime field guesses or game mutation.
