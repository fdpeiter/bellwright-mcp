local clock=100;os.time=function()return clock end
local armed=false
io.open=function()if armed then return {close=function()end} end end
os.remove=function()armed=false;return true end
local function object(name,t)local o=t or {};function o:IsValid()return true end;function o:GetFullName()return name end;return o end
function StaticFindObject()return object('Function')end
local callback;local registrations=0;local removed=0
function RegisterHook(_,fn)registrations=registrations+1;callback=fn;return 1,2 end
function UnregisterHook(_,a,b)assert(a==1 and b==2);removed=removed+1 end
local observe=dofile('mod/StewardFood/Scripts/observe.lua')
assert(observe().status=='inactive' and registrations==0);armed=true
assert(observe().status=='observing');observe();assert(registrations==1)
local source={Type=3,SlotIndex=0,NumItems=1,OriginCount=4,Item={Template=object('Stone'),CustomValue1=0,CustomValue2=0}}
local target={Type=2,SlotIndex=-1}
local function param(v)return {get=function()return v end,set=function()error('must never mutate parameters')end}end
assert(callback(param(object('Context')),param(source),param(target),param(12))==nil)
local state=observe();assert(#state.events==1 and state.events[1].source.slot_index==0)
assert(source.NumItems==1 and source.OriginCount==4 and target.SlotIndex==-1)
assert(state.events[1].completion_verified==false and state.write_available==false)
source.NumItems=99;assert(state.events[1].source.num_items==1) -- copied, not retained
clock=3701;assert(observe().status=='finished' and removed==1)
callback(param(object('Context')),param(source),param(target),param(13));assert(#observe().events==1)
armed=true;assert(observe().status=='observing' and registrations==2 and #observe().events==0)
print('Passive storage observer tests passed')
