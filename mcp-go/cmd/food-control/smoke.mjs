import {spawn} from 'node:child_process';
import readline from 'node:readline';
const [binary,log]=process.argv.slice(2);
if(!binary||!log)throw Error('Usage: node smoke.mjs <food-control binary> <private runtime log>');
const proc=spawn(binary,['-log',log],{stdio:['pipe','pipe','inherit']});
const pending=new Map();let seq=0;
readline.createInterface({input:proc.stdout}).on('line',line=>{const x=JSON.parse(line);if(x.id)pending.get(x.id)?.(x);});
const timeout=setTimeout(()=>{proc.kill();throw Error('MCP smoke timed out');},15000);
const call=(method,params)=>new Promise(resolve=>{const id=++seq;pending.set(id,resolve);proc.stdin.write(JSON.stringify({jsonrpc:'2.0',id,method,params})+'\n');});
try{
 const init=await call('initialize',{protocolVersion:'2025-03-26',capabilities:{},clientInfo:{name:'food-smoke',version:'1'}});if(init.error)throw Error(JSON.stringify(init.error));
 proc.stdin.write(JSON.stringify({jsonrpc:'2.0',method:'notifications/initialized'})+'\n');
 const list=await call('tools/list',{});const tool=list.result.tools.find(t=>t.name==='read_live_food_state');if(!tool?.outputSchema||!tool.annotations?.readOnlyHint)throw Error('Missing read tool schema/annotations');
 const r=await call('tools/call',{name:tool.name,arguments:{}});if(r.error||r.result.isError)throw Error(JSON.stringify(r));
 const s=r.result.structuredContent;if(!s||s.source!=='ue4ss_live'||s.write_available!==false)throw Error('Invalid live provenance or write gate');
 console.log(JSON.stringify({tool:tool.name,schema:true,source:s.source,status:s.status,available:s.available,stale:s.stale,age_seconds:s.age_seconds,write_available:s.write_available,buildings:s.data?.buildings?.length,settlements:s.data?.settlements?.length}));
 const storage=list.result.tools.find(t=>t.name==='read_live_storage_state');if(!storage?.outputSchema||!storage.annotations?.readOnlyHint)throw Error('Missing storage read schema');
 const sr=await call('tools/call',{name:storage.name,arguments:{}});if(sr.error||sr.result.isError)throw Error(JSON.stringify(sr));
 const st=sr.result.structuredContent;if(st.write_available!==false)throw Error('Storage writes must remain disabled');
 console.log(JSON.stringify({tool:storage.name,schema:true,status:st.status,available:st.available,stale:st.stale,containers:st.data?.storage?.containers?.length,write_available:st.write_available}));
}finally{clearTimeout(timeout);proc.stdin.end();proc.kill();}
