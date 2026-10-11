-- Read game-generated operation structs from loaded stone inventory widgets.
-- These are widget state, not executed requests or verified completions.
return function()
 local function read(o,k)local ok,v=pcall(function()return o[k]end);if ok then return v end end
 local function id(o)local ok,v=pcall(function()if o:IsValid()then return o:GetFullName()end end);if ok then return v end end
 local function scalar(v)if type(v)=='number' or type(v)=='string' or type(v)=='boolean' then return v end end
 local function fields(o)
  local item=read(o,'Item')
  return {type=scalar(read(o,'Type')),actor=id(read(o,'Actor')),container=id(read(o,'Container')),slot_index=scalar(read(o,'SlotIndex')),num_items=scalar(read(o,'NumItems')),origin_count=scalar(read(o,'OriginCount')),item=id(read(item,'Template'))}
 end
 local all=FindAllOf('MistItemSlotWidget') or {}
 local out={scope='loaded_stone_widget_state_not_executed_requests',examined=math.min(#all,2048),found=#all,truncated=#all>2048,records=setmetatable({},{array=true}),write_available=false}
 for i=1,math.min(#all,2048)do
  local w=all[i];local template=id(read(read(w,'Item'),'Template'))
  if template and template:match('^CrudeStone_C ') then
   if #out.records<16 then out.records[#out.records+1]={widget=id(w),count=scalar(read(w,'Count')),item=template,source=fields(read(w,'Source')),target=fields(read(w,'Target'))}
   else out.truncated=true end
  end
 end
 return out
end
