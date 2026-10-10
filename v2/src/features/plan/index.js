import { RESOURCES, xy, extractors, storages, stock, proposal, distance, frame } from './model.js';

const hasTool = (ctx, name) => ctx.tools.some(t => (typeof t === 'string' ? t : t.name) === name);
const queryArgs = (ctx, name, query) => {
  const tool = ctx.tools.find(t => t.name === name);
  return tool?.inputSchema?.properties?.limit ? query : {};
};
const coordinates = p => p ? p.map(n => n.toFixed(0)).join(', ') + ' engine units (XY)' : 'Coordinates unavailable';
function plot(ui, markers) {
  const scale = frame(markers.map(m => m.position));
  if (!scale) return ui.empty('Coordinate plot unavailable', 'Saved building coordinates remain unknown.');
  const ns = 'http://www.w3.org/2000/svg';
  const svg = document.createElementNS(ns, 'svg');
  svg.setAttribute('viewBox', '0 0 500 500'); svg.setAttribute('role', 'img');
  svg.setAttribute('aria-label', 'Shared XY coordinate plot. Full labeled coordinates appear in the list below. No terrain or route evidence.');
  svg.style.cssText = 'display:block;width:100%;max-width:600px;margin:auto;background:#F7F5EE;border:1px solid #D7DCD4';
  const append = (tag, attrs, text) => { const n = document.createElementNS(ns, tag); for (const [k,v] of Object.entries(attrs)) n.setAttribute(k, v); if (text) n.textContent = text; svg.append(n); return n; };
  for (let i = 50; i <= 450; i += 100) {
    append('line', {x1:i,y1:50,x2:i,y2:450,stroke:'#D7DCD4'});
    append('line', {x1:50,y1:i,x2:450,y2:i,stroke:'#D7DCD4'});
  }
  append('text', {x:50,y:30,fill:'#52625B','font-size':13}, 'Y ↑ · X → · engine units · equal XY scale');
  append('text', {x:50,y:480,fill:'#52625B','font-size':13}, 'Frame span ' + Math.round(scale.size) + ' engine units');
  for (const marker of markers.filter(m => xy(m.position))) {
    const [x,y] = scale.project(marker.position), color = marker.kind === 'proposal' ? '#7A5A1D' : '#1D4638';
    if (marker.kind === 'camp') append('polygon', {points:`${x},${y-9} ${x-9},${y+8} ${x+9},${y+8}`,fill:color});
    else if (marker.kind === 'proposal') append('polygon', {points:`${x},${y-10} ${x-10},${y} ${x},${y+10} ${x+10},${y}`,fill:color});
    else if (marker.kind === 'storage') append('rect', {x:x-7,y:y-7,width:14,height:14,fill:color});
    else append('circle', {cx:x,cy:y,r:4,fill:'#52625B'});
    // Resource labels are in the accessible fallback list; avoid unreadable overlapping plot text.
    if (marker.kind !== 'resource') append('text', {x:Math.min(x+12,380),y:y-12,fill:'#152D27','font-size':13}, marker.label);
  }
  return ui.el('figure', {}, svg, ui.el('figcaption', {}, '▲ Saved camp · ■ Saved storage · ◆ Local proposal · ● Static node. Equal scale, no terrain or walkable routes.'));
}

export async function render(ctx) {
  const u = ctx.ui, root = u.el('div', {className:'plan-feature'});
  root.append(u.heading('Where should we gather?', 'Compare saved extractor locations with a local position proposal. Resources and stock remain separate evidence.'));
  if (!hasTool(ctx, 'read_saved_world')) {
    root.append(u.empty('Saved-world tool unavailable', 'The connected MCP must expose read_saved_world to list saved extractors.', u.link('Check source health', 'workspace'))); return root;
  }
  const result = await ctx.loadTool('read_saved_world');
  if (!result.data) {
    root.append(u.empty('Saved evidence unavailable', result.error || 'No parsed save was returned.', u.button('Retry saved evidence', async () => { await ctx.loadTool('read_saved_world', {}, {refresh:true}); root.replaceWith(await render(ctx)); }))); return root;
  }
  const world = result.data, camps = extractors(world), stores = storages(world);
  const resource = RESOURCES.find(r => r.key === ctx.params.get('resource')) || RESOURCES[0];
  const selected = camps.find(c => c.id === ctx.params.get('building')) || camps[0];
  const proposed = proposal(ctx.params);
  const change = extra => ctx.navigate('plan', {resource:resource.key, ...(selected ? {building:selected.id}:{}), ...(proposed ? {proposalX:proposed[0],proposalY:proposed[1]}:{}), ...extra});
  if (result.status === 'stale') root.append(u.notice('Saved evidence retained after a failed refresh. It may be stale.', 'warning'));
  root.append(u.el('p', {className:'muted'}, `Player containers only · Snapshot ${world.snapshot_id || 'identifier unavailable'} · Ingested ${world.ingested_at || 'unknown'} · Game capture time ${world.meta?.capture_time || 'unknown'}`), u.source(result));
  const resourceSelect = u.el('select', {onChange:e => change({resource:e.target.value})}, RESOURCES.map(r => u.el('option', {value:r.key,selected:r.key===resource.key},r.label)));
  root.append(u.el('div', {className:'controls'}, u.field('Resource', resourceSelect, 'Berries and mushrooms are evaluated separately.')));
  if (!camps.length) {
    root.append(u.empty('No saved player Foraging Camps found', 'This first extractor model accepts only the validated ForagingCamp_C class with Player faction. Other extraction models are pending. Missing containers do not prove absence or demolition.', u.link('Inspect saved supplies', 'supplies'))); return root;
  }
  const campSelect = u.el('select', {onChange:e => change({building:e.target.value,proposalX:'',proposalY:''})}, camps.map((c,i) => u.el('option', {value:c.id,selected:c.id===selected.id},`Foraging Camp ${i+1} · ${c.id}`)));
  root.append(u.el('div', {className:'panel'}, u.el('h2', {}, 'Saved extractor directory'), u.field('Extractor', campSelect, 'Stable saved container ID. Runtime actors are not linked.'), u.el('p', {}, camps.length + ' parsed player Foraging Camps. Empty or missing buildings may be omitted by the parser.')));
  root.append(u.el('nav', {className:'section-links','aria-label':'Extractor record sections'}, ['Location','Balance','Workforce','Observations'].map(s => u.button(s, () => { const target=root.querySelector('#plan-'+s.toLowerCase());target.scrollIntoView({behavior:'instant',block:'start'});target.focus({preventScroll:true}); }))));
  const current = xy(selected.position), location = u.el('section', {id:'plan-location',className:'panel',tabindex:'-1'}, u.el('h2', {}, 'Location evidence'));
  location.append(u.el('p', {}, `Selected camp ${selected.id} · ${coordinates(current)}`));
  const query = current ? {limit:100,center:{x:current[0],y:current[1]},radius:15000} : {limit:100};
  const requests = await Promise.allSettled([
    hasTool(ctx,'read_resource_baseline') ? ctx.loadTool('read_resource_baseline',queryArgs(ctx,'read_resource_baseline',query)) : Promise.resolve(null),
    hasTool(ctx,'read_foraging_candidates') ? ctx.loadTool('read_foraging_candidates',queryArgs(ctx,'read_foraging_candidates',{...query,limit:10})) : Promise.resolve(null)
  ]);
  const baseline = requests[0].status === 'fulfilled' ? requests[0].value : null;
  const candidateResult = requests[1].status === 'fulfilled' ? requests[1].value : null;
  const returnedNodes = Array.isArray(baseline?.data?.nodes) ? baseline.data.nodes : Array.isArray(baseline?.data?.items) ? baseline.data.items : [];
  const nodes = returnedNodes.filter(n => xy(n.position) && (!current || distance(current,xy(n.position))<=15000)).slice(0,100);
  const markers = [{kind:'camp',label:'Current camp',position:current}, ...(proposed ? [{kind:'proposal',label:'Local proposal',position:proposed}]:[]), ...stores.filter(s => xy(s.position)).map((s,i) => ({kind:'storage',label:`Storage ${i+1}`,position:xy(s.position)})), ...nodes.map(n => ({kind:'resource',label:`${n.baseline_category || 'Unknown category'} · ${n.id || 'node'}`,position:xy(n.position)}))];
  location.append(plot(u, markers));
  const fullList = u.el('details', {}, u.el('summary', {}, 'Coordinates and source list (plot alternative)'), u.table(['Location','XY coordinates','Evidence'], markers.filter(m=>m.position).map(m=>[m.label,coordinates(m.position),m.kind==='proposal'?'Local proposal':m.kind==='resource'?'Static baseline; availability unknown':'Saved Player container']), 'All plotted positions. Coordinates have unverified engine-unit interpretation.'));
  location.append(fullList);
  const pointField = (name, v) => u.el('input', {type:'number',name,required:true,min:-1e8,max:1e8,step:'any',value:v ?? ''});
  const xInput = pointField('proposal-x', proposed?.[0]), yInput = pointField('proposal-y', proposed?.[1]);
  const feedback = u.el('p', {role:'status',className:'muted'});
  const form = u.el('form', {className:'controls',onSubmit:e => {
    e.preventDefault(); const parsed = proposal(new URLSearchParams({proposalX:xInput.value,proposalY:yInput.value}));
    if (!parsed) {feedback.textContent='Enter finite X and Y coordinates between −100000000 and 100000000.';return;}
    change({proposalX:parsed[0],proposalY:parsed[1]});
  }},u.field('Proposed X (engine units)',xInput),u.field('Proposed Y (engine units)',yInput),u.el('button',{type:'submit'},proposed?'Update location proposal':'Compare a proposed position'));
  location.append(u.el('h3', {}, 'Propose a location'), u.el('p', {}, 'This creates a plan in the URL; no camp moves in game. Ground height, access and buildability are unknown.'), form,feedback);
  if (proposed) location.append(u.button('Remove proposal',()=>change({proposalX:'',proposalY:''})));
  const cards = [u.el('article',{className:'panel'},u.el('h3',{},'Current camp'),u.el('p',{},coordinates(current))),u.el('article',{className:'panel'},u.el('h3',{},'Local proposal'),u.el('p',{},coordinates(proposed)))];
  for (let i=0;i<cards.length;i++) {
    const point = i===0?current:proposed;
    const ds = stores.map(s=>distance(point,xy(s.position))).filter(n=>n!==null);
    cards[i].append(u.el('p',{},'Resource coverage: unavailable; harvesting radius, complete baseline and item yields are unverified.'),u.el('p',{},'Current availability: unknown'),u.el('p',{},ds.length?'Nearest listed storage: '+Math.min(...ds).toFixed(0)+' engine units, straight-line XY estimate; walkable route unknown.':'Storage distance unavailable.'),u.el('p',{},'Delivered output: not measured'));
  }
  location.append(u.el('div',{className:'comparison'},cards));
  location.append(u.notice(baseline?.data ? `Static evidence sample: ${nodes.length} plotted of ${returnedNodes.length} returned records${typeof baseline.data.truncated==='number'?'; '+baseline.data.truncated+' matching records omitted by service':''}. Local display is limited to 100 nodes${current?' within an exploratory 15000 engine-unit radius of the saved camp':''}. ${Object.keys(queryArgs(ctx,'read_resource_baseline',query)).length?'The tool also supports a bounded query.':'The tool has no advertised bounded arguments; filtering is local.'} Static categories are shown together: the selected resource controls saved stock only, not inferred node yield. This is not a verified work radius, complete coverage, harvestable stock or resource-specific yield. Proposal surroundings may be outside this sample.` : 'Static resource baseline unavailable. Saved coordinates and proposals remain usable.'));
  if (baseline?.status==='stale') location.append(u.notice('Static evidence is cached after a failed refresh. Inspect its source before comparing.', 'warning'));
  location.append(u.button('Refresh location evidence',async e=>{
    e.currentTarget.disabled=true;
    await Promise.allSettled(['read_resource_baseline','read_foraging_candidates'].filter(name=>hasTool(ctx,name)).map(name=>ctx.loadTool(name,queryArgs(ctx,name,{...query,limit:name==='read_foraging_candidates'?10:100}),{refresh:true})));
    const updated=await render(ctx);root.replaceWith(updated);updated.querySelector('#plan-location')?.focus();
  }));
  if (baseline) location.append(u.source(baseline));
  const rawCandidates = Array.isArray(candidateResult?.data) ? candidateResult.data : candidateResult?.data?.items || candidateResult?.data?.candidates;
  if (Array.isArray(rawCandidates) && rawCandidates.length) location.append(u.el('details',{},u.el('summary',{},'Unlinked runtime candidate evidence'),u.notice('Runtime candidates have no proved relationship to this saved camp. No joins by actor suffix, name or proximity are made. Showing at most 10 returned candidates.'),u.table(['Runtime actor','Candidate position','Status'],rawCandidates.slice(0,10).map(c=>[c.target_runtime_actor_id||'Unknown',coordinates(xy(c.practical_candidate?.position)),'Unlinked; radius/units/buildability assumptions unverified']), 'Runtime evidence sample; not recommendations for the selected saved building.'),u.source(candidateResult)));
  else location.append(u.el('p',{className:'muted'},'Runtime candidate evidence unavailable or empty. No automatic saved-building recommendation.'));
  root.append(location);
  const qty = stock(selected,resource.key);
  const balance = u.el('section',{id:'plan-balance',className:'panel',tabindex:'-1'},u.el('h2',{},resource.label+' balance'),u.el('p',{},`Current saved stock: ${qty===null?'unknown':qty+' units'}`),u.notice('One current inventory is not an observation window. Opening/closing stocks and net stock change require comparable saved observations. Extraction and individual transfers remain unmeasured.'),u.el('details',{},u.el('summary',{},'Full building balance and missing measurements'),u.el('p',{},'Opening + extracted + transfers in − transfers out − consumption − losses = closing.'),u.table(['Term','Evidence'],['Opening stock','Extracted','Transfers in','Transfers out','Consumption','Losses','Closing stock','Reconciliation residual'].map(t=>[t,'Not measured for a complete work window']), 'Residual is calculated only when every balance term is known.')),u.link('Review automatic observations and saved snapshots','journal',{building:selected.id}),u.source(result));
  root.append(balance,u.el('section',{id:'plan-workforce',className:'panel',tabindex:'-1'},u.el('h2',{},'Workforce evidence'),u.el('p',{},'Saved skills, equipment and condition can inform a staffing proposal. Current assignment and time spent working are unknown.'),u.link('Review worker candidates','workforce',{building:selected.id,resource:resource.key})),u.el('section',{id:'plan-observations',className:'panel',tabindex:'-1'},u.el('h2',{},'Automatic observations'),u.notice('Production capture requires attributed deliveries and worker game-time counters. Current service support is evaluated in Journal; no manual measurements are entered.'),u.link('Open automatic MCP readiness','journal',{building:selected.id})));
  return root;
}
