import {hasTool,observationReadiness} from '../../data/capabilities.js';
export async function render(context) {
 const {ui,tools,loadTool}=context;const {el,heading,notice,source,button,empty,table}=ui;
 const root=el('div',{className:'workspace-feature'},heading('Source health','Inspect saved, runtime and packaged evidence separately.'));
 const status=el('div',{'aria-live':'polite'});root.append(button('Retry discovery and sources',async()=>{await context.refreshTools();context.navigate('workspace',{refresh:String(Date.now())});}),status);
 if(!tools.length){status.append(empty('MCP unavailable','No read tools have been discovered. Configure the server bridge or retry discovery.'));return root;}
 root.append(el('section',{className:'panel'},el('h2',{},'Read-only capabilities'),table(['Tool','Availability'],tools.map(t=>[typeof t==='string'?t:t.name,'Discovered']),'Discovery does not prove source freshness.')));
 const readiness=observationReadiness(tools);root.append(el('section',{className:'panel'},el('h2',{},'Automatic capture'),notice('Unsupported'),el('p',{},readiness.reason)));
 const entries=[['Saved snapshot','read_saved_world'],['Runtime actors','read_live_map'],['Packaged resource baseline','read_resource_baseline'],['Companion telemetry','read_companion_telemetry']];
 await Promise.all(entries.map(async([title,name])=>{const panel=el('section',{className:'panel'},el('h2',{},title));root.append(panel);if(!hasTool(tools,name)){panel.append(notice('Tool not discovered'));return;}const result=await loadTool(name,{}, {refresh:context.params.has('refresh')});if(result.status==='error'){panel.append(notice(result.error||'Source unavailable','error'));return;}const data=result.data;let summary;
 if(name==='read_saved_world')summary=`Snapshot ${ui.value(data?.snapshot_id)}. Ingested at ${ui.value(data?.ingested_at)}. Simulation playtime ${Number.isFinite(data?.meta?.playtimeSeconds)?(data.meta.playtimeSeconds/3600).toFixed(2)+' game hours':'unknown'}.`;
 else if(name==='read_live_map')summary=`${result.status==='stale'||data?.status?.stale?'Stale runtime evidence':'Runtime response received'}. Source capture ${ui.value(data?.status?.captured_at)}. Actor ownership remains unknown.`;
 else if(name==='read_resource_baseline')summary=`${Array.isArray(data?.nodes)?data.nodes.length:'Unknown'} packaged nodes. Source kind: ${ui.value(data?.source_kind)}. Current availability, terrain and access are unknown.`;
 else summary=data?.data==null?'No companion telemetry received. This does not invalidate independently available runtime-map evidence.':'Telemetry response received; inspect source limitations.';
 panel.append(el('p',{},summary));if(result.status==='stale')panel.append(notice('Cached or stale evidence. '+(result.error||'')));panel.append(source(result));}));
 root.append(el('section',{className:'panel'},el('h2',{},'Service attribution'),el('p',{},'This frontend is new Steward v2 code. Existing save parsing and MCP services retain their upstream attribution. Packaged game assets remain owned by their respective authors.')));return root;
}
