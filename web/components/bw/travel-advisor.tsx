'use client';
import {useState} from 'react';
import type {MapDef} from '@/lib/bw/map';
export type TravelPoint={key:string;label:string;sub:string;u:number;v:number;live?:boolean};
export function travelDistance(a:TravelPoint,b:TravelPoint,map:MapDef){
  const [x,y,,z,w]=map.transform,det=x*w-y*z,du=a.u-b.u,dv=a.v-b.v;
  return Math.hypot((w*du-y*dv)/det,(-z*du+x*dv)/det);
}
export function TravelAdvisor({points,map,onRoute,onFocus,source}:{points:TravelPoint[];map:MapDef;onRoute:(ids:[string,string]|null)=>void;onFocus:(a:TravelPoint,b:TravelPoint)=>void;source:string}){
  const [hubID,setHub]=useState(''),[targetID,setTarget]=useState(''),[trips,setTrips]=useState(1);
  const hub=points.find(p=>p.key===hubID),target=points.find(p=>p.key===targetID);
  const hubs=points.filter(p=>/stockpile|chest|barn|storage|cellar/i.test(p.label+' '+p.sub));
  const rows=hub?points.filter(p=>p.key!==hub.key).map(p=>({p,d:travelDistance(p,hub,map)})).sort((a,b)=>b.d-a.d):[];
  const distance=hub&&target?travelDistance(hub,target,map):null;
  const select='mt-1 mb-2 w-full rounded border border-line-3 bg-ink px-2 py-2 text-sand-200';
  return <section className="my-3 rounded-lg border border-line-3 bg-ink p-3 text-xs text-sand-400">
    <h2 className="mb-2 font-semibold text-sm text-gold-bright">Building travel advisor</h2>
    <p className="mb-2">{source} Choose buildings you own. Longer connections appear first.</p>
    <label>Receiving storage hub<select aria-label="Receiving storage hub" className={select} value={hub?hubID:''} onChange={e=>{setHub(e.target.value);setTarget('');onRoute(null)}}><option value="">Choose storage…</option>{hubs.map(p=><option key={p.key} value={p.key}>{p.label} · {p.live?'live':'save'}</option>)}</select></label>
    <label>Building to connect<select aria-label="Building to connect" className={select} value={target?targetID:''} onChange={e=>{setTarget(e.target.value);onRoute(hub?[hub.key,e.target.value]:null)}}><option value="">Choose building…</option>{rows.map(({p,d})=><option key={p.key} value={p.key}>{p.label} · {Math.round(d).toLocaleString()} units · {p.live?'live':'save'}</option>)}</select></label>
    <label>Estimated round trips per day<input aria-label="Estimated round trips per day" className={select} type="number" min={1} max={1000} value={trips} onChange={e=>setTrips(Number(e.target.value))}/></label>
    {distance!==null&&<><p>One-way separation: {Math.round(distance).toLocaleString()} engine units. {Number.isFinite(trips)&&trips>=1&&trips<=1000?<>Estimated daily round-trip distance: {Math.round(distance*2*trips).toLocaleString()} units.</>:'Enter 1–1000 trips.'}</p><p className="mt-2">{/mining|logging|foraging|trapper|mud|lumber/i.test(target?.label+' '+target?.sub)?'Keep this workplace near its resource. Test nearby receiving storage or outpost supplies when hauling dominates work.':'If hauling dominates work, consider moving this building closer to its input storage.'} Inspect a clear corridor between entrances and compare actual trips before and after changes.</p><button className="mt-2 rounded border border-line-3 px-2 py-2 text-gold" onClick={()=>hub&&target&&onFocus(hub,target)}>Focus proposed connection</button></>}
    <p className="mt-2">Cyan dashed line is schematic. Terrain, entrances, ownership and road speed benefits are unverified. Distance is a planning estimate, not measured productivity. Live and save identities may differ.</p>
  </section>;
}
