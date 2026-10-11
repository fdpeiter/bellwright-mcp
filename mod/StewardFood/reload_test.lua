local source='return function()return "first" end'
io.open=function()return {read=function()return source end,close=function()end} end
local timer,count;count=0
function LoopInGameThreadWithDelay(_,fn)count=count+1;timer=fn end
local messages={};local savedPrint=print;print=function(s)messages[#messages+1]=s end
dofile('mod/StewardFood/Scripts/main.lua');timer();assert(messages[#messages]=='BWFOOD first\n')
source='return function()return "second" end';timer();assert(messages[#messages]=='BWFOOD second\n')
source='invalid Lua source';timer();assert(messages[#messages]:find('reader_compile_failed',1,true))
source='return function()return "third" end';timer();assert(messages[#messages]=='BWFOOD third\n')
assert(count==1);print=savedPrint;print('Food automatic reload tests passed')
