export type StockAssumption = { item: string; rate: string; hours: string; safety: string; planned: string };
export const blankStockAssumption = (): StockAssumption => ({ item: '', rate: '', hours: '', safety: '0.25', planned: '0' });
export function readStockAssumptions(raw: string | null): StockAssumption[] {
  try {
    const value: unknown = JSON.parse(raw ?? 'null');
    if (!Array.isArray(value) || value.length < 1 || value.length > 100) return [blankStockAssumption()];
    if (!value.every(row => row && typeof row === 'object' && ['item', 'rate', 'hours', 'safety', 'planned'].every(key => typeof row[key] === 'string' && row[key].length <= (key === 'item' ? 256 : 200)))) return [blankStockAssumption()];
    return value.map(({ item, rate, hours, safety, planned }) => ({ item, rate, hours, safety, planned }));
  } catch { return [blankStockAssumption()]; }
}
export function calculateStock(row: StockAssumption): { target: number | null; error: string | null } {
  if (!/^[A-Za-z0-9_]{1,256}$/.test(row.item)) return { target: null, error: 'Enter an exact item class using letters, numbers and underscores.' };
  const values = [row.rate, row.hours, row.safety, row.planned];
  if (values.some(value => !value.trim())) return { target: null, error: 'Enter all planning assumptions.' };
  const [rate, hours, safety, planned] = values.map(Number);
  if ([rate, hours, safety, planned].some(value => !Number.isFinite(value) || value < 0 || value > Number.MAX_SAFE_INTEGER) || hours <= 0 || safety > 2 || !Number.isSafeInteger(planned)) return { target: null, error: 'Use nonnegative values, positive hours, safety from 0 to 2, and whole planned demand.' };
  const demand = rate * hours;
  const safetyStock = demand * safety;
  const buffered = demand * (1 + safety);
  const sum = buffered + planned;
  const target = Math.ceil(sum);
  if ([demand, safetyStock, buffered, sum].some(value => !Number.isFinite(value) || value > Number.MAX_SAFE_INTEGER) || !Number.isSafeInteger(target)) return { target: null, error: 'Target exceeds the supported quantity range.' };
  return { target, error: null };
}
export function storedQuantity(world: unknown, item: string): number | null {
  if (!world || typeof world !== 'object') return null;
  const totals = (world as { storage?: { totals?: unknown } }).storage?.totals;
  if (!totals || typeof totals !== 'object' || Array.isArray(totals)) return null;
  const entries = Object.entries(totals);
  const valid = (qty: unknown): qty is number => typeof qty === 'number' && Number.isSafeInteger(qty) && qty >= 0;
  if (Object.hasOwn(totals, item)) {
    const qty: unknown = (totals as Record<string, unknown>)[item];
    return valid(qty) ? qty : null;
  }
  return entries.every(([, qty]) => valid(qty)) ? 0 : null;
}



