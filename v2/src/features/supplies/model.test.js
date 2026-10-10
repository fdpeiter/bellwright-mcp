import test from 'node:test';
import assert from 'node:assert/strict';
import {playerContainers, quantity, resources, distribution, coordinates} from './model.js';
test('Player scope excludes hostile and missing faction regardless of all-world totals',()=>{
  const world={storage:{totals:{Berry_C:1001},containers:[{id:'own',faction:'Player',items:{Berry_C:1}},{id:'other',faction:'Bandit',items:{Berry_C:1000}},{id:'unknown',items:{Berry_C:800}}]}};
  const own=playerContainers(world);assert.deepEqual(own.map(c=>c.id),['own']);assert.equal(quantity(own[0],'Berry_C'),1);
  assert.equal(playerContainers({}),null);assert.deepEqual(playerContainers({storage:{containers:[]}}),[]);
});
test('zero, absent, null and invalid quantities stay distinct',()=>{
  const c={items:{zero:0,missing:null,negative:-1,invalid:'8',known:8}};
  assert.equal(quantity(c,'zero'),0);assert.equal(quantity(c,'absent'),null);assert.equal(quantity(c,'missing'),null);assert.equal(quantity(c,'negative'),null);assert.equal(quantity(c,'invalid'),null);assert.equal(quantity(c,'known'),8);
});
test('positive filter never treats unknown as zero or an all-world quantity',()=>{
  const cs=[{id:'a',name:'North',items:{berry:3}},{id:'b',name:'South',items:{berry:0}},{id:'c',name:'East',items:{}}];
  assert.equal(distribution(cs,'berry').length,3);assert.equal(distribution(cs,'berry','',true).length,1);assert.equal(distribution(cs,'berry','South').length,1);
  assert.deepEqual(resources(cs),['berry']);
});
test('saved coordinate evidence is never a walkable route',()=>{
  assert.equal(coordinates({position:[0,0,0]}),'0, 0, 0 engine units');assert.equal(coordinates({position:[1]}),'Unknown');
});
