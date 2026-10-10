// Executed inside companion by preflight. Never emit private response bodies.
export function classifyLive(data) {
  const s = data?.status;
  if (!s || !Array.isArray(data.points) || typeof s.stale !== 'boolean' ||
      typeof s.waiting_for_player !== 'boolean' || !Number.isInteger(s.invalid_records)) {
    throw new Error('Invalid live response');
  }
  if (!s.captured_at) return '[PROBE_UNAVAILABLE] No accepted protocol 0.3 records; verify probe deployment/log path.';
  if (s.stale) return '[FEED_STALE] Source capture older than 10 seconds; game may be inactive or probe stopped.';
  if (s.waiting_for_player) return '[WAITING_FOR_PLAYER] Fresh heartbeat, no loaded player.';
  if (!data.points.some(p => p.kind === 'player' && p.stale === false)) {
    return '[PLAYER_UNAVAILABLE] Feed active but no fresh player point.';
  }
  return '[LIVE_FRESH] Feed and player timestamps are fresh; movement/session lifecycle still require manual verification.';
}

export async function checkServices(fetcher = fetch, report = console.log) {
  const checks = [
    ['COMPANION', 'http://127.0.0.1:8710/'],
    ['MAP_BACKGROUND', 'http://127.0.0.1:8710/api/map/background'],
    ['LIVE_MAP', 'http://127.0.0.1:18126/api/map'],
    ['CALIBRATION', 'http://127.0.0.1:18126/api/background'],
    ['RESOURCE_BASELINE', 'http://127.0.0.1:18126/api/resources'],
    ['PLACEMENTS', 'http://127.0.0.1:18126/api/placements'],
  ];
  let failed = false;
  for (const [name, url] of checks) {
    let response;
    try {
      response = await fetcher(url, {signal: AbortSignal.timeout(5000)});
    } catch {
      report(`[SERVICE_UNAVAILABLE] ${name}: check containers and shared network attachment.`);
      failed = true;
      continue;
    }
    try {
      if (!response.ok) {
        report(`[SERVICE_ERROR] ${name}: HTTP ${response.status}`);
        failed = true;
      } else if (name === 'LIVE_MAP') {
        const data = await response.json();
        report(classifyLive(data));
        if (data.status.invalid_records) report('[PROBE_INVALID_RECORDS] Parser rejected records; inspect private log locally.');
      } else {
        report(`[SERVICE_OK] ${name}`);
      }
    } catch {
      report(`[SERVICE_INVALID_RESPONSE] ${name}: inspect service compatibility privately.`);
      failed = true;
    } finally {
      await response.body?.cancel().catch(() => {});
    }
  }
  return failed ? 2 : 0;
}
