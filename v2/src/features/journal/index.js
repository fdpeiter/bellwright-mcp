import {hasTool,observationReadiness,stockComparison} from '../../data/capabilities.js';
export async function render(context) {
 const {ui,loadTool,tools}=context;const {el,heading,notice,table,source,link,empty,button,field}=ui;
 const root=el('div',{className:'journal-feature'},heading('Journal','Saved evidence and automatic measurement readiness.'));
 const readiness=observationReadiness(tools);
 root.append(el('section',{className:'panel',id:'observations'},el('h2',{},'Automatic extractor observations'),notice('Production capture unavailable'),el('p',{},readiness.reason),el('ul',{},readiness.missing.map(text=>el('li',{},text))),link('Inspect extractor evidence','plan',Object.fromEntries(context.params)),el('p',{className:'muted'},'MCP must record measured quantities and worker time automatically. There is no manual measurement form.')));
 const holder=el('section',{className:'panel'},el('h2',{},'Saved observations'));root.append(holder);
 if(!hasTool(tools,'list_save_snapshots')){holder.append(empty('Snapshot history unavailable','The MCP tool is not currently discovered.',link('Check source health','workspace')));return root;}
 async function refresh(forceRefresh=false){holder.setAttribute('aria-busy','true');const result=await loadTool('list_save_snapshots',{}, {refresh:forceRefresh});holder.replaceChildren(el('h2',{},'Saved observations'),button('Refresh saved evidence',()=>refresh(true)));
  if(result.status==='error'){holder.append(notice(result.error||'History unavailable','error'));holder.removeAttribute('aria-busy');return;}
  if(result.status==='stale')holder.append(notice('Showing cached snapshots. Refresh failed: '+(result.error||'source stale')));
  const snapshots=result.data?.snapshots??[];
  holder.append(snapshots.length?table(['Snapshot','Ingested at (wall clock)','Simulation playtime','Region'],snapshots.map(s=>[ui.value(s.id),ui.value(s.ingested_at),Number.isFinite(s.playtime_seconds)?`${(s.playtime_seconds/3600).toFixed(2)} game hours`:'Unknown',ui.value(s.region)]),'Save metadata; ingestion time is not a game capture time'):empty('No saved observations','Wait for the existing service to ingest a save.'));
  holder.append(source(result));holder.removeAttribute('aria-busy');
 }
 await refresh();
 const stock=el('section',{className:'panel',id:'stock'},el('h2',{},'Stock comparisons'));root.append(stock);
 if(!readiness.canCompareStock){stock.append(empty('Stock comparisons unavailable','The save trend tool is not discovered.',link('Check source health','workspace')));return root;}
 const result=await loadTool('read_save_trends');
 if(result.status==='error'){stock.append(notice(result.error||'Stock history unavailable','error'));return root;}
 if(result.status==='stale')stock.append(notice('Cached stock evidence; refresh the workspace to retry.'));
 stock.append(notice('History totals may span all world containers. This is not player-settlement output or building production.'),el('p',{},'Transfers, extraction, consumption and losses are not separated. No production rate or efficiency can be computed.'));
 const points=result.data?.points??[];
 const items=[...new Set(points.flatMap(p=>Object.keys(p.items??{})))].sort();
 if(!items.length){stock.append(empty('No stock history','No resource quantities are available in the saved observations.'),source(result));return root;}
 const select=el('select',{},items.map(item=>el('option',{value:item},item.replace(/_C$/,''))));const detail=el('div',{'aria-live':'polite'});
 const update=()=>{const comparison=stockComparison(points,select.value);detail.replaceChildren();if(comparison.status!=='ok'){detail.append(notice(comparison.reason));return;}detail.append(table(['Opening stock','Closing stock','Stock change','Simulation interval'],[[comparison.opening,comparison.closing,comparison.change>0?'+'+comparison.change:comparison.change,(comparison.gameSeconds/3600).toFixed(2)+' game hours']],'First and last returned saved observations'),el('p',{className:'muted'},`Endpoints: snapshot ${ui.value(comparison.first.snapshot_id)} → ${ui.value(comparison.last.snapshot_id)}. Quantities are resource units; inventory change is not production.`));};
 select.addEventListener('change',update);stock.append(field('Resource',select),detail,source(result));update();return root;
}
