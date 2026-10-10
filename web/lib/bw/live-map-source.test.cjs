const assert=require('node:assert/strict');
(async()=>{
const {freshRuntimePoints}=await import('./live-map-source.ts');
const point={id:'player',name:'Player',class:'player',kind:'player',position:[3,4,0],stale:false};
assert.equal(freshRuntimePoints({status:{stale:false},points:[point]}).length,1);
assert.deepEqual(freshRuntimePoints({status:{stale:true},points:[point]}),[]);
assert.deepEqual(freshRuntimePoints({}),[]);
assert.deepEqual(freshRuntimePoints({status:{stale:false,waiting_for_player:true},points:[point]}),[]);
assert.deepEqual(freshRuntimePoints({status:{stale:false},points:[{...point,stale:true},{...point,position:[Infinity,4]}]}),[]);
console.log('Live, stale, unavailable, waiting-player and invalid-coordinate fallback checks passed');

})();

