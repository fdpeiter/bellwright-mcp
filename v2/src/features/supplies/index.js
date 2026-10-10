import { playerContainers, quantity, resources, label, containerName, coordinates, distribution } from './model.js';

export async function render(ctx) {
  const u=ctx.ui;
  const root=u.el('div',{className:'supplies-feature'},u.heading('Where is this resource stored?','Find saved stock in explicit Player containers, then inspect the storage evidence. Quantities are saved inventory, not production or delivery.'));
  async function refreshEvidence(button,params=ctx.params) {
    button.disabled=true;root.setAttribute('aria-busy','true');
    try {
      await ctx.loadTool('read_saved_world',{}, {refresh:true});
      const updated=await render({...ctx,params});root.replaceWith(updated);
      const heading=updated.querySelector('h1');heading?.setAttribute('tabindex','-1');heading?.focus();
    } finally {root.removeAttribute('aria-busy');button.disabled=false;}
  }
  root.append(u.el('style',{},'.supplies-feature .distribution-list{display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,260px),1fr));gap:16px}.supplies-feature .distribution-list article{margin:0}.supplies-feature .storage-title:focus{outline:3px solid var(--accent);outline-offset:4px}.supplies-feature .item-id{overflow-wrap:anywhere}.supplies-feature .check-field{flex-direction:row;align-items:center}.supplies-feature .check-field input{min-height:24px;width:24px;height:24px}'));
  if(!ctx.tools.some(t => (typeof t==='string'?t:t.name)==='read_saved_world')) {
    root.append(u.empty('Saved container evidence unavailable','The connected service does not expose read_saved_world.',u.link('Check source health','workspace')));return root;
  }
  const result=await ctx.loadTool('read_saved_world',{});
  if(!result?.data) {
    root.append(u.notice(result?.error || 'Saved evidence could not be loaded.','error'),u.button('Retry saved evidence',e=>refreshEvidence(e.currentTarget)),u.link('Check source health','workspace'));return root;
  }
  if(result.status==='stale') root.append(u.notice('Saved stock retained after a failed refresh. '+(result.error||''),'error'));
  const world=result.data;
  const containers=playerContainers(world);
  const itemIDs=resources(containers);
  let item=ctx.params.get('item') || ctx.params.get('resource') || itemIDs[0] || '';
  let selected=ctx.params.get('container') || '';
  let query='';let positiveOnly=false;
  const options=item && !itemIDs.includes(item)?[item,...itemIDs]:itemIDs;
  const itemControl=u.el('select',{onChange:e=>{item=e.target.value;draw();}},options.map(id=>u.el('option',{value:id},`${label(id)} · ${id}`)));
  itemControl.value=item;
  const search=u.el('input',{type:'search',placeholder:'Filter container names',onInput:e=>{query=e.target.value;drawDistribution();}});
  const check=u.el('input',{type:'checkbox',onChange:e=>{positiveOnly=e.target.checked;drawDistribution();}});
  root.append(u.notice('Scope: containers whose saved faction is exactly Player. Unknown and other factions are excluded. No all-world totals are used.'),u.el('div',{className:'controls'},u.field('Resource',itemControl,'Exact saved item class; no stock aggregation across item types.'),u.field('Find storage',search),u.el('label',{className:'field check-field'},check,'Show positive saved stock only')));
  const list=u.el('section',{'aria-label':'Resource distribution'});
  const detail=u.el('section',{className:'panel','aria-label':'Storage evidence'});
  root.append(list,detail);
  function drawDistribution() {
    list.replaceChildren(u.el('h2',{},item?`${label(item)} distribution`:'Resource distribution'));
    if(containers===null) {list.append(u.empty('Container records missing','The source did not provide a container list. Missing evidence is not zero stock.'));return;}
    if(!containers.length) {list.append(u.empty('No explicit Player containers found','Unknown ownership is excluded. Empty buildings may be omitted by the parser; this is not proof that storage or stock is zero.'));return;}
    if(!item) {list.append(u.empty('No saved resource keys available','Player container records exist, but no resource keys are recorded. Missing quantities remain unknown.'));return;}
    const rows=distribution(containers,item,query,positiveOnly);
    if(!rows.length) {list.append(u.empty('No containers match these filters','Try a different name or include unknown/zero quantities. Absence is not evidence of zero settlement stock.'));return;}
    list.append(u.el('p',{className:'muted'},'A resource absent from a container record is shown as not recorded. Only an explicit numeric zero means known zero.'),u.el('div',{className:'distribution-list'},rows.map(c=>{
      const qty=quantity(c,item);
      return u.el('article',{className:'panel'},u.el('h3',{},containerName(c)),u.el('dl',{},u.el('dt',{},'Saved quantity'),u.el('dd',{},qty===null?'Not recorded / unknown':`${qty} units`),u.el('dt',{},'Saved position'),u.el('dd',{},coordinates(c))),u.button(`Inspect ${containerName(c)}`,()=>{selected=c.id;drawDetail();detail.querySelector('h2')?.focus();},{disabled:!c.id}));
    })));
  }
  function drawDetail() {
    const c=containers?.find(c=>c.id && c.id===selected);
    detail.replaceChildren();
    if(!c) {detail.append(u.el('h2',{},selected?'Selected container not in Player evidence':'Inspect storage'),u.el('p',{},selected?'The saved ID was absent or not explicitly Player-owned. This does not establish demolition, emptiness or ownership.':'Choose a distribution record to inspect its saved contents.'));return;}
    detail.append(u.el('h2',{tabindex:'-1',className:'storage-title'},containerName(c)),u.el('dl',{},u.el('dt',{},'Stable saved ID'),u.el('dd',{className:'item-id'},c.id),u.el('dt',{},'Saved class'),u.el('dd',{},u.value(c.cls)),u.el('dt',{},'Saved faction'),u.el('dd',{},c.faction),u.el('dt',{},'Saved position'),u.el('dd',{},coordinates(c)),u.el('dt',{},'Capacity'),u.el('dd',{},'Unknown; no capacity contract'),u.el('dt',{},'Walkable route'),u.el('dd',{},'Not verified'),u.el('dt',{},'Observed transfers'),u.el('dd',{},'Not measured')));
    const ids=c.items && typeof c.items==='object'?Object.keys(c.items).sort():null;
    if(ids?.length) detail.append(u.el('h3',{},'Saved contents'),u.el('ul',{className:'list'},ids.map(id=>u.el('li',{},u.el('strong',{},label(id)),u.el('p',{className:'item-id'},id),u.el('p',{},quantity(c,id)===null?'Quantity unknown':`${quantity(c,id)} units saved`)))));
    else detail.append(u.notice(ids===null?'Container contents were not provided. Quantity remains unknown.':'No item keys were recorded in this container. This is not a measured production or delivery result.'));
    detail.append(u.notice('Storage supports supply access. Saved contents alone cannot establish throughput, stock targets or a storage efficiency score.'));
  }
  function draw(){drawDistribution();drawDetail();}
  draw();
  root.append(u.el('section',{className:'panel'},u.el('h2',{},'Source and context'),u.el('dl',{},u.el('dt',{},'Snapshot'),u.el('dd',{},u.value(world.snapshot_id)),u.el('dt',{},'Save capture time'),u.el('dd',{},u.value(world.captured_at||world.meta?.captured_at)),u.el('dt',{},'Ingested time'),u.el('dd',{},u.value(world.ingested_at))),u.source(result),u.button('Refresh saved stock',e=>refreshEvidence(e.currentTarget,new URLSearchParams({...Object.fromEntries(ctx.params),item,container:selected}))),u.link('Return to location planning','plan',Object.fromEntries(ctx.params))));
  return root;
}
