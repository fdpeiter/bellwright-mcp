import {spawn} from 'node:child_process';
const docker=process.env.DOCKER_BIN || 'docker';
const proc=spawn(docker,['compose','run','--rm','--no-deps','-T','mcp'],{stdio:['pipe','pipe','inherit']});
let buf='',seq=0;const pending=new Map();
proc.stdout.on('data',chunk=>{buf+=chunk;let end;while((end=buf.indexOf('\n'))>=0){const line=buf.slice(0,end);buf=buf.slice(end+1);if(!line.trim())continue;const msg=JSON.parse(line);pending.get(msg.id)?.(msg);pending.delete(msg.id)}});
const call=(method,params)=>new Promise(resolve=>{const id=++seq;pending.set(id,resolve);proc.stdin.write(JSON.stringify({jsonrpc:'2.0',id,method,params})+'\n')});
const timeout=setTimeout(()=>{console.error('MCP smoke test timed out');proc.kill();process.exitCode=1},45000);
try{
 const init=await call('initialize',{protocolVersion:'2025-06-18',capabilities:{},clientInfo:{name:'companion-smoke',version:'1'}});if(init.error)throw Error(JSON.stringify(init.error));
 proc.stdin.write(JSON.stringify({jsonrpc:'2.0',method:'notifications/initialized'})+'\n');
 const list=await call('tools/list',{});console.log('Tools:',list.result.tools.length);
 const toolArgs={rank_saved_workers:{skill:"craft"},compare_saved_stock_targets:{targets:{Iron_C:5},scope:"stored"},read_item_history:{item_id:process.env.SMOKE_ITEM_ID || 'Iron_C'},locate_saved_inventory:{item:process.env.SMOKE_ITEM_ID || 'Iron_C'}};
 for(const tool of list.result.tools){const res=await call('tools/call',{name:tool.name,arguments:toolArgs[tool.name] || {}});if(res.error||res.result?.isError)throw Error(tool.name+': '+JSON.stringify(res));console.log(tool.name+': OK');if(tool.name==='read_live_map'){const result=res.result.structuredContent || JSON.parse(res.result.content[0].text);console.log('Live freshness:',JSON.stringify(result.data.status));}}
}catch(e){console.error(e);process.exitCode=1}finally{clearTimeout(timeout);proc.stdin.end()}
