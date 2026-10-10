import { readdir, mkdir, cp } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
const root = fileURLToPath(new URL('.', import.meta.url));
async function check(dir) {
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) await check(path);
    else if (/\.m?js$/.test(entry.name)) {
      const result = spawnSync(process.execPath, ['--check', path], { encoding: 'utf8' });
      if (result.status !== 0) throw new Error(result.stderr || `Syntax check failed: ${entry.name}`);
    }
  }
}
await check(join(root, 'src'));
await mkdir(resolve(root, 'dist'), { recursive: true });
await cp(join(root, 'public'), join(root, 'dist'), { recursive: true });
await cp(join(root, 'src'), join(root, 'dist', 'src'), { recursive: true });
console.log('Steward v2 production assets built.');
