// Offline source audit only. Does not load UE4SS or contact the game.
import fs from 'node:fs';
import path from 'node:path';
const root=process.argv[2];
if(!root)throw Error('Usage: node audit_bridge.mjs <directory containing official UE4SS source files>');
const object=fs.readFileSync(path.join(root,'LuaUObject.cpp'),'utf8');
const wrapper=fs.readFileSync(path.join(root,'LuaMadeSimple.hpp'),'utf8');
const implementation=fs.readFileSync(path.join(root,'LuaMadeSimple.cpp'),'utf8');
const start=object.indexOf('auto push_structproperty(');
const end=object.indexOf('auto push_arrayproperty(',start);
if(start<0||end<0)throw Error('Unrecognized source: inspect the struct handler manually');
const handler=object.slice(start,end);
const checksRequestedIndex=/is_userdata\(params\.stored_at_index\)/.test(handler);
const readsDefaultIndex=/get_userdata<UScriptStruct>\(\)/.test(handler);
const defaultIndexOne=/get_userdata\(int32_t force_index = 1/.test(wrapper);
const nilTagCanBecomeZero=/lua_getiuservalue\(get_lua_state\(\), force_index, 2\)/.test(wrapper)
 && /userdata_internal_type = get_integer\(-1\)/.test(wrapper)
 && /enum UserdataInternalType[^}]*\{\s*LuaOwnedStackObject\s*,/s.test(wrapper)
 && /lua_tointeger\(get_lua_state\(\), force_index\)/.test(implementation);
console.log(JSON.stringify({
 scope:'offline_source_audit',
 nested_struct_stack_index_mismatch:checksRequestedIndex&&readsDefaultIndex&&defaultIndexOne,
 unchecked_userdata_tag_conversion:nilTagCanBecomeZero,
 native_runtime_verified:false,
 transfer_readiness:'unverified',
 note:'These findings identify hazardous source paths. They do not symbolize the crash or validate a replacement runtime.'
},null,2));
