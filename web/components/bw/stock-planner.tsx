'use client';
import { useEffect, useState } from 'react';
import { blankStockAssumption, calculateStock, readStockAssumptions, storedQuantity, type StockAssumption } from '@/lib/bw/stock-recommendation';
const STORAGE_KEY = 'steward.stock-planning-assumptions.v1';
const inputClass = 'w-full rounded border border-line-2 bg-iron-900 p-2 text-sm text-sand-100 focus:outline-2 focus:outline-gold-bright';
export function StockPlanner({ snapshotId }: { snapshotId?: number }) {
  const [rows, setRows] = useState<StockAssumption[]>([blankStockAssumption()]);
  const [hydrated, setHydrated] = useState(false);
  const [storageWarning, setStorageWarning] = useState(false);
  const [world, setWorld] = useState<unknown>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'empty' | 'error'>('loading');
  const [refresh, setRefresh] = useState(0);
  const [retrieved, setRetrieved] = useState('');
  useEffect(() => {
    try { setRows(readStockAssumptions(localStorage.getItem(STORAGE_KEY))); }
    catch { setStorageWarning(true); }
    setHydrated(true);
  }, []);
  useEffect(() => {
    if (!hydrated) return;
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(rows)); }
    catch { setStorageWarning(true); }
  }, [rows, hydrated]);
  useEffect(() => {
    const controller = new AbortController();
    setStatus('loading'); setWorld(null);
    fetch('/api/world', { cache: 'no-store', signal: controller.signal }).then(async response => {
      if (response.status === 404) { if (!controller.signal.aborted) setStatus('empty'); return; }
      if (!response.ok) throw new Error('Stock unavailable');
      const data: unknown = await response.json();
      if (!data || typeof data !== 'object' || Array.isArray(data)) throw new Error('Invalid snapshot');
      if (!controller.signal.aborted) { setWorld(data); setRetrieved(new Date().toLocaleString()); setStatus('ready'); }
    }).catch(() => { if (!controller.signal.aborted) setStatus('error'); });
    return () => controller.abort();
  }, [snapshotId, refresh]);
  const change = (index: number, key: keyof StockAssumption, value: string) => setRows(current => current.map((row, i) => i === index ? { ...row, [key]: value } : row));
  const metadata = world as { snapshot_id?: number; ingested_at?: string } | null;
  const fields: [keyof StockAssumption, string, string][] = [['item', 'Item class', 'Iron_C'], ['rate', 'Use per hour', '2'], ['hours', 'Replenishment hours', '6'], ['safety', 'Safety fraction', '0.25'], ['planned', 'Planned demand', '0']];
  return <section aria-labelledby="stock-planner-title" className="mb-7 border-y border-line-2 py-5">
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div><h2 id="stock-planner-title" className="font-serif text-xl text-sand-100">Stock target planner</h2>
        <p className="mt-1 max-w-3xl text-sm text-sand-400">Calculate the minimum stock for your supplied demand and replenishment assumptions. These estimates do not measure consumption or change game settings.</p></div>
      <button type="button" onClick={() => setRefresh(value => value + 1)} className="rounded border border-line-2 px-3 py-2 text-sm text-sand-100">Refresh stock</button>
    </div>
    <p className="my-3 font-mono text-xs text-sand-400">Target = ceil(use/hour × replenishment hours × (1 + safety fraction) + planned demand)</p>
    <p className="mb-3 text-xs text-sand-400">Only assumptions are remembered in this browser. Stock refreshes and targets recalculate whenever this page reloads.</p>
    {storageWarning && <p role="status" className="mb-3 text-sm text-rust-soft">This browser cannot remember assumptions. You can still calculate targets here.</p>}
    <div role="status" className="mb-4 text-sm text-sand-300">
      {status === 'loading' && 'Refreshing saved stock…'}
      {status === 'empty' && 'No save snapshot yet. Targets can still be calculated; current stock is unknown.'}
      {status === 'error' && <>Stock could not be loaded. <button type="button" onClick={() => setRefresh(value => value + 1)} className="underline">Retry</button></>}
      {status === 'ready' && <>Saved snapshot {typeof metadata?.snapshot_id === 'number' ? metadata.snapshot_id : 'unknown'} · checked {retrieved}{typeof metadata?.ingested_at === 'string' && <> · ingested {metadata.ingested_at}</>}</>}
    </div>
    <p className="mb-4 text-xs text-sand-400">Stored totals cover all parsed world containers, including other villages. Carried and equipped items are excluded. An absent item is counted as zero only when the entire stored aggregate is valid; Missing or invalid quantities stay unknown; valid observed quantities are preserved.</p>
    <div className="space-y-5">
      {rows.map((row, index) => {
        const duplicate = row.item.trim() !== '' && rows.some((other, i) => i !== index && other.item.trim() === row.item.trim());
        const result = calculateStock(row);
        const error = duplicate ? 'Use each item class once.' : result.error;
        const target = error ? null : result.target;
        const current = status === 'ready' && /^[A-Za-z0-9_]{1,256}$/.test(row.item) ? storedQuantity(world, row.item) : null;
        const deficit = current !== null && target !== null ? Math.max(0, target - current) : null;
        return <fieldset key={index} className="border-l-2 border-line-2 pl-3">
          <legend className="mb-2 text-sm text-sand-200">Item {index + 1}</legend>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-5">{fields.map(([key, label, placeholder]) => <label key={key} className="text-xs text-sand-400">{label}<input className={`${inputClass} mt-1`} type={key === 'item' ? 'text' : 'number'} min={key === 'item' ? undefined : key === 'hours' ? '0.000001' : '0'} max={key === 'safety' ? '2' : undefined} step={key === 'planned' ? '1' : 'any'} maxLength={key === 'item' ? 256 : 200} placeholder={placeholder} value={row[key]} onChange={event => change(index, key, event.target.value)} aria-label={`${label}, item ${index + 1}`} /></label>)}</div>
          <div aria-live="polite" className="mt-3 flex flex-wrap items-center gap-x-6 gap-y-2 border-t border-line-2 pt-2 text-sm">
            {target !== null && <span className="text-sand-400">Demand {Number(row.rate) * Number(row.hours)} + safety {Number(row.rate) * Number(row.hours) * Number(row.safety)} + planned {Number(row.planned)} → rounded target</span>}
            <span>Stored <strong className="text-sand-100">{current ?? 'unknown'}</strong></span>
            <span>Suggested target <strong className="text-gold-bright">{target ?? '—'}</strong></span>
            <span>Shortfall <strong className="text-sand-100">{deficit ?? 'unknown'}</strong></span>
            {error && <span className="text-sand-400">{error}</span>}
            <button type="button" disabled={rows.length === 1} onClick={() => setRows(current => current.filter((_, i) => i !== index))} className="ml-auto text-sand-400 underline disabled:opacity-40" aria-label={`Remove item ${index + 1}`}>Remove</button>
          </div>
        </fieldset>;
      })}
    </div>
    <button type="button" disabled={rows.length >= 100 || !hydrated} onClick={() => setRows(current => [...current, blankStockAssumption()])} className="mt-4 rounded border border-line-2 px-3 py-2 text-sm text-sand-100 disabled:opacity-40">Add item ({rows.length}/100)</button>
  </section>;
}


