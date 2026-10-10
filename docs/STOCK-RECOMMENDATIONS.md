# Read-only stock recommendations

Dashboard: Insights → Stock target planner. It fetches /api/world with cache:no-store on mount, reload, snapshot change and Refresh stock. Only planning assumptions are retained in this browser; targets are recalculated. No target settings or game saves are written.

MCP: recommend_stock_targets accepts plans (1–100), each with exact item, rate_per_hour, replenishment_hours, safety_buffer_fraction (0–2) and integer planned_demand. Limit defaults25/max100. Formula: ceil(rate * hours * (1 + safety) + planned). Output explains demand, safety stock, planned demand, current stored quantity and deficit/surplus, plus snapshot/ingestion/retrieval provenance. Rates and replenishment times are supplied assumptions, not measured consumption. This sizing policy does not optimize global storage capacity or production allocation.

Example: Iron_C rate2/hour, replenishment3hours, safety0.5, planned1 → target10. Stored3 → shortfall7. Reload after stock changes to5 → target10, shortfall5 with inputs retained. All quantities use parsed world-container totals; recruited ownership is not established. Carried/equipped scopes are excluded. Null/malformed quantities remain unknown. Missing keys imply zero only in a valid complete aggregate.

Verification: Go tests and actual MCP output-schema validation, vet, build, real-stdio smoke across21 tools. Node regression tests cover formulas, invalid assumptions, partial/unknown inventory and malformed browser preferences. Production web build/typecheck and existing synthetic production API smoke pass. Browser reload verified retained inputs and changed stock; desktop/mobile layouts and keyboard navigation inspected. UI comparison is recorded separately. All preview evidence is synthetic.
