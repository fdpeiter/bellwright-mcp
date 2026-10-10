// Synthetic stdio smoke: no game, private saves or extracted assets required.
import assert from 'node:assert/strict';
import http from 'node:http';
import {readFile} from 'node:fs/promises';
import {spawn} from 'node:child_process';
import {fileURLToPath} from 'node:url';
const fixtures = JSON.parse(await readFile(new URL('./testdata/capabilities.json', import.meta.url), 'utf8'));
let failBackend = false;
const backend = http.createServer((req,res) => {
 const path = new URL(req.url,'http://localhost').pathname;
 if(req.method !== 'GET' || !Object.hasOwn(fixtures,path)) {res.writeHead(404).end();return}
 res.writeHead(failBackend?503:200,{'content-type':'application/json'});
 res.end(JSON.stringify(fixtures[path]));
});
await new Promise((resolve,reject)=>{backend.once('error',reject);backend.listen(0,'127.0.0.1',resolve)});
const origin='http://127.0.0.1:'+backend.address().port;
const binary=process.argv[2] || fileURLToPath(new URL(process.platform==='win32'?'./bellwright-mcp.exe':'./bellwright-mcp',import.meta.url));
const child=spawn(binary,['--companion-url',origin,'--live-url',origin],{stdio:['pipe','pipe','inherit']});
const pending=new Map();let buffer='',nextID=0;
function rejectPending(error){for(const p of pending.values())p.reject(error);pending.clear()}
child.on('error',rejectPending);child.stdin.on('error',rejectPending);
const exited=new Promise(resolve=>child.on('close',(code,signal)=>{rejectPending(Error('MCP exit '+code+' '+signal));resolve(code)}));
child.stdout.on('data',chunk=>{
 buffer+=chunk;let end;
 while((end=buffer.indexOf('\n'))>=0){
  const line=buffer.slice(0,end);buffer=buffer.slice(end+1);
  if(!line.trim())continue;
  try{const msg=JSON.parse(line);const p=pending.get(msg.id);if(p){pending.delete(msg.id);p.resolve(msg)}}
  catch(error){rejectPending(error)}
 }
});
function call(method,params){return new Promise((resolve,reject)=>{const id=++nextID;pending.set(id,{resolve,reject});child.stdin.write(JSON.stringify({jsonrpc:'2.0',id,method,params})+'\n')})}
const deadline=setTimeout(()=>{rejectPending(Error('MCP smoke timed out'));child.kill()},45000);
const args={
 read_saved_world:{},list_save_snapshots:{},read_save_trends:{},read_companion_telemetry:{},
 read_live_map:{},read_resource_baseline:{limit:1},read_foraging_candidates:{limit:1},
 read_gear_plan:{limit:1},read_item_history:{item_id:'Iron_C',limit:2},
 search_saved_villagers:{guid:'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',limit:1},
 locate_saved_inventory:{item:'Iron_C',limit:1},inspect_saved_squads:{limit:1,member_limit:2},
 read_saved_village_progress:{limit:1},read_connection_health:{},
 rank_saved_workers:{skill:'craft',limit:1},audit_saved_job_priorities:{limit:1},
 read_saved_housing:{limit:1},compare_saved_stock_targets:{targets:{Iron_C:5},scope:'stored',limit:1},
 search_observed_items:{query:'iron',limit:1},summarize_resource_coverage:{limit:1}
};
try{
 const init=await call('initialize',{protocolVersion:'2025-06-18',capabilities:{},clientInfo:{name:'capability-smoke',version:'1'}});
 assert.ok(init.result,JSON.stringify(init));
 child.stdin.write(JSON.stringify({jsonrpc:'2.0',method:'notifications/initialized'})+'\n');
 const list=await call('tools/list',{});assert.equal(list.result.tools.length,Object.keys(args).length);
 for(const tool of list.result.tools){
  assert.ok(tool.inputSchema&&tool.outputSchema&&tool.annotations.readOnlyHint);
  assert.ok(Object.hasOwn(args,tool.name));
  const response=await call('tools/call',{name:tool.name,arguments:args[tool.name]});
  assert.ok(response.result&&!response.result.isError,JSON.stringify(response));
  const out=response.result.structuredContent;assert.ok(out);
  assert.ok(!JSON.stringify(out).includes('private-path-must-not-leak'));
  if(tool.name==='read_item_history'){assert.equal(out.points[0].observed_qty,null);assert.equal(out.points[0].observed_playtime,null);assert.equal(out.points[1].observed_qty,3)}
  if(tool.name==='inspect_saved_squads'){assert.equal(out.groups[0].members[1].resolution,'unmatched');assert.equal(out.groups[0].members[1].equipment,null)}
  if(tool.name==='read_live_map')assert.equal(out.data.status.stale,true);
  if(tool.name==='read_connection_health'){assert.equal(out.sources[1].availability,'no_data');assert.equal(out.sources[2].freshness,'stale')}
  if(tool.name==='rank_saved_workers'){assert.equal(out.items[0].score,4);assert.equal(out.items[0].rank,1)}
  if(tool.name==='read_saved_housing'){assert.equal(out.capacity_deficit,1);assert.equal(out.unknown_housing,1)}
  if(tool.name==='compare_saved_stock_targets'){assert.equal(out.items[0].quantity,3);assert.equal(out.items[0].deficit,2)}
  if(tool.name==='search_observed_items'){assert.equal(out.items[0].item_class,'Iron_C');assert.equal(out.items[0].official_label,null)}
  if(tool.name==='summarize_resource_coverage'){assert.equal(out.total_nodes,3);assert.equal(out.categories[0].nodes,2);assert.equal(out.categories[0].bounds.min_z,-1);assert.equal(out.truncated,1)}
  console.log(tool.name+': stdio OK');
 }
 for(const name of ['read_resource_baseline','read_foraging_candidates','read_gear_plan','read_item_history','search_saved_villagers','locate_saved_inventory','inspect_saved_squads','read_saved_village_progress','rank_saved_workers','audit_saved_job_priorities','read_saved_housing','compare_saved_stock_targets','search_observed_items','summarize_resource_coverage']){
  const bad=await call('tools/call',{name,arguments:{...args[name],limit:0}});
  assert.ok(bad.error||bad.result?.isError,name+': invalid limit accepted');
 }
 failBackend=true;
 const error=await call('tools/call',{name:'read_saved_world',arguments:{}});
 assert.equal(error.result?.isError,true);assert.equal(error.result?.structuredContent,undefined);
 const health=await call('tools/call',{name:'read_connection_health',arguments:{}});
 assert.ok(!health.result.isError);assert.ok(health.result.structuredContent.sources.every(s=>s.availability==='unavailable'&&s.freshness==='unknown'));
 child.stdin.end();assert.equal(await exited,0);
 console.log(Object.keys(args).length+' tools, invalid limits, stale/null evidence, independent outages and EOF: PASS');
}catch(error){console.error(error);process.exitCode=1;child.kill()}
finally{clearTimeout(deadline);child.stdin.destroy();backend.closeAllConnections();await new Promise(resolve=>backend.close(resolve))}
