# Local Docker setup

Run in this repository:

    docker compose up -d --build companion live-map publisher
    docker compose build mcp

Open http://127.0.0.1:8710. The companion needs an ingested save before showing
settlement data. Use the publisher documented in the upstream README with
http://127.0.0.1:8710/api/ingest. UE4SS remains installed in the Windows game;
the live-map container reads its log through a read-only mount.

Connect an MCP client using mcp-go/mcp-config.example.json. It launches a stdio
container per connection through `docker compose run --rm --no-deps -T mcp`.
Start companion and live-map first. MCP and live-map share the companion network
namespace so configured loopback URLs stay private. The web dashboard is bound
only to Windows localhost. Save data persists in the companion-data volume.

Copy your own paths into .env (ignored by Git): UE4SS_LOG_DIR, MAP_ASSETS_DIR,
RESOURCE_ANALYSIS_DIR. Extracted game assets, logs and saves are not distributed
with this repository. Existing Windows live map can continue independently.

Docker Desktop must be running in Linux-container mode; WSL setup may require
administrator approval/reboot. No game-save writes or gameplay controls.
