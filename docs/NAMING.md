# Product naming proposal

Recommendation: **Bellwright Steward**.

A steward manages a settlement's people, stores and work. This covers the dashboard, save analysis, live telemetry, planning and MCP access without implying that the tool already controls the game. It distinguishes this fork from upstream Bellwright Companion while keeping the game association obvious.

Suggested short display name: **Steward**. Suggested future repository slug: `bellwright-steward`. Suggested description: **Settlement insights and live game data for Bellwright.**

This is a proposal, not an applied rename. Keep the current repository, MCP server identity, package names and Docker service names until the user chooses a name.

## Alternatives

| Name | Strength | Tradeoff |
| --- | --- | --- |
| Bellwright Hearthwatch | Distinctive medieval feel; fits observing a village | Sounds more focused on defense than logistics |
| Bellwright Village Ledger | Clear association with data and settlement records | Longer and less suitable for live planning features |

## Collision check

On 2026-10-10, exact phrase web searches for Bellwright Steward and Bellwright Hearthwatch returned no obvious competing software project in the returned results. Search indexing is incomplete; this does not establish name availability or trademark clearance. Recheck repository, package registry and domain availability immediately before publishing under the selected name. Avoid unqualified Steward as a globally unique package name.

The known upstream project remains [Bellwright Companion by claygorman](https://github.com/claygorman/bellwright-companion). Preserve its author credits, license and provenance after rebranding.

## Migration checklist after selection

- Change the user-visible application title, browser title, README heading and project description together. Retain a clear fork attribution link.
- Inventory existing identifiers before deciding which need migration: Docker service `companion`, volume `companion-data`, frontend workspace `bellwright-companion-web`, MCP Go module `github.com/claygorman/bellwright-companion/mcp-go`, and CLI `--companion-url` are present today.
- Treat a Docker volume rename as a data migration. Back up and preserve existing data; changing a volume key alone can start an empty application.
- Update Go module/import paths and frontend package filters atomically if selected. Run their existing build/test checks.
- Keep the currently connected MCP server key stable during cosmetic rebranding. If its key, executable path or working directory changes, migrate client configs and verify a fresh MCP session can initialize and list tools.
- Retain old CLI flags as aliases for a documented transition if renaming them. Update Compose internal references together if service names change.
- Rename the GitHub repository only after reviewing remote URLs, board links, automation references and client configs. Test redirects and update local remotes explicitly.
- Verify Docker Compose startup, existing settlement data, dashboard routes, publisher ingestion, live-feed freshness and all MCP tools after migration.

## Suggested GitHub issue

Title: **Select product name and migrate branding without breaking Docker or MCP**

Acceptance: selected name recorded; upstream attribution preserved; visible branding consistent; existing persistent data retained; configured MCP connection still initializes and lists tools; package/build references pass checks. Repository and runtime identifier changes should be separately reviewed from display-name edits.

## Selected name and applied compatibility plan

The user selected **Bellwright Steward** on 10 October 2026. The application header, browser metadata, README introduction and MCP implementation display name use that name. The MIT license and upstream attribution remain intact.

Repository slug, frontend package name, Go module/import paths, CLI flags, configured MCP client identifier, Compose project `bellwright`, services and persistent `companion-data` volume remain stable. Any future identifier migration is separate work with explicit data and client migration checks.
