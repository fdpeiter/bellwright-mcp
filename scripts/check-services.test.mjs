import {test} from 'node:test';
import assert from 'node:assert/strict';
import {classifyLive, checkServices} from './check-services.mjs';

const data = (status = {}, points = []) => ({
  status: {captured_at: 'synthetic', stale: false, waiting_for_player: false, invalid_records: 0, ...status}, points,
});
test('probe, stale feed, menu and player freshness are distinct', () => {
  assert.match(classifyLive(data({captured_at: '', stale: true})), /^\[PROBE_UNAVAILABLE\]/);
  assert.match(classifyLive(data({stale: true, waiting_for_player: true})), /^\[FEED_STALE\]/);
  assert.match(classifyLive(data({waiting_for_player: true})), /^\[WAITING_FOR_PLAYER\]/);
  assert.match(classifyLive(data({}, [{kind: 'player', stale: true}])), /^\[PLAYER_UNAVAILABLE\]/);
  assert.match(classifyLive(data({}, [{kind: 'player'}])), /^\[PLAYER_UNAVAILABLE\]/);
  assert.match(classifyLive(data({}, [{kind: 'player', stale: false}])), /^\[LIVE_FRESH\]/);
  assert.throws(() => classifyLive({}), /Invalid live response/);
});
test('checks all endpoints without exposing response bodies', async () => {
  const messages = [], urls = [];
  const code = await checkServices(async url => {
    urls.push(url);
    if (url.endsWith('/api/map')) return Response.json(data({stale: true, invalid_records: 1}));
    return new Response('PRIVATE_RESPONSE_SENTINEL');
  }, message => messages.push(message));
  assert.equal(code, 0); // Inactive/stale gameplay does not fail infrastructure.
  assert.equal(urls.length, 6);
  assert.ok(urls.includes('http://127.0.0.1:18126/api/map'));
  assert.ok(messages.some(m => m.startsWith('[PROBE_INVALID_RECORDS]')));
  assert.ok(!messages.join('\n').includes('PRIVATE_RESPONSE_SENTINEL'));
});
test('connection, HTTP and malformed-response failures remain separate', async () => {
  const messages = [];
  const code = await checkServices(async url => {
    if (url.endsWith('/api/resources')) throw Error('PRIVATE_ERROR_SENTINEL');
    if (url.endsWith('/api/map')) return new Response('PRIVATE_INVALID_JSON');
    if (url.endsWith('/api/background')) return new Response('PRIVATE_HTTP_ERROR', {status: 503});
    return new Response('ok');
  }, message => messages.push(message));
  assert.equal(code, 2);
  assert.ok(messages.some(m => m.startsWith('[SERVICE_UNAVAILABLE] RESOURCE_BASELINE')));
  assert.ok(messages.some(m => m.startsWith('[SERVICE_INVALID_RESPONSE] LIVE_MAP')));
  assert.ok(messages.some(m => m.startsWith('[SERVICE_ERROR] CALIBRATION: HTTP 503')));
  assert.ok(!messages.join('\n').includes('PRIVATE_'));
});
