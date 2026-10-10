import assert from 'node:assert/strict';
import test from 'node:test';
import { desiredGearItem, gearCoverage } from './gear-planning.ts';

const preset = { key: 'kit', name: 'Kit', slots: { weapon: [{ item: 'SteelSword_C', rank: 2 }, { item: 'Sword_C', rank: 1 }] } };
const demand = worn => worn.reduce((counts, item) => {
  const desired = desiredGearItem(preset, 'weapon', item, 'Sword_C');
  if (desired !== null) counts[desired] = (counts[desired] ?? 0) + 1;
  return counts;
}, {});

test('partially equipped custom squad still needs every missing sword', () => {
  const need = demand([...Array(5).fill('Sword_C'), ...Array(5).fill(null)]);
  assert.equal(need.Sword_C, 10);
  assert.deepEqual(gearCoverage('Sword_C', need.Sword_C, 5, 0, 0), {
    item: 'Sword_C', need: 10, equipped: 5, stored: 0, reserve: 0, target: 10, have: 5, deficit: 5,
  });
});
test('higher-tier satisfied gear remains allocated without lower-tier demand', () => {
  const need = demand(['SteelSword_C', null]);
  assert.deepEqual(need, { SteelSword_C: 1, Sword_C: 1 });
  assert.equal(gearCoverage('SteelSword_C', need.SteelSword_C, 1, 0, 0).deficit, 0);
  assert.equal(gearCoverage('Sword_C', need.Sword_C, 0, 0, 0).deficit, 1);
});
test('manual override remains covered and does not demand a preset replacement', () => {
  assert.deepEqual(demand(['QuestAxe_C']), {});
  assert.equal(gearCoverage('QuestAxe_C', 0, 1, 0, 0).deficit, 0);
});
test('lower ranked worn gear requires the upgraded pick', () => {
  assert.equal(desiredGearItem(preset, 'weapon', 'Sword_C', 'SteelSword_C'), 'SteelSword_C');
});
test('built-in assignments count full demand; stock covers deficits and reserve adds demand', () => {
  assert.equal(gearCoverage('Sword_C', 10, 5, 3, 2).deficit, 4);
  assert.equal(gearCoverage('Sword_C', 10, 5, 5, 0).deficit, 0);
  assert.equal(gearCoverage('Sword_C', 0, 5, 0, 2).deficit, 2);
});
