local function object(name,fields)
 local o=fields or {};function o:IsValid()return true end;function o:GetFullName()return name end;function o:GetClass()return object('Class Test') end;return o
end
local world=object('World Test')
local player=object('Player Test',{bIsHostPlayer=true})
function player:GetWorld()return world end
function player:K2_GetActorLocation()return {X=0,Y=0,Z=0} end
local settlement=object('Settlement Test',{Residents={},Structures={},Storage={Items={}}})
local settlementActor=object('SettlementActor Test');function settlementActor:GetWorld()return world end;settlement.Actor=settlementActor
local item=object('Item Mushroom')
local data=object('FoodData Test',{Settlement=settlement,Orders={{Item=item}},StorageRulesAmounts={}})
local near=object('Camp Test',{StructureData=data});function near:GetWorld()return world end;function near:K2_GetActorLocation()return {X=10,Y=0,Z=0} end
local far=object('FarCamp Test',{StructureData=data});function far:GetWorld()return world end;function far:K2_GetActorLocation()return {X=90000,Y=0,Z=0} end
local active=player
package.preload.UEHelpers=function()return {GetPlayer=function()return active end} end
function FindAllOf(name) if name=='ForagingCamp_C' then return {near,far} elseif name=='MistSettlementData' then return {settlement} end;return {} end
local callback;function LoopInGameThreadWithDelay(_,fn)callback=fn end
local messages={};local savedPrint=print;print=function(s)messages[#messages+1]=s end
local sample=dofile('mod/StewardFood/Scripts/sample.lua');callback=function()messages[#messages+1]=sample() end;callback()
assert(messages[#messages]:find('"status":"observed"',1,true))
assert(messages[#messages]:find('Camp Test',1,true))
assert(not messages[#messages]:find('FarCamp Test',1,true))
assert(messages[#messages]:find('"write_available":false',1,true))
assert(not messages[#messages]:find('"target"',1,true)) -- missing fields must not become zero
active=nil;callback();assert(messages[#messages]:find('"status":"waiting_for_player"',1,true))
assert(not messages[#messages]:find('Camp Test',1,true))
print=savedPrint;print('Food probe tests passed')
