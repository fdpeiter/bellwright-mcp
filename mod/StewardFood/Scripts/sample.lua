-- Bellwright Steward: bounded live food investigation. No game or save writes.
local helpers=require('UEHelpers')
local session='food-1-'..tostring(os.time())
local function valid(o) local ok,v=pcall(function() return o and o:IsValid() end); return ok and v==true end
local function read(o,k) local ok,v=pcall(function() return o[k] end); if ok then return v end;return nil end
local function id(o) if not valid(o) then return nil end; local ok,v=pcall(function() return o:GetFullName() end); if ok then return v end end
local function ref(o) return id(o) end
local function arr() return setmetatable({},{array=true}) end
local function quote(s) return '"'..tostring(s):sub(1,1024):gsub('[%z\1-\31\\"]',function(c) if c=='"' then return '\\"' elseif c=='\\' then return '\\\\' else return string.format('\\u%04x',string.byte(c)) end end)..'"' end
local function json(v)
 if v==nil then return 'null' end
 if type(v)=='string' then return quote(v) end
 if type(v)=='boolean' then return v and 'true' or 'false' end
 if type(v)=='number' then return v==v and v>-math.huge and v<math.huge and tostring(v) or 'null' end
 local t={}; if getmetatable(v) and getmetatable(v).array then for _,x in ipairs(v) do t[#t+1]=json(x) end; return '['..table.concat(t,',')..']' end
 for k,x in pairs(v) do t[#t+1]=quote(k)..':'..json(x) end; return '{'..table.concat(t,',')..'}'
end
local function each(v,limit,fn)
 local ok,n=pcall(function() return #v end); if not ok or type(n)~='number' then return {available=false} end
 local failed=0; for i=1,math.min(n,limit) do if not pcall(fn,v[i],i) then failed=failed+1 end end
 return {available=true,total=n,truncated=n>limit,failed=failed}
end
local foodClasses={'ForagingCamp_C','Campfire_C','SmokingRack_C','SmallFarm_C','TrapperCamp_C','RabbitTrap_C','WaterCollector_C'}
local function sameWorld(a,wid) local ok,w=pcall(function() return a:GetWorld() end);return ok and id(w)==wid end
local function sample()
 local p=helpers.GetPlayer(); local out={source='ue4ss_live',session_id=session,captured_at=os.date('!%Y-%m-%dT%H:%M:%SZ'),available=valid(p),write_available=false,write_status='setter_unverified',buildings=arr(),settlements=arr(),discovery=arr(),scope='food_buildings_within_500m_engine_units_unverified'}
 if not valid(p) then out.status='waiting_for_player'; return out end
 local world=p:GetWorld(); local wid=id(world); local origin=p:K2_GetActorLocation(); local relevant={}
 out.player=id(p);out.player_data=ref(read(p,'Data')); out.current_settlement=ref(read(p,'CurrentSettlement')); out.host=read(p,'bIsHostPlayer')==true
 for _,cls in ipairs(foodClasses) do
  local found=FindAllOf(cls) or {}
  out.discovery[#out.discovery+1]={class=cls,total=#found,examined=math.min(#found,64),truncated=#found>64}
  for i=1,math.min(#found,64) do
   local a=found[i]
   local near=false
   if valid(a) and sameWorld(a,wid) then local pok,pos=pcall(function()return a:K2_GetActorLocation()end);if pok then near=(pos.X-origin.X)^2+(pos.Y-origin.Y)^2+(pos.Z-origin.Z)^2<=50000^2 end end
   if near then
    local data=read(a,'StructureData'); local orders=arr(); local rules=arr()
    local orderStatus=each(read(data,'Orders'),32,function(o,index)
     orders[#orders+1]={index=index,item=ref(read(o,'Item')),target=tonumber(read(o,'TargetAmount')),amount=tonumber(read(o,'Amount')),order_id=ref(read(o,'ID'))}
    end)
    local ruleStatus=each(read(data,'StorageRulesAmounts'),32,function(r,index) rules[#rules+1]={index=index,min=tonumber(read(r,'Min')),max=tonumber(read(r,'Max'))} end)
    local amounts=arr();local amountStatus={available=false,status='map_unavailable'};local count=0
    local mok=pcall(function()read(data,'InfiniteCraftingAmounts'):ForEach(function(k,v)
     count=count+1;amounts[#amounts+1]={item=ref(k:get()),amount=tonumber(v:get())};return count>=32
    end)end)
    if mok then amountStatus={available=true,exported=count,truncated=count>=32} elseif count>0 then amountStatus={available=false,status='partial_map_read',exported=#amounts,truncated=true} end
    local settlement=ref(read(data,'Settlement')); if settlement then relevant[settlement]=true end
    out.buildings[#out.buildings+1]={id=id(a),class=cls,data=id(data),settlement=settlement,ownership='unknown',orders=orders,orders_status=orderStatus,storage_rules=rules,storage_rules_status=ruleStatus,infinite_crafting_amounts=amounts,infinite_crafting_amounts_status=amountStatus}
   end
  end
 end
 local settlements=FindAllOf('MistSettlementData') or {}
 for i=1,math.min(#settlements,64) do
  local s=settlements[i]; local a=read(s,'Actor')
  if relevant[id(s)] and valid(a) and sameWorld(a,wid) then
   local residents=arr(); local residentStatus=each(read(s,'Residents'),64,function(n) residents[#residents+1]={id=id(n),faction=ref(read(n,'Faction')),actor=ref(read(n,'NpcActor'))} end)
   local structures=arr(); local structureStatus=each(read(s,'Structures'),128,function(d) structures[#structures+1]={id=id(d),actor=ref(read(d,'StructureActor'))} end)
   out.settlements[#out.settlements+1]={id=id(s),faction=ref(read(s,'Faction')),residents=residents,residents_status=residentStatus,structures=structures,structures_status=structureStatus,ownership='unknown'}
  end
 end
 local containers=arr();out.containers=containers
 local found=FindAllOf('MistContainerComponent') or {};out.container_discovery={total=#found,examined=math.min(#found,2048),truncated=#found>2048}
 for i=1,math.min(#found,2048) do
  local c=found[i];local aok,a=pcall(function()return c:GetOwner()end)
  if aok and valid(a) and sameWorld(a,wid) then
   local sd=read(a,'StructureData');local settlement=ref(read(sd,'Settlement'))
   if settlement and relevant[settlement] and #containers<32 then
    local slots=arr();local slotsStatus=each(read(c,'Slots'),64,function(slot,index)
     local quantity=tonumber(read(slot,'Count'));if quantity and quantity>0 then slots[#slots+1]={index=index,item=ref(read(read(slot,'Item'),'Template')),quantity=quantity} end
    end)
    containers[#containers+1]={id=id(c),actor=id(a),settlement=settlement,slots=slots,slots_status=slotsStatus,ownership='unknown'}
   elseif settlement and relevant[settlement] then out.container_discovery.truncated=true end
  end
 end
 out.status='observed'; return out
end
return function()
 local ok,out=pcall(sample)
 if ok and out.available then
  local dir=debug.getinfo(1,'S').source:match('^@(.+[/\\])')
  local sok,storage=pcall(function()return dofile(dir..'storage.lua')(helpers.GetPlayer())end)
  if sok then
   out.storage=storage
   local observed,trace=pcall(function()return dofile(dir..'observe.lua')()end)
   if observed then out.storage.transfer_observer=trace else out.storage.observer_status='unavailable' end
  else out.storage_status='read_failed' end
 end
 if not ok then out={source='ue4ss_live',session_id=session,captured_at=os.date('!%Y-%m-%dT%H:%M:%SZ'),available=false,status='read_failed',read_error=tostring(out):sub(1,512),write_available=false} end
 local line=json(out)
 if #line>60000 then
  out.budget_truncated=true
  for _,s in ipairs(out.settlements or {}) do s.structures=nil;if s.structures_status then s.structures_status.truncated=true end end
  line=json(out)
  while #line>60000 and out.containers and #out.containers>0 do table.remove(out.containers);out.container_discovery.truncated=true;line=json(out) end
  while #line>60000 and out.storage and #out.storage.containers>0 do table.remove(out.storage.containers);out.storage.discovery.truncated=true;line=json(out) end
 end
 if #line>60000 then return json({source='ue4ss_live',session_id=session,captured_at=out.captured_at,available=false,status='budget_exceeded',write_available=false}) end
 return line
end
-- Automatic reload verification: behavior unchanged.
