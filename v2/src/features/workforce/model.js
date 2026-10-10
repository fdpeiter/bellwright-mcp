export const SKILLS = { harvest: 'Harvesting', labour: 'Labour', strength: 'Strength' };
export function playerPeople(world) {
  return Array.isArray(world?.npcs) ? world.npcs.filter(p => p?.is_player_npc === true) : null;
}
export function nameOf(person) {
  return [person?.first_name, person?.last_name].filter(v => typeof v === 'string' && v.trim()).join(' ') || 'Unnamed saved villager';
}
export function finite(value) { return typeof value === 'number' && Number.isFinite(value) ? value : null; }
export function evidence(person, skill) {
  const record = person?.skills?.[skill];
  return { level: finite(record?.level), cap: finite(record?.cap), morale: finite(person?.morale),
    priority: finite(person?.job_priorities?.[skill]),
    injuries: Array.isArray(person?.injuries) ? person.injuries : null,
    tools: person?.equipment && typeof person.equipment === 'object' ? Object.entries(person.equipment).filter(([slot, value]) => /tool|hand|weapon/.test(slot) && typeof value === 'string' && value) : null };
}
export function extractionBuildings(world) {
  return Array.isArray(world?.storage?.containers) ? world.storage.containers.filter(c => c?.faction === 'Player' && ['ForagingCamp_C','LoggingCamp_C','MiningCamp_C'].includes(c.cls)) : [];
}
export function shortlist(people, query, skill) {
  return (people || []).filter(p => nameOf(p).toLowerCase().includes(query.toLowerCase()))
    .sort((a,b) => (evidence(b, skill).level ?? -Infinity) - (evidence(a, skill).level ?? -Infinity) || nameOf(a).localeCompare(nameOf(b)));
}
export function proposalRecord(person, building, skill, snapshot) {
  if (person?.is_player_npc !== true || !person?.guid || !building?.id || building.faction !== 'Player' || !Object.hasOwn(SKILLS, skill)) return null;
  return { version: 1, person: person.guid, building: building.id, skill, snapshot: snapshot ?? null, type: 'staffing-proposal', gameWrite: false };
}
