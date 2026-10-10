import type { GearPreset } from './types.ts';

// Full desired allocation: satisfied slots retain the actual higher/equal-tier
// item; missing or lower-ranked slots demand the preset pick. Manual overrides
// remain outside preset demand and are protected by equipped coverage.
export function desiredGearItem(preset: GearPreset, slot: string, worn: string | null, pick: string): string | null {
  const ranked = preset.slots[slot] ?? [];
  const wornEntry = ranked.find(r => r.item === worn);
  if (worn !== null && !wornEntry) return null;
  const pickRank = ranked.find(r => r.item === pick)?.rank ?? -1;
  return wornEntry && wornEntry.rank >= pickRank ? worn : pick;
}

export function gearCoverage(item: string, need: number, equipped: number, stored: number, reserve: number) {
  const target = Math.max(need, equipped) + reserve;
  const have = equipped + stored;
  return { item, need, equipped, stored, reserve, target, have, deficit: target - have };
}
