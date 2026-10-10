export function playerContainers(world) {
  return Array.isArray(world?.storage?.containers) ? world.storage.containers.filter(c => c?.faction === 'Player') : null;
}
export function quantity(container, item) {
  const n = container?.items?.[item];
  return typeof n === 'number' && Number.isFinite(n) && n >= 0 ? n : null;
}
export function resources(containers) {
  return [...new Set((containers || []).flatMap(c => c.items && typeof c.items === 'object' ? Object.keys(c.items) : []))].sort();
}
export function label(raw) {
  return typeof raw === 'string' ? raw.replace(/_C$/, '').replace(/([a-z])([A-Z])/g, '$1 $2').replace(/_/g, ' ') : 'Unknown';
}
export function containerName(c) { return c.name || label(c.cls) + ' · ' + (c.id || 'Unknown ID'); }
export function coordinates(c) {
  return Array.isArray(c?.position) && c.position.length === 3 && c.position.every(n => typeof n === 'number' && Number.isFinite(n)) ? c.position.join(', ') + ' engine units' : 'Unknown';
}
export function distribution(containers, item, query = '', positiveOnly = false) {
  return (containers || []).filter(c => containerName(c).toLowerCase().includes(query.toLowerCase()) && (!positiveOnly || (quantity(c, item) !== null && quantity(c, item) > 0)));
}
