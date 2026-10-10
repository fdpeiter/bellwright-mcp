// Fixed internal reader: browser traffic stays on Steward's origin.
export const dynamic = 'force-dynamic';
export async function GET() {
  try {
    const r = await fetch('http://127.0.0.1:18126/api/map', {cache:'no-store', signal:AbortSignal.timeout(2500)});
    if (!r.ok) throw new Error('Reader unavailable');
    return Response.json(await r.json(), {headers:{'Cache-Control':'no-store'}});
  } catch {
    return Response.json({status:{stale:true},points:[], unavailable:true}, {headers:{'Cache-Control':'no-store'}});
  }
}
