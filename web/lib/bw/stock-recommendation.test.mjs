import assert from 'node:assert/strict';
import test from 'node:test';
import { calculateStock, readStockAssumptions, storedQuantity } from './stock-recommendation.ts';
const assumption = { item: 'Iron_C', rate: '2.5', hours: '3', safety: '0.2', planned: '4' };
test('replenishment demand plus buffer plus planned demand rounds upward', () => {
  assert.equal(calculateStock(assumption).target, 13);
  assert.equal(calculateStock({ ...assumption, safety: '0', planned: '0' }).target, 8);
});
test('invalid assumptions cannot produce a target', () => {
  for (const row of [{ ...assumption, hours: '' }, { ...assumption, hours: '0' }, { ...assumption, safety: '2.1' }, { ...assumption, item: 'Not an item' }, { ...assumption, rate: '-1' }, { ...assumption, safety: 'Infinity' }, { ...assumption, planned: '0.5' }, { ...assumption, rate: '1e100' }]) assert.equal(calculateStock(row).target, null);
});
test('unknown stock stays unknown and omitted item is zero only in a valid aggregate', () => {
  assert.equal(storedQuantity({ storage: { totals: {} } }, 'Iron_C'), 0);
  assert.equal(storedQuantity({ storage: { totals: { Iron_C: 3 } } }, 'Iron_C'), 3);
  assert.equal(storedQuantity({ storage: { totals: { Iron_C: 3, Other_C: null } } }, 'Iron_C'), 3);
  assert.equal(storedQuantity({ storage: { totals: { Iron_C: 3, Other_C: null } } }, 'Absent_C'), null);
  for (const totals of [null, [], { Iron_C: null }, { Other_C: -1 }]) assert.equal(storedQuantity({ storage: { totals } }, 'Iron_C'), null);
});
test('malformed or excessive local assumptions reset safely and extra fields are discarded', () => {
  assert.equal(readStockAssumptions('{').length, 1);
  assert.equal(readStockAssumptions(JSON.stringify(Array(101).fill(assumption))).length, 1);
  assert.deepEqual(readStockAssumptions(JSON.stringify([{ ...assumption, target: 99 }])), [assumption]);
});


