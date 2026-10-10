import test from 'node:test';
import assert from 'node:assert/strict';
import {extractors,storages,stock,proposal,xy,distance,frame} from './model.js';
import {render} from './index.js';
import * as ui from '../../ui.js';

test('only validated player extractor classes enter the directory',()=>{
  const world={storage:{containers:[{id:'camp',cls:'ForagingCamp_C',faction:'Player'},{id:'hostile',cls:'ForagingCamp_C',faction:'Bandits'},{id:'unknown',cls:'ForagingCamp_C'},{id:'producer',cls:'SmallFarm_C',faction:'Player'},{id:'name-match',cls:'FancyForagingCamp_C',faction:'Player'},{id:'store',cls:'Stockpile_C',faction:'Player'}]}};
  assert.deepEqual(extractors(world).map(c=>c.id),['camp']);
  assert.deepEqual(storages(world).map(c=>c.id),['store']);
});
test('missing resource records remain unknown while known zero survives',()=>{
  assert.equal(stock({items:{Berry_C:0}},'Berry_C'),0);
  for(const c of [{},{items:{}},{items:{Berry_C:-1}},{items:{Berry_C:'20'}},{items:{Berry_C:NaN}}])assert.equal(stock(c,'Berry_C'),null);
});
test('proposal coordinates are explicit finite values within engine bounds',()=>{
  assert.equal(proposal(new URLSearchParams()),null);
  for(const value of ['', 'Infinity','NaN','100000001'])assert.equal(proposal(new URLSearchParams({proposalX:value,proposalY:'0'})),null);
  assert.deepEqual(proposal(new URLSearchParams({proposalX:'0',proposalY:'-120.5'})),[0,-120.5]);
  assert.equal(xy(['0',0]),null);
  assert.equal(distance(null,[0,0]),null);
});
test('comparison uses equal scale and shared coordinate frame',()=>{
  const f=frame([[0,0],[1000,0],[0,1000]]);
  const a=f.project([0,0]),b=f.project([1000,0]),c=f.project([0,1000]);
  assert.equal(b[0]-a[0],a[1]-c[1]);
  assert.equal(frame([null,[]]),null);
});

class Element {
  constructor(tag){this.tag=tag;this.children=[];this.attributes={};this.style={};}
  append(...children){this.children.push(...children);}
  setAttribute(k,v){this.attributes[k]=v;}
  addEventListener(){}
}
globalThis.Node=Element;
globalThis.document={createElement:tag=>new Element(tag),createElementNS:(_,tag)=>new Element(tag),createTextNode:text=>String(text)};
function texts(node){return typeof node==='string'?node:[node.textContent||'',...node.children.map(texts)].flat().join(' ');}
function context(tools,world){const calls=[];return {calls,ui,tools,params:new URLSearchParams(),navigate(){},async loadTool(name,args){calls.push({name,args});return {status:'ok',data:name==='read_saved_world'?world:name==='read_resource_baseline'?{nodes:[]}:[],source:'synthetic_test',retrievedAt:'test-time'};}};}
const world={snapshot_id:'synthetic',storage:{containers:[{id:'camp',cls:'ForagingCamp_C',faction:'Player',position:[0,0,0],items:{Berry_C:0}}]}};
test('runtime discovery controls bounds; no manual production or capture call',async()=>{
  const ctx=context([{name:'read_saved_world'},{name:'read_resource_baseline',inputSchema:{properties:{limit:{}}}},{name:'read_foraging_candidates',inputSchema:{properties:{limit:{}}}}],world);
  const page=await render(ctx);
  assert.deepEqual(ctx.calls[1].args,{limit:100,center:{x:0,y:0},radius:15000});
  assert.equal(ctx.calls[2].args.limit,10);
  assert.match(texts(page),/Current saved stock: 0 units/);
  assert.match(texts(page),/Consumption/);assert.match(texts(page),/Losses/);
  assert.match(texts(page),/Ground height, access and buildability are unknown/);
  assert.equal(ctx.calls.some(c=>/start|capture|assign/.test(c.name)),false);
});
test('legacy unbounded discovery calls no unsupported query arguments',async()=>{
  const ctx=context([{name:'read_saved_world'},{name:'read_resource_baseline'}],world);
  const page=await render(ctx);assert.deepEqual(ctx.calls[1].args,{});
  assert.match(texts(page),/filtering is local/);
});
test('no qualifying extractor produces a truthful empty state without resource calls',async()=>{
  const ctx=context([{name:'read_saved_world'},{name:'read_resource_baseline'}],{storage:{containers:[]}});
  const page=await render(ctx);assert.equal(ctx.calls.length,1);
  assert.match(texts(page),/Missing containers do not prove absence/);
});
test('bounded items preserve truncation and runtime candidates remain unlinked',async()=>{
  const ctx=context([{name:'read_saved_world'},{name:'read_resource_baseline',inputSchema:{properties:{limit:{}}}},{name:'read_foraging_candidates',inputSchema:{properties:{limit:{}}}}],world);
  ctx.loadTool=async name=>({status:'ok',source:'synthetic_test',data:name==='read_saved_world'?world:name==='read_resource_baseline'?{items:[{id:'static-berry',baseline_category:'Berry',position:[100,0,0]}],truncated:8}:{items:[{target_runtime_actor_id:'runtime-camp',practical_candidate:{position:[100,200,0]}}]}});
  const page=await render(ctx),text=texts(page);
  assert.match(text,/1 plotted of 1 returned records; 8 matching records omitted/);
  assert.match(text,/static-berry/);assert.match(text,/runtime-camp/);
  assert.match(text,/No joins by actor suffix, name or proximity/);
  assert.match(text,/Resource coverage: unavailable/);
});
