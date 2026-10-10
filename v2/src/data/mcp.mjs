import { spawn } from 'node:child_process';
import { TOOL_NAMES } from './demo.mjs';
export const ALLOWLIST = new Set([...TOOL_NAMES,'get_mcp_capabilities','read_connection_health']);
export function validateArguments(value,schema={type:'object',properties:{},additionalProperties:false}) {
 if(schema.type==='object'){if(!value||typeof value!=='object'||Array.isArray(value))return false;const properties=schema.properties??{};if(schema.additionalProperties===false&&Object.keys(value).some(k=>!Object.hasOwn(properties,k)))return false;if((schema.required??[]).some(k=>!Object.hasOwn(value,k)))return false;return Object.entries(value).every(([k,v])=>properties[k]?validateArguments(v,properties[k]):false);}
 if(schema.type==='number'||schema.type==='integer')return typeof value==='number'&&Number.isFinite(value)&&(schema.type!=='integer'||Number.isInteger(value))&&(schema.minimum===undefined||value>=schema.minimum)&&(schema.maximum===undefined||value<=schema.maximum)&&(schema.exclusiveMinimum===undefined||value>schema.exclusiveMinimum);
 if(schema.type==='string')return typeof value==='string'&&(schema.maxLength===undefined||value.length<=schema.maxLength);
 if(schema.type==='boolean')return typeof value==='boolean';
 return false;
}
export function unwrap(result) {
 if(result?.isError) throw new Error('MCP reported a tool error');
 let output=result?.structuredContent;
 if(output===undefined) { const block=result?.content?.find(c=>c.type==='text'); if(block) {try{output=JSON.parse(block.text);}catch{output={note:block.text};}} }
 output??=result;
 const data=output && Object.hasOwn(output,'data')?output.data:output;
 const stale=data?.status?.stale===true || data?.stale===true;
 return {status:stale?'stale':'ok',data,source:output?.source??'MCP',note:output?.note??'',retrievedAt:output?.retrieved_at??new Date().toISOString()};
}
export class StdioMcp {
 constructor(command,args=[],{timeout=15000,maxBytes=32*1024*1024}={}) {this.command=command;this.args=args;this.timeout=timeout;this.maxBytes=maxBytes;this.pending=new Map();this.nextId=1;this.buffer='';}
 async connect(){
  if(this.ready) return this.ready;
  this.ready=(async()=>{
   if(!this.command)throw new Error('MCP command is not configured');
   this.child=spawn(this.command,this.args,{stdio:['pipe','pipe','pipe'],shell:false,windowsHide:true});
   this.child.stdout.setEncoding('utf8');this.child.stderr.resume();
   this.child.stdout.on('data',chunk=>{this.buffer+=chunk;if(Buffer.byteLength(this.buffer)>this.maxBytes){this.fail(new Error('MCP response exceeded size limit'));this.child.kill();return;}let newline;while((newline=this.buffer.indexOf('\n'))>=0){const line=this.buffer.slice(0,newline);this.buffer=this.buffer.slice(newline+1);if(!line.trim())continue;try{const response=JSON.parse(line);const p=this.pending.get(response.id);if(p){clearTimeout(p.timer);this.pending.delete(response.id);response.error?p.reject(new Error('MCP request failed')):p.resolve(response.result);}}catch{this.fail(new Error('Invalid MCP JSON response'));this.child.kill();}}});
   this.child.on('error',()=>this.fail(new Error('MCP process could not start')));this.child.on('exit',()=>this.fail(new Error('MCP process disconnected')));
   await this.request('initialize',{protocolVersion:'2024-11-05',capabilities:{},clientInfo:{name:'steward-v2-readonly',version:'2.0'}});
   this.child.stdin.write(JSON.stringify({jsonrpc:'2.0',method:'notifications/initialized'})+'\n');
  })();
  try{await this.ready;}catch(error){this.ready=null;throw error;}
 }
 fail(error){for(const p of this.pending.values()){clearTimeout(p.timer);p.reject(error);}this.pending.clear();this.ready=null;}
 request(method,params){if(this.pending.size>=8)return Promise.reject(new Error('MCP request capacity reached'));return new Promise((resolve,reject)=>{const id=this.nextId++;const timer=setTimeout(()=>{this.pending.delete(id);reject(new Error('MCP request timed out'));},this.timeout);this.pending.set(id,{resolve,reject,timer});this.child.stdin.write(JSON.stringify({jsonrpc:'2.0',id,method,params})+'\n',error=>{if(error){clearTimeout(timer);this.pending.delete(id);reject(new Error('MCP write failed'));}});});}
 async list(){await this.connect();const result=await this.request('tools/list',{});return (result?.tools??[]).filter(t=>ALLOWLIST.has(t.name));}
 async call(name,args){if(!ALLOWLIST.has(name))throw new Error('Tool is not allowed');await this.connect();return unwrap(await this.request('tools/call',{name,arguments:args}));}
 close(){this.child?.kill();this.fail(new Error('MCP closed'));}
}
