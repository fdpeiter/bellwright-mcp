-- Passive native-call observation only: no writes, overrides or replay.
local path='/Script/Mist.MistInventoryComponent:ServerMove'
local key='StewardStorageObserverV1'
local function read(o,k)local ok,v=pcall(function()return o[k]end);if ok then return v end end
local function id(o)local ok,v=pcall(function()if o:IsValid()then return o:GetFullName()end end);if ok then return v end end
local function scalar(v)if type(v)=='number' or type(v)=='string' or type(v)=='boolean' then return v end end
local function fields(o)
 local item=read(o,'Item')
 return {type=scalar(read(o,'Type')),actor=id(read(o,'Actor')),container=id(read(o,'Container')),slot_index=scalar(read(o,'SlotIndex')),num_items=scalar(read(o,'NumItems')),origin_count=scalar(read(o,'OriginCount')),item=id(read(item,'Template')),custom_value_1=read(item,'CustomValue1')~=nil and tostring(read(item,'CustomValue1')) or nil,custom_value_2=read(item,'CustomValue2')~=nil and tostring(read(item,'CustomValue2')) or nil}
end
return function()
 local s=rawget(_G,key)
 if not s then
  local dir=debug.getinfo(1,'S').source:match('^@(.+[/\\])')
  local arm=dir and io.open(dir..'storage-observe.arm','rb')
  if not arm then return {status='inactive',write_available=false} end
  arm:close()
  if not os.remove(dir..'storage-observe.arm') then return {status='arm_unavailable',write_available=false} end
  s={status='unavailable',events=setmetatable({},{array=true}),deadline=os.time()+600};rawset(_G,key,s)
  local ok,f=pcall(function()return StaticFindObject(path)end)
  if ok and f and f:IsValid() and type(RegisterHook)=='function' then
   local registered,pre,post=pcall(RegisterHook,path,function(context,source,target,operation)
    if s.status~='observing' or #s.events>=6 or os.time()>s.deadline then return end
    local captured,event=pcall(function()
     return {captured_at=os.date('!%Y-%m-%dT%H:%M:%SZ'),context=id(context:get()),source=fields(source:get()),target=fields(target:get()),operation_id=tostring(operation:get()),kind='observed_request',completion_verified=false}
    end)
    if captured then s.events[#s.events+1]=event else s.status='read_failed' end
    -- No return value: the native call and its parameters remain untouched.
   end)
   if registered and type(pre)=='number' and type(post)=='number' then s.pre=pre;s.post=post;s.status='observing' end
  end
 end
 if s.pre and (os.time()>s.deadline or #s.events>=6 or s.status=='read_failed') then
  local ok=pcall(UnregisterHook,path,s.pre,s.post)
  if ok then s.pre=nil;s.post=nil;if s.status~='read_failed' then s.status='finished' end else s.status='unregister_failed' end
 end
 return {status=s.status,events=s.events,write_available=false,deadline=s.deadline,scope='native_move_requests_not_verified_completions'}
end
