// Parse the actual embedded browser JavaScript before rebuilding the executable.
const fs=require('fs'),vm=require('vm'),path=require('path');
const html=fs.readFileSync(path.join(__dirname,'index.html'),'utf8');
new vm.Script(html.split('<script>')[1].split('</script>')[0],{filename:'livemap/index.html'});
console.log('Embedded map JavaScript syntax OK');
