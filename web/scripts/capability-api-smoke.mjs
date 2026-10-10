// Exercise the production gear/history routes against a temporary synthetic DB.
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import {mkdtemp, rm, readFile} from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import net from 'node:net';
import {spawn} from 'node:child_process';
import {setTimeout as pause} from 'node:timers/promises';
import {fileURLToPath} from 'node:url';

const root = fileURLToPath(new URL('../..',import.meta.url));
const req = createRequire(path.join(root,'web/package.json'));
const Database = req('better-sqlite3');
const {drizzle} = req('drizzle-orm/better-sqlite3');
const {migrate} = req('drizzle-orm/better-sqlite3/migrator');
const fixtures = JSON.parse(await readFile(path.join(root,'mcp-go/testdata/capabilities.json'),'utf8'));
const dir = await mkdtemp(path.join(os.tmpdir(),'steward-capability-'));
const file = path.join(dir,'synthetic.db');
let child;
try {
 const db = new Database(file);
 migrate(drizzle(db),{migrationsFolder:path.join(root,'web/drizzle')});
 const insert = db.prepare('INSERT INTO snapshots (id,ingested_at,save_name,region,playtime_seconds,npc_count,mine_count,world) VALUES (?,?,?,?,?,?,?,?)');
 for(const id of [6,7]) {
  const world = structuredClone(fixtures['/api/world']);
  world.meta = {character:'synthetic-character',map:'synthetic-map'};
  world.storage.totals = id === 6 ? {} : {Iron_C:3};
  world.gear_presets = [{key:'kit',name:'Kit',slots:{weapon:[{item:'Sword_C',rank:1}]}}];
  world.npcs = Array.from({length:10},(_,i)=>({guid:'npc-'+i,is_player_npc:true,first_name:'Synthetic',gear_preset:'kit',equipment:i<5?{weapon:'Sword_C'}:{}}));
  const at = id === 6 ? '2026-10-10T11:00:00Z' : '2026-10-10T12:00:00Z';
  world.ingested_at = at;
  insert.run(id,at,'synthetic-save','synthetic-region',id===6?null:60,10,10,JSON.stringify(world));
 }
 db.close();
 const reservation = net.createServer();
 await new Promise(resolve=>reservation.listen(0,'127.0.0.1',resolve));
 const port = reservation.address().port;
 await new Promise(resolve=>reservation.close(resolve));
 child = spawn(process.execPath,[path.join(root,'web/node_modules/next/dist/bin/next'),'start','web','--hostname','127.0.0.1','--port',String(port)],{cwd:root,env:{...process.env,DATA_DIR:dir,DATABASE_URL:'file:'+file,MIGRATIONS_DIR:path.join(root,'web/drizzle'),NEXT_TELEMETRY_DISABLED:'1'},stdio:['ignore','pipe','pipe']});
 let logs='';child.stdout.on('data',c=>logs+=c);child.stderr.on('data',c=>logs+=c);
 const origin='http://127.0.0.1:'+port;
 let gear;
 for(let i=0;i<100;i++) {
  try { const res=await fetch(origin+'/api/gear');if(res.ok){gear=await res.json();break} }
  catch { /* production server startup */ }
  await pause(100);
 }
 assert.ok(gear,'Production gear route did not respond: '+logs);
 assert.equal(gear.snapshot_id,7);assert.equal(gear.ingested_at,'2026-10-10T12:00:00Z');
 assert.equal(gear.craft.find(i=>i.item==='Sword_C')?.deficit,5);
 const res=await fetch(origin+'/api/trends/item?name=Iron_C');assert.equal(res.status,200);
 const history=await res.json();assert.equal(history.points[0].qty,0);
 assert.equal(history.points[0].observed_qty,null);assert.equal(history.points[0].observed_playtime,null);
 assert.equal(history.points[1].observed_qty,3);assert.equal(history.points[1].snapshot_id,7);
 assert.equal(history.points[1].character,'synthetic-character');assert.equal(history.points[1].map,'synthetic-map');
 assert.equal(history.total_snapshots,2);assert.equal(history.sampled_points,2);assert.equal(history.sampled,false);
 console.log('Production API gear shortage, snapshot provenance, nullable history, lineage and sampling: PASS');
} finally {
 if(child&&child.exitCode===null){child.kill();await new Promise(resolve=>child.once('exit',resolve))}
 // Only the exact mkdtemp-created synthetic directory is removed.
 if(path.dirname(path.resolve(dir))!==path.resolve(os.tmpdir()) || !path.basename(dir).startsWith('steward-capability-')) throw Error('Unexpected synthetic directory');
 await rm(dir,{recursive:true,force:true});
}
