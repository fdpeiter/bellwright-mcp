# Journal and Workspace / #51

New frontend and transport code; no upstream frontend source imported or copied. MCP and save-parser attribution retained. Synthetic fixture is written from invented identities/coordinates/quantities only.

Upstream inspected individually before implementation using GitHub connector on 2026-10-10:
- Events: https://github.com/claygorman/bellwright-companion/blob/main/web/components/bw/events-tab.tsx blob bf437fde17dc5f295d433b635a2476c21573b72d. Heartbeat banner, severity pills and event rows. Replacement is saved-evidence ledger and automatic capture readiness, without severity feed or reader-heartbeat inference.
- Trends: https://github.com/claygorman/bellwright-companion/blob/main/web/components/bw/trends.tsx blob 0bd3b1588b534c7a173bd1a46a5382861bed7023. Many simultaneous charts, runway and produced/consumed net lists. Replacement selects one resource and compares known saved endpoints; explicitly all-world stock change, never production, rate or efficiency.
- Settings: https://github.com/claygorman/bellwright-companion/blob/main/web/components/bw/app.tsx blob fd5bf5cebc19b25887e9cf370c0e4f4761cd6776. Header dropdown with realtime toggle. Replacement full source-health page, separate source provenance and no monitor toggle.

After implementation source composition comparison: v2 contains no reader events/pills, chart dashboard/runway, dropdown settings or inherited interaction sequence. Runtime desktop/mobile/keyboard inspection must be recorded by integrator; source comparison is not visual QA.

Build `node build.mjs`; tests `node --test src/data/transport.test.mjs`; demo `node server.mjs --demo`. Open http://127.0.0.1:18130. Normal mode requires environment STEWARD_MCP_COMMAND (executable only) and optional STEWARD_MCP_ARGS (JSON string array). No shell command interpolation. Optional STEWARD_V2_PORT.

Bridge listens exclusively on 127.0.0.1, validates Host/Origin, limits JSON request bodies, response size, pending calls and timeouts, and exposes a fixed read-only tool allowlist. Browser cannot configure command or origin. Arguments must match each discovered schema, including bounded resource/candidate parameters on newer services. Demo tools take empty arguments. Capture is deliberately unsupported until #54 contract and #44 capabilities are supported. Snapshot history refresh is retrieval of automatically recorded evidence, not manual measurement.

PowerShell real-mode launch: `$env:STEWARD_MCP_COMMAND = 'C:/absolute/path/bellwright-mcp.exe'`; `$env:STEWARD_MCP_ARGS = '[]'`; `node server.mjs`. MCP defaults to existing local companion/live service ports. Host environment, not browser input, selects the executable.
