import test from 'node:test';
import assert from 'node:assert/strict';
import {playerPeople, evidence, shortlist, extractionBuildings, proposalRecord} from './model.js';
test('only explicit recruited Player people can enter a staffing proposal',()=>{
  const world={npcs:[{guid:'player',is_player_npc:true},{guid:'runtime',faction:'Player'},{guid:'other',is_player_npc:false}]};
  assert.deepEqual(playerPeople(world).map(p=>p.guid),['player']);
  assert.equal(playerPeople({}),null);
  assert.equal(proposalRecord(world.npcs[0],{id:'camp',faction:null},'harvest',1),null);
});
test('missing values never become zero, present zero stays known',()=>{
  const v=evidence({skills:{harvest:{level:0}},morale:0,job_priorities:{harvesting:0},injuries:[]},'harvest');
  assert.equal(v.level,0);assert.equal(v.cap,null);assert.equal(v.morale,0);assert.equal(v.priority,0);assert.deepEqual(v.injuries,[]);
  assert.equal(evidence({},'harvest').injuries,null);assert.equal(evidence({},'harvest').tools,null);
});
test('skill evidence maps harvesting job explicitly and never guesses labour/strength priority',()=>{
  const person={job_priorities:{harvest:9,harvesting:2,labour:3,strength:4}};
  assert.equal(evidence(person,'harvest').priority,2);
  assert.equal(evidence({job_priorities:{harvest:9}},'harvest').priority,null);
  assert.equal(evidence(person,'labour').priority,null);
  assert.equal(evidence(person,'strength').priority,null);
});
test('extraction shortlist sorts unknown last without assignment claims',()=>{
  const people=[{first_name:'Unknown'},{first_name:'Zero',skills:{harvest:{level:0}}},{first_name:'Known',skills:{harvest:{level:3}}}];
  assert.deepEqual(shortlist(people,'','harvest').map(p=>p.first_name),['Known','Zero','Unknown']);
  assert.equal(shortlist(people,'known','harvest').length,2);
});
test('extractors exclude producers, storage and unknown ownership; proposals are local and explicit',()=>{
  const containers=[{id:'camp',cls:'ForagingCamp_C',faction:'Player'},{id:'barn',cls:'Barn_C',faction:'Player'},{id:'unknown',cls:'ForagingCamp_C'},{id:'producer',cls:'Workshop_C',faction:'Player'}];
  const buildings=extractionBuildings({storage:{containers}});
  assert.deepEqual(buildings.map(b=>b.id),['camp']);
  assert.equal(proposalRecord({guid:'known',is_player_npc:true},buildings[0],'harvest',0).snapshot,0);
  assert.equal(proposalRecord({guid:'known',is_player_npc:true},buildings[0],'harvest',0).gameWrite,false);
  assert.equal(proposalRecord({guid:'unknown'},buildings[0],'harvest',0),null);
});
