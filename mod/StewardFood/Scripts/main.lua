-- Stable loader: exactly one timer; reader edits require no global mod restart.
local scriptDir=debug.getinfo(1,'S').source:match('^@(.+[/\\])')
local previous,reader=nil,nil
local function unavailable(status)
 print('BWFOOD {"source":"ue4ss_live","session_id":"food-loader","captured_at":"'..os.date('!%Y-%m-%dT%H:%M:%SZ')..'","available":false,"write_available":false,"status":"'..status..'"}\n')
end
local function tick()
 local file=scriptDir and io.open(scriptDir..'sample.lua','rb')
 if not file then unavailable('reader_file_unavailable');return end
 local source=file:read(128*1024+1);file:close()
 if not source or #source>128*1024 then unavailable('reader_file_invalid');return end
 if source~=previous then
  local chunk=load(source,'@'..scriptDir..'sample.lua','t',_ENV)
  if not chunk then unavailable('reader_compile_failed');return end
  local ok,nextReader=pcall(chunk)
  if not ok or type(nextReader)~='function' then unavailable('reader_load_failed');return end
  reader=nextReader;previous=source
  print('BWFOOD_RELOAD reader_loaded\n')
 end
 local ok,line=pcall(reader)
 if not ok or type(line)~='string' or #line>60000 then unavailable('reader_sample_failed');return end
 print('BWFOOD '..line..'\n')
end
if type(LoopInGameThreadWithDelay)~='function' then unavailable('timer_unavailable');return end
LoopInGameThreadWithDelay(10000,tick)
