local function object(name,fields)
 local o=fields or {};function o:GetFullName()return name end;function o:IsValid()return true end;return o
end
local world=object('World Test');local player=object('Player Test')
function player:GetWorld()return world end
function player:K2_GetActorLocation()return {X=0,Y=0,Z=0}end
local template=object('Stone Test',{MaxStackSize=100})
local actor=object('Stockpile_C Test',{StructureData={Settlement=object('Settlement Test',{Faction=object('Faction Test')})}})
function actor:GetWorld()return world end
function actor:K2_GetActorLocation()return {X=0,Y=0,Z=0}end
local container=object('Container Test',{Slots={{Count=3,Item={Template=template,CustomValue1=0,CustomValue2=0}}},bCanTake=true,bCanPut=true,bPrivate=false})
function container:GetOwner()return actor end
function container:IsFull()return false end
function container:GetEmptySlotCount()return 0 end
function container:HasSpaceForItems()return false end
function FindAllOf()return {container}end
local sample=dofile('mod/StewardFood/Scripts/storage.lua')
local out=sample(player);local c=out.containers[1]
assert(out.write_available==false and out.ownership=='unknown')
assert(c.slots[1].quantity==3 and c.slots[1].custom_value_1==0)
assert(c.native_capacity.full==false and c.native_capacity.empty_slots==0)
assert(c.native_capacity.existing_item_fit[1].fits_one==false)
container.IsFull=function()error('unavailable')end
out=sample(player);assert(out.containers[1].native_capacity.full_available==false)
container.Slots={};for i=1,129 do container.Slots[i]={Count=0,Item={}}end
out=sample(player);assert(#out.containers[1].slots==128 and out.containers[1].slots_status.truncated)
print('Storage probe tests passed')
