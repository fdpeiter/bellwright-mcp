import { SKILLS, playerPeople, nameOf, evidence, extractionBuildings, shortlist, proposalRecord } from './model.js';

export async function render(ctx) {
  const u = ctx.ui;
  const root = u.el('div', { className: 'workforce-feature' }, u.heading('Who could gather here?', 'Compare saved evidence for an extractor. Skills and condition indicate suitability; present assignment and available work time remain unknown.'));
  async function refreshEvidence(button, params=ctx.params) {
    button.disabled=true; root.setAttribute('aria-busy','true');
    try {
      await ctx.loadTool('read_saved_world', {}, {refresh:true});
      const updated=await render({...ctx,params}); root.replaceWith(updated);
      const heading=updated.querySelector('h1'); heading?.setAttribute('tabindex','-1');heading?.focus();
    } finally { root.removeAttribute('aria-busy');button.disabled=false; }
  }
  root.append(u.el('style', {}, '.workforce-feature .candidate-list{display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,280px),1fr));gap:16px}.workforce-feature .candidate-list article{margin:0}.workforce-feature .detail-title:focus{outline:3px solid var(--accent);outline-offset:4px}'));
  if (!ctx.tools.some(t => (typeof t === 'string' ? t : t.name) === 'read_saved_world')) {
    root.append(u.empty('Saved villager evidence unavailable', 'The connected service does not expose read_saved_world.', u.link('Check source health', 'workspace')));
    return root;
  }
  const result = await ctx.loadTool('read_saved_world', {});
  if (!result?.data) {
    root.append(u.notice(result?.error || 'Saved evidence could not be loaded.', 'error'), u.button('Retry saved evidence', e => refreshEvidence(e.currentTarget)), u.link('Check source health', 'workspace'));
    return root;
  }
  if (result.status === 'stale') root.append(u.notice('Saved evidence retained after a failed refresh. ' + (result.error || ''), 'error'));
  const world = result.data;
  const people = playerPeople(world);
  const buildings = extractionBuildings(world);
  let buildingID = ctx.params.get('building') || '';
  let selected = ctx.params.get('person') || '';
  let skill = Object.hasOwn(SKILLS, ctx.params.get('skill')) ? ctx.params.get('skill') : 'harvest';
  let query = '';
  const returnLink=u.link('Return to location planning','plan',Object.fromEntries(ctx.params));
  const select = (values, value, onChange) => { const control = u.el('select', { onChange }, values.map(([id,text]) => u.el('option', { value:id },text))); control.value=value; return control; };
  const titleOf = b => b.name || b.cls + ' · ' + b.id;
  const buildingControl = select([['','Choose a saved extractor'], ...buildings.map(b => [b.id,titleOf(b)])], buildings.some(b => b.id === buildingID) ? buildingID : '', e => { buildingID = e.target.value; selected=''; draw(); });
  const skillControl = select(Object.entries(SKILLS), skill, e => { skill=e.target.value; draw(); });
  const searchControl = u.el('input', { type:'search', placeholder:'Search saved villagers', onInput:e => { query=e.target.value; drawCandidates(); } });
  root.append(u.el('div', { className:'controls' }, u.field('Extractor context',buildingControl,'Only explicit Player foraging, logging and mining camp container records are listed. Missing empty buildings remain unknown.'), u.field('Compare saved skill',skillControl,'Choose the relevant evidence; no automatic productivity ranking.'),u.field('Find a candidate',searchControl)));
  const status = u.el('div');
  const candidates = u.el('section', { 'aria-label':'Candidate evidence' });
  const detail = u.el('section', { className:'panel', 'aria-label':'Selected person evidence' });
  const proposal = u.el('section', { className:'panel' });
  root.append(status,candidates,detail,proposal);
  const fmt = value => u.value(value);
  const injuries = value => value === null ? 'Unknown' : value.length ? value.map(i => typeof i === 'string' ? i : i.type || 'Recorded injury').join(', ') : 'None recorded in save';
  const tools = value => value === null ? 'Unknown' : value.length ? value.map(([slot, item]) => `${slot}: ${item}`).join('; ') : 'No tool/hand item recorded';
  function drawCandidates() {
    const rows = shortlist(people, query, skill);
    candidates.replaceChildren(u.el('h2', {}, `${SKILLS[skill]} evidence`), u.el('p', { className:'muted' }, 'Sorted by recorded skill level; unknown levels last. This is not observed work or an assignment.'));
    if (people === null) { candidates.append(u.empty('Villager records missing','The saved source did not provide an NPC list. Missing evidence is not an empty workforce.')); return; }
    if (!rows.length) { candidates.append(u.empty(people.length ? 'No candidates match' : 'No explicit Player villagers found',people.length ? 'Change the name filter.' : 'Only records with is_player_npc=true are eligible. Unknown ownership is excluded.')); return; }
    candidates.append(u.el('div', { className:'candidate-list' },rows.map(p => {
      const ev=evidence(p,skill);
      return u.el('article',{className:'panel'},u.el('h3',{},nameOf(p)),u.el('dl',{},u.el('dt',{},'Saved level / cap'),u.el('dd',{},`${fmt(ev.level)} / ${fmt(ev.cap)}`),u.el('dt',{},'Saved morale'),u.el('dd',{},fmt(ev.morale)),u.el('dt',{},'Injuries'),u.el('dd',{},injuries(ev.injuries))),u.button(`Inspect ${nameOf(p)}`,() => { selected=p.guid; drawDetail(); detail.querySelector('h2')?.focus(); },{disabled:!p.guid}));
    })));
  }
  function drawDetail() {
    const p = people?.find(p => p.guid && p.guid === selected);
    detail.replaceChildren();
    if (!p) { detail.append(u.el('h2', {},'Inspect a person'),u.el('p',{},'Choose a candidate to review tools, condition and saved priorities before proposing staffing.')); return; }
    const ev=evidence(p,skill);
    detail.append(u.el('h2',{tabindex:'-1',className:'detail-title'},nameOf(p)),u.el('dl',{},u.el('dt',{},SKILLS[skill]+' level / cap'),u.el('dd',{},`${fmt(ev.level)} / ${fmt(ev.cap)}`),u.el('dt',{},'Saved equipment evidence'),u.el('dd',{},tools(ev.tools)),u.el('dt',{},'Saved morale'),u.el('dd',{},fmt(ev.morale)),u.el('dt',{},'Recorded injuries'),u.el('dd',{},injuries(ev.injuries)),u.el('dt',{},'Saved priority value'),u.el('dd',{},fmt(ev.priority)),u.el('dt',{},'Present assignment'),u.el('dd',{},'Unknown'),u.el('dt',{},'Available worker game-time'),u.el('dd',{},'Not measured')),
      u.notice('Priorities describe a saved preference, not activity. Priority direction is not inferred here. Equipment presence does not establish tool suitability or condition.'),u.el('details',{},u.el('summary',{},'Other saved skills'),u.el('ul',{},Object.entries(p.skills || {}).map(([key,s]) => u.el('li',{},`${key}: ${fmt(s?.level)} / ${fmt(s?.cap)}`)))));
    drawProposal();
  }
  function drawProposal() {
    const p=people?.find(p => p.guid && p.guid===selected);
    const building=buildings.find(b => b.id===buildingID);
    proposal.replaceChildren(u.el('h2',{},'Staffing proposal'),u.el('p',{},'Save a local plan for this extractor. Make assignments in game; Steward does not change the game or measure work manually.'));
    const message=u.el('p',{role:'status'});
    proposal.append(u.button('Save staffing proposal',() => {
      const record=proposalRecord(p,building,skill,world.snapshot_id);
      if (!record) { message.textContent='Select a known person and saved Player extractor first.'; return; }
      try {
        sessionStorage.setItem('steward-v2-staffing:'+building.id,JSON.stringify(record));
        message.textContent=`Proposal saved for ${nameOf(p)} at ${titleOf(building)} in this browser session. No game assignment was made.`;
      } catch { message.textContent='Browser session storage is unavailable. The proposal could not be saved; game state was not changed.'; }
    },{className:'primary',disabled:!p?.guid || !building}),message);
    if (building) {
      try { const saved=JSON.parse(sessionStorage.getItem('steward-v2-staffing:'+building.id)||'null'); if(saved) proposal.append(u.el('p',{className:'muted'},`Local proposal exists for saved person ${saved.person}, snapshot ${fmt(saved.snapshot)}. Recheck evidence before using it.`)); } catch { /* Storage errors are reported on save. */ }
    }
  }
  function draw() {
    const building=buildings.find(b => b.id===buildingID);
    returnLink.setAttribute('href', '#plan?' + new URLSearchParams({...Object.fromEntries(ctx.params),building:buildingID}));
    status.replaceChildren(u.notice(building ? `Planning for ${titleOf(building)}. Known saved skill is evidence, not actual assignment.` : 'Select an extractor to enable a staffing proposal. Saved workforce evidence remains available.'));
    drawCandidates(); drawDetail(); drawProposal();
  }
  draw();
  root.append(u.el('section',{className:'panel'},u.el('h2',{},'Source and context'),u.el('dl',{},u.el('dt',{},'Snapshot'),u.el('dd',{},fmt(world.snapshot_id)),u.el('dt',{},'Save capture time'),u.el('dd',{},fmt(world.captured_at || world.meta?.captured_at)),u.el('dt',{},'Ingested time'),u.el('dd',{},fmt(world.ingested_at))),u.source(result),u.button('Refresh saved evidence',e => refreshEvidence(e.currentTarget,new URLSearchParams({...Object.fromEntries(ctx.params),building:buildingID,person:selected,skill}))),returnLink));
  return root;
}
