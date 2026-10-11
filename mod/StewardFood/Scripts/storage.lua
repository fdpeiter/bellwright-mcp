-- Storage transfer prerequisites. This module performs no writes.
return function(player)
 local function read(o,k)local ok,v=pcall(function()return o[k]end);if ok then return v end;return nil end
 local function id(o)local ok,v=pcall(function()if o:IsValid()then return o:GetFullName()end end);if ok then return v end end
 local function list()return setmetatable({},{array=true})end
 local function position(o)local ok,p=pcall(function()return o:K2_GetActorLocation()end);if ok then return {x=p.X,y=p.Y,z=p.Z} end end
 local origin=position(player);local world=id(player:GetWorld());local out={containers=list(),scope='loaded_storage_buildings',ownership='unknown',write_available=false}
 local found=FindAllOf('MistContainerComponent') or {};out.discovery={found=#found,examined=math.min(#found,2048),truncated=#found>2048}
 for i=1,math.min(#found,2048)do
  local c=found[i];local ok,a=pcall(function()return c:GetOwner()end);local aid=ok and id(a)
  if aid and (aid:match('^Stockpile_C ') or aid:match('^SharedChest_C ') or aid:match('^ShackChest_C '))then
   local wok,w=pcall(function()return a:GetWorld()end)
   local pos=position(a)
   if wok and id(w)==world and pos and origin and (pos.x-origin.x)^2+(pos.y-origin.y)^2<=50000^2 and #out.containers<16 then
    local sd=read(a,'StructureData');local settlement=read(sd,'Settlement');local slots=list();local values=read(c,'Slots')
    local nok,n=pcall(function()return #values end);local failed=0
    if nok then for index=1,math.min(n,128)do
     local sok=pcall(function()
      local slot=values[index];local item=read(slot,'Item');local template=read(item,'Template');local quantity=tonumber(read(slot,'Count'))
      slots[#slots+1]={lua_index=index,quantity=quantity,item=id(template),item_max_stack=tonumber(read(template,'MaxStackSize')),custom_value_1=tonumber(read(item,'CustomValue1')),custom_value_2=tonumber(read(item,'CustomValue2'))}
     end);if not sok then failed=failed+1 end
    end end
    local fullok,full=pcall(function()return c:IsFull()end)
    local emptyok,empty=pcall(function()return c:GetEmptySlotCount()end)
    local fit=list()
    if nok then for index=1,math.min(n,8)do
     local slot=values[index];local item=read(slot,'Item')
     if id(read(item,'Template')) then
      local fok,f=pcall(function()return c:HasSpaceForItems(item,1)end)
      fit[#fit+1]={lua_index=index,available=fok,fits_one=fok and f or nil}
     end
    end end
    out.containers[#out.containers+1]={id=id(c),actor=aid,position=pos,settlement=id(settlement),faction=id(read(settlement,'Faction')),slots=slots,slots_status={available=nok,total=nok and n or nil,truncated=nok and n>128 or false,failed=failed},native_capacity={full_available=fullok,full=fullok and full or nil,empty_available=emptyok,empty_slots=emptyok and empty or nil,existing_item_fit=fit},can_take=read(c,'bCanTake'),can_put=read(c,'bCanPut'),private=read(c,'bPrivate'),allow_all_items=read(c,'bAllowAllItems'),allow_merging=read(c,'bAllowItemMerging'),max_stack_override=tonumber(read(c,'MaxStackSizeOverride')),initial_capacity=tonumber(read(c,'InitialCapacity')),max_open_distance=tonumber(read(c,'MaxOpenDistance'))}
    local capacity=out.containers[#out.containers].native_capacity
    if fullok then capacity.full=full end
    capacity.fit_sample_truncated=nok and n>8 or false
    for _,f in ipairs(fit)do if f.available and f.fits_one==nil then f.fits_one=false end end
   elseif wok and id(w)==world and pos and origin and (pos.x-origin.x)^2+(pos.y-origin.y)^2<=50000^2 then
    out.discovery.truncated=true
   end
  end
 end
 out.player_position=origin;out.open_container_actor=id(read(read(player,'InventoryComponent'),'OpenContainerActor'))
 return out
end
