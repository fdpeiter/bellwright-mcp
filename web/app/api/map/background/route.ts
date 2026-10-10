// Fixed private game asset only; never accept a path or URL from a request.
import { readFile, stat } from 'node:fs/promises';

export const runtime = 'nodejs';
export const dynamic = 'force-dynamic';
const MAX_BYTES = 64 * 1024 * 1024;

export async function GET(request: Request) {
  const file = process.env.MAP_BACKGROUND_FILE;
  if (!file) return Response.json({ error: 'Private terrain image is not configured.' }, { status: 503 });
  try {
    const info = await stat(file);
    if (!info.isFile() || info.size < 24 || info.size > MAX_BYTES) throw new Error('Invalid asset size');
    const etag = `"terrain-${info.size}-${Math.floor(info.mtimeMs)}"`;
    const headers = { 'Content-Type': 'image/png', 'Cache-Control': 'private, max-age=3600', ETag: etag, 'X-Content-Type-Options': 'nosniff' };
    if (request.headers.get('if-none-match') === etag) return new Response(null, { status: 304, headers });
    const image = await readFile(file);
    if (image.subarray(0, 8).toString('hex') !== '89504e470d0a1a0a' || image.readUInt32BE(16) !== 4096 || image.readUInt32BE(20) !== 4096) throw new Error('Expected Karvenia square PNG');
    return new Response(new Uint8Array(image), { headers });
  } catch {
    return Response.json({ error: 'Private terrain image is unavailable or invalid.' }, { status: 503 });
  }
}

