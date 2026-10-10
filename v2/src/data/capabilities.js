export const hasTool = (tools,name) => tools.some(tool=>(typeof tool==='string'?tool:tool.name)===name);
export function observationReadiness(tools) {
 return {canCapture:false,canCompareStock:hasTool(tools,'read_save_trends'),reason:'Automatic extractor capture requires attributed deliveries, worker game-time and MCP capture records. These capabilities are not supported by this v2 contract.',missing:['Attributed deliveries per extractor/resource','Worker game-time per extractor','Automatic MCP capture lifecycle and recorded results']};
}
export function stockComparison(points,item) {
 if(!Array.isArray(points)||points.length<2)return {status:'missing',reason:'At least two saved observations are required.'};
 const first=points[0],last=points.at(-1),a=first.items?.[item],b=last.items?.[item];
 if(!Number.isFinite(a)||!Number.isFinite(b))return {status:'missing',reason:'This resource is missing from one endpoint. Missing is not zero.'};
 if(!Number.isFinite(first.playtime)||!Number.isFinite(last.playtime)||last.playtime<=first.playtime)return {status:'invalid',reason:'Simulation playtime does not advance across these observations.'};
 return {status:'ok',opening:a,closing:b,change:b-a,gameSeconds:last.playtime-first.playtime,first,last};
}
