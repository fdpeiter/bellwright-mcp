// Independent save contracts: never infer runtime/save identity or missing stock.
export const RESOURCES = [{ key: 'Berry_C', label: 'Berries' }, { key: 'Mushroom_C', label: 'Mushrooms' }];
export function xy(position) {
  return Array.isArray(position) && position.length >= 2 && position.slice(0, 2).every(n => typeof n === 'number' && Number.isFinite(n)) ? position.slice(0, 2) : null;
}
export function extractors(world) {
  return (world?.storage?.containers || []).filter(c => c.faction === 'Player' && c.cls === 'ForagingCamp_C' && typeof c.id === 'string');
}
export function storages(world) {
  return (world?.storage?.containers || []).filter(c => c.faction === 'Player' && ['Stockpile_C', 'Barn_C'].includes(c.cls) && typeof c.id === 'string');
}
export function stock(container, resource) {
  const n = container?.items?.[resource];
  return typeof n === 'number' && Number.isFinite(n) && n >= 0 ? n : null;
}
export function proposal(params) {
  const values = ['proposalX', 'proposalY'].map(k => params.get(k));
  if (values.some(v => v === null || v.trim() === '')) return null;
  const point = values.map(Number);
  return point.every(n => Number.isFinite(n) && Math.abs(n) <= 1e8) ? point : null;
}
export function distance(a, b) {
  return a && b ? Math.hypot(a[0] - b[0], a[1] - b[1]) : null;
}
export function frame(points) {
  const valid = points.filter(p => xy(p));
  if (!valid.length) return null;
  const xs = valid.map(p => p[0]), ys = valid.map(p => p[1]);
  const size = Math.max(Math.max(...xs) - Math.min(...xs), Math.max(...ys) - Math.min(...ys), 1000) * 1.3;
  const center = [(Math.max(...xs) + Math.min(...xs)) / 2, (Math.max(...ys) + Math.min(...ys)) / 2];
  return { size, project: p => [50 + 400 * (p[0] - center[0]) / size + 200, 250 - 400 * (p[1] - center[1]) / size] };
}
