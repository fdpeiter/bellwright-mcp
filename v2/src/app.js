import * as ui from './ui.js';
const routes = {
  plan: './features/plan/index.js',
  workforce: './features/workforce/index.js',
  supplies: './features/supplies/index.js',
  journal: './features/journal/index.js',
  workspace: './features/workspace/index.js'
};
const cache = new Map();
let tools = [];
let request = 0;
async function loadTool(name, args = {}, { refresh = false } = {}) {
  const key = JSON.stringify([name, args]);
  if (!refresh && cache.has(key)) return cache.get(key);
  try {
    const response = await fetch('/api/tools/call', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name, arguments: args }), signal: AbortSignal.timeout(20000) });
    const result = await response.json();
    if (!response.ok || result.status === 'error') throw new Error(result.error || `MCP request failed (${response.status})`);
    cache.set(key, result);
    return result;
  } catch (error) {
    if (cache.has(key)) {
      const stale = { ...cache.get(key), status: 'stale', error: error.message };
      cache.set(key, stale);
      return stale;
    }
    return { status: 'error', data: null, error: error.message, source: null };
  }
}
async function refreshTools() {
  try {
    const response = await fetch('/api/tools', { signal: AbortSignal.timeout(12000) });
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || 'MCP unavailable');
    tools = data.tools || [];
    document.getElementById('source-context').textContent = data.demo ? 'Synthetic demonstration · No private game data · Production measurements unavailable' : `Read-only MCP · ${tools.length} discovered tools · Saved and runtime evidence are separate`;
  } catch {
    document.getElementById('source-context').textContent = 'MCP connection unavailable · Previously loaded evidence remains accessible · Check Source health';
  }
}
function navigate(route, params = {}) {
  location.hash = route + (Object.keys(params).length ? '?' + new URLSearchParams(params) : '');
}
async function render() {
  const generation = ++request;
  const [rawRoute, query = ''] = location.hash.slice(1).split('?');
  const route = routes[rawRoute] ? rawRoute : 'plan';
  const main = document.getElementById('main');
  main.setAttribute('aria-busy', 'true');
  main.replaceChildren(ui.notice('Loading evidence…'));
  document.getElementById('section').value = route;
  for (const a of document.querySelectorAll('.desktop-nav a, .workspace-link')) {
    if (a.hash === `#${route}`) a.setAttribute('aria-current', 'page'); else a.removeAttribute('aria-current');
  }
  try {
    const feature = await import(routes[route]);
    const content = await feature.render({ ui, loadTool, navigate, params: new URLSearchParams(query), tools, refreshTools });
    if (generation !== request) return;
    main.replaceChildren(content);
  } catch (error) {
    if (generation !== request) return;
    main.replaceChildren(ui.heading('This workspace is being prepared'), ui.notice('The independent v2 feature has not been integrated yet.'), ui.link('Check source health', 'workspace'));
    console.error(error);
  }
  main.setAttribute('aria-busy', 'false');
  main.focus({ preventScroll: true });
}
document.getElementById('section').addEventListener('change', e => navigate(e.target.value));
window.addEventListener('hashchange', render);
await refreshTools();
await render();
