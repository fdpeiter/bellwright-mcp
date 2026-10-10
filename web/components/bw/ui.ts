// Design tokens for the companion UI (Bellwright-inspired palette).
import type { CSSProperties } from 'react';

export const SERIF = "var(--font-steward-serif), 'Libre Baskerville', serif";
export const SANS = "var(--font-plex-sans), 'IBM Plex Sans', system-ui, sans-serif";
export const MONO = "var(--font-plex-mono), 'IBM Plex Mono', monospace";

export const C = {
  accent: 'var(--accent)',      // #D4B46A
  gold: 'var(--gold)',          // #E3CB92
  pageBg: '#101D19',
  panelBg: '#13231D',
  cardBg: '#1A2B24',
  cardBg2: '#1A160F',
  headBg: '#1D3027',
  inputBg: '#0B1712',
  border: '#2C4438',
  border2: '#344D40',
  border3: '#466052',
  borderRow: '#21362C',
  text: '#E9EADD',
  textBright: '#F1EBDD',
  textSerifBright: '#F7F2E6',
  textDim: '#B4C1B7',
  textDim2: '#A5B6A9',
  textFaint: '#9CAD9F',
  textFainter: '#92A497',
  textDisabled: '#84988A',
  green: '#7DB068',
  greenText: '#8FBF74',
  red: '#C4553B',
  redText: '#EDA593',
  redBar: '#D9614A',
  blue: '#5FA8CE',
  blueText: '#5FB4D6',
} as const;

export const rowBorder: CSSProperties = { borderBottom: `1px solid ${C.borderRow}` };

export const thStyle: CSSProperties = {
  position: 'sticky', top: 0, zIndex: 6, background: C.headBg,
  borderBottom: `1px solid ${C.border3}`, textAlign: 'left', padding: '8px 12px',
  color: '#BDCABC', fontWeight: 600, fontSize: 11, letterSpacing: '.6px',
  textTransform: 'uppercase',
};

export const searchBoxStyle = (h: number): CSSProperties => ({
  display: 'flex', alignItems: 'center', gap: 8, padding: '0 11px', height: h,
  background: C.inputBg, border: '1px solid #3C5145', borderRadius: 8,
});

export const searchInputStyle: CSSProperties = {
  background: 'transparent', border: 'none', outline: 'none', color: C.text,
  fontSize: 13, fontFamily: 'inherit',
};

export const avatarStyle = (size: number, bg: string, radius = 7): CSSProperties => ({
  width: size, height: size, borderRadius: radius, flex: '0 0 auto',
  display: 'flex', alignItems: 'center', justifyContent: 'center',
  fontSize: Math.round(size * 0.38), fontWeight: 600, color: '#0B1712', background: bg,
});

export const badge = (color: string, bg: string, bd: string): CSSProperties => ({
  display: 'inline-flex', alignItems: 'center', gap: 4, padding: '2px 7px',
  borderRadius: 5, fontSize: 10.5, fontWeight: 500, color, background: bg,
  border: `1px solid ${bd}`, whiteSpace: 'nowrap',
});

export const sectionH3: CSSProperties = {
  margin: 0, fontFamily: SERIF, fontSize: 15, fontWeight: 600,
  color: '#E9EADD', letterSpacing: '.3px',
};

export const miniBar = (
  width: number | string, height: number, pct: number, color: string,
): { outer: CSSProperties; inner: CSSProperties } => ({
  outer: {
    width, height, borderRadius: Math.ceil(height / 2),
    background: 'rgba(255,255,255,.08)', overflow: 'hidden',
  },
  inner: { height: '100%', width: `${pct}%`, background: color },
});

export const SEV = {
  alert: { c: '#E0997F', bg: 'rgba(196,85,59,.12)', bd: 'rgba(196,85,59,.4)', a: '#C4553B' },
  warn: { c: '#E8C07A', bg: 'rgba(212,180,106,.1)', bd: 'rgba(212,180,106,.35)', a: '#D4B46A' },
  suggest: { c: '#9AD0E6', bg: 'rgba(95,180,214,.1)', bd: 'rgba(95,180,214,.32)', a: '#5FB4D6' },
  tip: { c: '#A6D08C', bg: 'rgba(125,176,104,.1)', bd: 'rgba(125,176,104,.32)', a: '#7DB068' },
} as const;
