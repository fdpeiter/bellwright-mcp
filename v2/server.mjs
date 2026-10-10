import {createServer} from 'node:http';
import {readFile} from 'node:fs/promises';
import {resolve,extname,sep} from 'node:path';
import {fileURLToPath} from 'node:url';
import {StdioMcp,ALLOWLIST,validateArguments} from './src/data/mcp.mjs';
import {DEMO_TOOLS,demoCall} from './src/data/demo.mjs';
const root=fileURLToPath(new URL('./dist/',import.meta.url));
const demo=process.argv.includes('--demo');
let args=[];try{args=JSON.parse(process.env.STEWARD_MCP_ARGS||'[]');if(!Array.isArray(args)||args.some(a=>typeof a!=='string'))throw new Error();}catch{throw new Error('STEWARD_MCP_ARGS must be a JSON string array');}
const mcp=new StdioMcp(process.env.STEWARD_MCP_COMMAND,args);
const mime={'.html':'text/html; charset=utf-8','.js':'text/javascript; charset=utf-8','.mjs':'text/javascript; charset=utf-8','.css':'text/css; charset=utf-8','.svg':'image/svg+xml'};
const json=(res,status,data)=>{res.writeHead(status,{'Content-Type':'application/json','Cache-Control':'no-store'});res.end(JSON.stringify(data));};
export const server=createServer(async(req,res)=>{
 res.setHeader('X-Content-Type-Options','nosniff');res.setHeader('Content-Security-Policy',"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'");
 const host=req.headers.host;let url;try{url=new URL(req.url,'http://'+host);}catch{return json(res,400,{error:'Invalid request'});}
 if(!/^127\.0\.0\.1(?::\d+)?$/.test(host||''))return json(res,403,{error:'Loopback host required'});
 if(req.headers.origin&&req.headers.origin!==url.origin)return json(res,403,{error:'Same origin required'});
 try{
  if(url.pathname==='/api/tools'&&req.method==='GET'){const tools=demo?DEMO_TOOLS:await mcp.list();return json(res,200,{tools,demo});}
  if(url.pathname==='/api/tools/call'&&req.method==='POST'){
   if(!req.headers['content-type']?.startsWith('application/json'))return json(res,415,{error:'JSON required'});
   let body='';for await(const chunk of req){body+=chunk;if(Buffer.byteLength(body)>8192)return json(res,413,{error:'Request too large'});}
   let input;try{input=JSON.parse(body);}catch{return json(res,400,{error:'Invalid JSON'});}
   if(!ALLOWLIST.has(input.name)||typeof input.arguments!=='object'||input.arguments===null||Array.isArray(input.arguments)||Object.keys(input).some(k=>!['name','arguments'].includes(k)))return json(res,400,{error:'Invalid read-only tool request'});
   const discovered=demo?DEMO_TOOLS:await mcp.list();const tool=discovered.find(t=>t.name===input.name);
   if(!tool||!validateArguments(input.arguments,tool.inputSchema))return json(res,400,{error:'Tool arguments do not match discovered schema'});
   return json(res,200,demo?demoCall(input.name):await mcp.call(input.name,input.arguments));
  }
  if(req.method!=='GET'&&req.method!=='HEAD')return json(res,405,{error:'Method not supported'});
  const path=resolve(root,'.'+decodeURIComponent(url.pathname==='/'?'/index.html':url.pathname));
  if(!path.startsWith(root.endsWith(sep)?root:root+sep))return json(res,403,{error:'Invalid path'});
  const data=await readFile(path);res.writeHead(200,{'Content-Type':mime[extname(path)]||'application/octet-stream'});res.end(req.method==='HEAD'?undefined:data);
 }catch(error){json(res,url.pathname.startsWith('/api/')?503:404,{status:'error',data:null,source:null,note:'',retrievedAt:new Date().toISOString(),error:url.pathname.startsWith('/api/')?error.message:'File unavailable'});}
});
const port=Number(process.env.STEWARD_V2_PORT||18130);if(!Number.isInteger(port)||port<1024||port>65535)throw new Error('Invalid STEWARD_V2_PORT');
server.requestTimeout=20000;server.headersTimeout=10000;
server.listen(port,'127.0.0.1',()=>console.log(`Steward v2 at http://127.0.0.1:${port}${demo?' · SYNTHETIC DEMO':''}`));
for(const signal of ['SIGINT','SIGTERM'])process.on(signal,()=>{mcp.close();server.close();});
