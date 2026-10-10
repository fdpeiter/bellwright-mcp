export type RuntimePoint={id:string;name:string;class:string;kind:string;position:number[];stale:boolean};
export function freshRuntimePoints(feed:{status?:{stale?:boolean;waiting_for_player?:boolean};points?:RuntimePoint[]}){
  if(feed.status?.stale!==false||feed.status.waiting_for_player)return [];
  return (feed.points||[]).filter(p=>!p.stale&&p.position?.length>=2&&p.position.slice(0,2).every(Number.isFinite));
}
