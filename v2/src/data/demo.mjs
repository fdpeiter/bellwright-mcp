// Entirely synthetic fixtures. Never derived from a private save or capture.
export const TOOL_NAMES = ['list_save_snapshots','read_saved_world','read_save_trends','read_live_map','read_resource_baseline','read_foraging_candidates','read_companion_telemetry'];
const time = '2026-01-01T12:00:00Z';
export const DEMO_TOOLS = TOOL_NAMES.map(name => ({name,description:'Synthetic demonstration; read-only evidence',inputSchema:{type:'object',properties:{},additionalProperties:false}}));
const person = (guid,first_name,level,morale) => ({guid,first_name,last_name:'Demo',is_player_npc:true,skills:{harvest:{level,cap:8,xp:100},labour:{level:3,cap:7,xp:50},strength:{level:4,cap:8,xp:60}},equipment:{mainhand:'SimpleAxe_C'},injuries:[],morale,job_priorities:{harvesting:1},traits:[],position:[1000,1000,0]});
const containers = [
 {id:'demo-camp-north',cls:'ForagingCamp_C',name:'North demonstration camp',faction:'Player',position:[0,0,0],items:{Berry_C:12,Mushroom_C:4}},
 {id:'demo-camp-south',cls:'ForagingCamp_C',name:'South demonstration camp',faction:'Player',position:[12000,0,0],items:{Berry_C:0}},
 {id:'demo-stockpile',cls:'Stockpile_C',name:'Demonstration stockpile',faction:'Player',position:[2000,3000,0],items:{Berry_C:20,Mushroom_C:7,Wood_C:15}},
 {id:'demo-world-store',cls:'SmallBarn_C',name:'Other faction demonstration',faction:'DemoVillage',position:[20000,0,0],items:{Berry_C:500}}
];
export function demoCall(name) {
 let data;
 switch(name) {
 case 'list_save_snapshots': data={snapshots:[{id:2,save_name:'Synthetic demonstration',region:'Demo region',ingested_at:time,playtime_seconds:7200,mine_count:2},{id:1,save_name:'Synthetic demonstration',region:'Demo region',ingested_at:'2026-01-01T11:50:00Z',playtime_seconds:6600,mine_count:2}]};break;
 case 'read_saved_world':data={snapshot_id:2,ingested_at:time,meta:{saveName:'Synthetic demonstration',region:'Demo region',playtimeSeconds:7200,savedBuild:'synthetic'},npcs:[person('demo-person-ada','Ada',5,65),person('demo-person-ben','Ben',3,50)],carried:{},storage:{containers,totals:{Berry_C:532,Mushroom_C:11,Wood_C:15}},groups:[],player:{guid:'demo-player',position:[1000,0,0],equipment:{},skills:{},carried:[]},villages:[],pois:[]};break;
 case 'read_save_trends':data={scope:'all_world_containers',points:[{snapshot_id:1,ingested_at:'2026-01-01T11:50:00Z',playtime:6600,items:{Berry_C:530,Mushroom_C:12}},{snapshot_id:2,ingested_at:time,playtime:7200,items:{Berry_C:532,Mushroom_C:11}}],all_items:['Berry_C','Mushroom_C'],top_items:['Berry_C','Mushroom_C'],deltas:null};break;
 case 'read_live_map':data={points:containers.slice(0,3).map((c,i)=>({id:'demo-runtime-'+i,kind:'building',name:c.name,class:c.cls,position:c.position,captured_at:time,stale:true,age_seconds:null,details:{ownership:'unknown',unit:'engine_units_unverified'}})),status:{stale:true,captured_at:time,session_id:'synthetic-session',actors:3,buildings:3,villagers:0},unknowns:['Synthetic positions; ownership correlation and scale are not verified.']};break;
 case 'read_resource_baseline':data={schema:'demo',source_kind:'synthetic_demo',nodes:Array.from({length:24},(_,i)=>({id:'demo-node-'+i,baseline_category:i%2?'WildProduce':'Berry',position:[i*900-6000,(i%5)*1200,0]})),unknowns:['Synthetic node positions; categories do not prove item yields or current availability.']};break;
 case 'read_foraging_candidates':data=[{target_runtime_actor_id:'demo-unlinked-actor',current:{position:[0,0,0],baseline_food_node_count:10,distance_to_core_engine_units:3606},practical_candidate:{position:[3000,2000,0],baseline_food_node_count:14,distance_to_core_engine_units:1414},radius_engine_units_assumed:15000,assumptions:['Synthetic candidate is unlinked to saved camp identity.','Coverage, radius, distance, terrain and buildability are unverified.']}];break;
 case 'read_companion_telemetry':data={data:null};break;
 default:throw new Error('Tool is not available');
 }
 return {status:name==='read_live_map'?'stale':'ok',data,source:'synthetic_demo:'+name,note:'Synthetic demonstration only. No private game data. Not measured production.',retrievedAt:time};
}
