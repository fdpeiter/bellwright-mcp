'use client';
// Bellwright Steward — app shell (header, tabs, toolbar) + tab bodies.
import { useEffect, useMemo, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import type { Npc, World } from '@/lib/types';
import { agoLabel, playtimeLabel } from '@/lib/bw/format';
import { classifyRole, hireGateOf, insightsFor, npcName, playerNpcs, recruitNpcs, ROLE_COLORS, specialtyOf, upgradeSuggestions, type RecruitRole } from '@/lib/bw/model';
import { shapeContainers } from '@/lib/bw/storage';
import { CompareModal, CompareTray } from './compare';
import { UploadButton } from './upload-modal';
import { RaidWatcher } from './raid-alert';
import { EventsTab } from './events-tab';
import { useRealtime } from './use-realtime';
import { BwSelect } from '@/components/ui/dropdown-menu';
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent } from '@/components/ui/dropdown-menu';
import { Drawer } from './drawer';
import { Insights } from './insights';
import { MapTab } from './map-tab';
import { MeTab } from './me';
import { Roster, guidOf, type SortState } from './roster';
import { SearchIcon } from './icons';
import { StorageTab, storagePressure } from './storage-tab';
import { Trends } from './trends';
import { cn } from '@/lib/utils';
import { useTooltips } from './use-tooltips';

const MOBILE_BREAKPOINT = 760;
const META_WIDE_BREAKPOINT = 1180;
const MAX_COMPARE = 3;
// live refresh: poll the version endpoint and re-render server data when a
// new snapshot lands (router.refresh() keeps all client state — tab, filters,
// open drawer — only the world prop changes)
const VERSION_POLL_MS = 30_000;

type TabKey = 'me' | 'villagers' | 'npcs' | 'trends' | 'insights' | 'storage' | 'map' | 'events';
export type RosterView = 'skills' | 'gear';

// tab <-> URL slug (deep-linkable routes served by app/[tab]/page.tsx)
const TAB_SLUG: Record<TabKey, string> = {
  me: 'me', villagers: 'population', npcs: 'recruits',
  trends: 'trends', insights: 'insights', storage: 'storage', map: 'map', events: 'events',
};
const SLUG_TAB = Object.fromEntries(
  Object.entries(TAB_SLUG).map(([k, s]) => [s, k as TabKey]),
) as Record<string, TabKey>;

export const CompanionApp = ({ world, initialSlug }: { world: World; initialSlug?: string }) => {
  const rootRef = useRef<HTMLDivElement>(null);
  useTooltips(rootRef);
  const router = useRouter();

  // watch for new snapshots and pull fresh server data in place
  useEffect(() => {
    let stopped = false;
    const check = async () => {
      if (stopped || document.hidden) return;
      try {
        const res = await fetch('/api/world/version', { cache: 'no-store' });
        if (!res.ok) return;
        const v = (await res.json()) as { snapshot_id: number | null; ingested_at?: string };
        if (v.snapshot_id != null
          && (v.snapshot_id !== world.snapshot_id || v.ingested_at !== world.ingested_at)) {
          router.refresh();
        }
      } catch { /* endpoint unreachable — try again next tick */ }
    };
    const t = setInterval(check, VERSION_POLL_MS);
    const onVisible = () => { if (!document.hidden) void check(); };
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      stopped = true;
      clearInterval(t);
      document.removeEventListener('visibilitychange', onVisible);
    };
  }, [router, world.snapshot_id, world.ingested_at]);

  const [tab, setTab] = useState<TabKey>(() => {
    const t = SLUG_TAB[initialSlug ?? ''] ?? 'villagers';
    return t === 'me' && !world.player ? 'villagers' : t;
  });
  const [rosterView, setRosterView] = useState<RosterView>('skills');
  const [q, setQ] = useState('');
  const [sort, setSort] = useState<SortState>({ key: 'name', dir: 1 });
  const [selGuid, setSelGuid] = useState<string | null>(null);
  const [compareMode, setCompareMode] = useState(false);
  const [compareSet, setCompareSet] = useState<string[]>([]);
  const [showCompare, setShowCompare] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [villageFilter, setVillageFilter] = useState<Set<string>>(new Set());
  const [profFilter, setProfFilter] = useState<Set<string>>(new Set());
  const [roleFilter, setRoleFilter] = useState<Set<RecruitRole>>(new Set());
  const [rankFilter, setRankFilter] = useState<Set<string>>(new Set());
  const [metaWide, setMetaWide] = useState(true);
  const [ago, setAgo] = useState<string | null>(null); // client-only (avoids hydration mismatch)
  const [realtime, setRealtime] = useRealtime(); // opt-in live monitor (per-browser)

  useEffect(() => {
    const onResize = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
      setMetaWide(window.innerWidth >= META_WIDE_BREAKPOINT);
    };
    onResize();
    window.addEventListener('resize', onResize);
    const tick = () => setAgo(agoLabel(world.ingested_at));
    tick();
    const t = setInterval(tick, 30_000);
    return () => { window.removeEventListener('resize', onResize); clearInterval(t); };
  }, [world.ingested_at]);

  useEffect(() => {
    setVillageFilter(new Set()); setProfFilter(new Set()); setRoleFilter(new Set()); setRankFilter(new Set());
  }, [tab]);

  // switch tab + reflect it in the URL so refresh/back keep the active tab
  const openTab = (k: TabKey) => {
    setTab(k);
    window.history.pushState(null, '', `/${TAB_SLUG[k]}`);
  };
  useEffect(() => {
    const onPop = () => {
      const t = SLUG_TAB[window.location.pathname.replace(/^\//, '')] ?? 'villagers';
      setTab(t === 'me' && !world.player ? 'villagers' : t);
    };
    window.addEventListener('popstate', onPop);
    return () => window.removeEventListener('popstate', onPop);
  }, [world.player]);

  const villagers = useMemo(() => playerNpcs(world), [world]);
  const recruits = useMemo(() => recruitNpcs(world), [world]);
  const insights = useMemo(() => insightsFor(villagers), [villagers]);
  const upgrades = useMemo(
    () => upgradeSuggestions(villagers, recruits, world.villages),
    [villagers, recruits, world.villages]);
  const containers = useMemo(() => shapeContainers(world.storage), [world.storage]);
  const storageAlert = useMemo(() => storagePressure(containers), [containers]);
  // homelessness: villagers beyond total sleeping quarters (aggregate deficit)
  const homeless = world.housing ? Math.max(0, villagers.length - world.housing.quarters) : 0;
  const villagerAlert = villagers.some(v => v.injuries.length > 0) || homeless > 0;

  const npcByRawGuid = useMemo(() => {
    const m = new Map<string, Npc>();
    for (const v of world.npcs) if (v.guid) m.set(v.guid, v);
    return m;
  }, [world.npcs]);

  const squadsOfSel = (guid: string | null) =>
    guid == null ? [] : (world.groups ?? [])
      .filter(g => g.name && g.members.includes(guid))
      .map(g => ({
        name: g.name as string,
        members: g.members.map(m => npcByRawGuid.get(m)).filter((x): x is Npc => Boolean(x)),
      }));

  const byGuid = useMemo(() => {
    const m = new Map<string, Npc>();
    for (const v of world.npcs) m.set(guidOf(v), v);
    return m;
  }, [world.npcs]);

  const isTable = tab === 'villagers' || tab === 'npcs';
  const dataset = tab === 'npcs' ? recruits : villagers;
  const ql = q.trim().toLowerCase();
  const villages = useMemo(() =>
    [...new Set(dataset.map(v => v.village).filter((x): x is string => Boolean(x)))].sort(), [dataset]);
  const professions = useMemo(() =>
    [...new Set(dataset.map(v => specialtyOf(v)).filter((x): x is string => Boolean(x)))].sort(), [dataset]);
  // profession specialties are liberation unlocks; Commoner/Beggar hire on
  // village Trust — split the chip groups to mirror the game's system
  const specialistChips = useMemo(() => professions.filter(p => p !== 'Commoner' && p !== 'Beggar'), [professions]);
  const commonChips = useMemo(() => professions.filter(p => p === 'Commoner' || p === 'Beggar'), [professions]);
  // trust ranks present in the dataset, in ladder order (exact hire-gate data)
  const RANK_ORDER = ['Stranger', 'Associate', 'Friend', 'Protector', 'Leader'];
  const ranks = useMemo(() => {
    const present = new Set(dataset.map(v => hireGateOf(v)?.trust).filter((x): x is string => Boolean(x)));
    return RANK_ORDER.filter(r => present.has(r));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dataset]);
  const rows = useMemo(() => {
    let list = dataset.filter(v => !ql
      || npcName(v).toLowerCase().includes(ql)
      || (v.template ?? '').toLowerCase().includes(ql)
      || (v.village ?? '').toLowerCase().includes(ql));
    if (villageFilter.size) list = list.filter(v => v.village != null && villageFilter.has(v.village));
    if (profFilter.size) { const p = specialtyOf; list = list.filter(v => { const x = p(v); return x != null && profFilter.has(x); }); }
    if (roleFilter.size) list = list.filter(v => roleFilter.has(classifyRole(v)));
    if (rankFilter.size) list = list.filter(v => { const g = hireGateOf(v); return g != null && rankFilter.has(g.trust); });
    const { key: k, dir } = sort;
    list = [...list].sort((a, b) => {
      if (k === 'name') {
        const av = `${a.last_name ?? ''}${a.first_name ?? ''}`, bv = `${b.last_name ?? ''}${b.first_name ?? ''}`;
        return av < bv ? -dir : av > bv ? dir : 0;
      }
      let av: number, bv: number;
      if (k === 'morale') { av = a.morale ?? -1; bv = b.morale ?? -1; }
      else {
        av = a.skills[k]?.level ?? -1; bv = b.skills[k]?.level ?? -1;
        if (av === bv) { av = a.skills[k]?.xp ?? 0; bv = b.skills[k]?.xp ?? 0; }
      }
      return (av - bv) * dir;
    });
    return list;
  }, [dataset, ql, sort, tab, villageFilter, profFilter, roleFilter, rankFilter]);

  const onSort = (key: string, defDir: 1 | -1) =>
    setSort(s => (s.key === key ? { key, dir: s.dir === 1 ? -1 : 1 } : { key, dir: defDir }));

  const toggleCompare = (g: string) =>
    setCompareSet(s => (s.includes(g) ? s.filter(x => x !== g) : s.length < MAX_COMPARE ? [...s, g] : s));

  const compareNpcs = compareSet.map(g => byGuid.get(g)).filter((v): v is Npc => Boolean(v));
  const sel = selGuid ? byGuid.get(selGuid) ?? null : null;

  const tabs: { key: TabKey; label: string; count: number | null; alert: boolean }[] = [
    ...(world.player ? [{ key: 'me' as TabKey, label: 'My character', count: null, alert: false }] : []),
    { key: 'villagers', label: 'Population', count: villagers.length, alert: villagerAlert },
    { key: 'npcs', label: 'Recruits', count: recruits.length, alert: false },
    { key: 'trends', label: 'Trends', count: null, alert: false },
    { key: 'insights', label: 'Insights', count: insights.length, alert: homeless > 0 },
    { key: 'storage', label: 'Storage', count: containers.length, alert: storageAlert },
    { key: 'map', label: 'Map', count: null, alert: false },
    ...(realtime ? [{ key: 'events' as TabKey, label: 'Events', count: null, alert: false }] : []),
  ];

  return (
    <div ref={rootRef}
      className="steward-shell bg-iron-900 text-sand-200 h-full max-w-[100vw] overflow-hidden flex flex-col font-sans antialiased"
      >
      {realtime && <RaidWatcher />}
      {/* header */}
      <header className="flex items-center gap-2 sm:gap-4.5 min-h-14 py-2 px-3 sm:px-4.5 min-w-0 border-b border-line-2 bg-gradient-to-b from-[#1A2B24] to-[#13231D] flex-none relative z-20">
        <div className="steward-brand">
          {/* Original approved logo 04 + lettering 08; ivory field preserves contrast. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/brand/steward-lockup.png" alt="Bellwright Steward" width={200} height={48} />
        </div>
        <div className="hidden lg:block w-px h-6.5 bg-line-3" />
        <div className="hidden lg:flex items-center gap-3.5 text-[11.5px] md:text-[12.5px] flex-1 min-w-0 flex-wrap max-h-10 overflow-hidden">
          <Meta k="save" v={world.meta.saveName ?? '—'} />
          <Meta k="build" v={world.meta.savedBuild ?? '—'} mono />
          {metaWide && <Meta k="region" v={world.meta.region ?? '—'} />}
          {metaWide && <Meta k="played" v={playtimeLabel(world.meta.playtimeSeconds)} mono />}
        </div>
        <div className="hidden xl:flex items-center gap-2 py-1.25 px-2.75 rounded-full bg-white/[.03] border border-line-3 flex-none">
          <span className="w-1.75 h-1.75 rounded-full bg-sand-500" />
          <span className="text-[11.5px] md:text-[12.5px] text-sand-400 whitespace-nowrap">
            {ago ? `Imported ${ago}` : 'Save snapshot'}
          </span>
        </div>
        <UploadButton onIngested={() => router.refresh()} />
        <SettingsMenu realtime={realtime} onRealtime={setRealtime} />
      </header>

      <details className="flex-none border-b border-line bg-iron-800 px-3 sm:px-4.5 py-1.5 text-[11px] md:text-[12px] text-sand-400 leading-relaxed">
        <summary className="cursor-pointer">Save snapshot · {ago ? `Imported ${ago}` : 'Import time unavailable'} · Source details</summary>
        <div className="py-2 space-y-1">
          <div>Updates after save ingestion. Game capture time is unavailable; import time does not establish current game state.</div>
          <div>Imported: <time dateTime={world.ingested_at}>{world.ingested_at}</time> · Save: {world.meta.saveName ?? 'Unknown'} · Build: {world.meta.savedBuild ?? 'Unknown'}</div>
          <div>Map pins identify save or telemetry sources. Wiki resources are static locations; current availability and yield are unknown.</div>
        </div>
      </details>
      <div className="flex flex-1 min-h-0">
        <aside className="steward-rail hidden lg:flex flex-col flex-none p-4">
          <div className="text-xs uppercase tracking-widest text-sand-500 mb-1">Settlement ledger</div>
          <div className="font-serif text-xl text-sand-100 mb-5">{world.meta.region ?? 'Your settlement'}</div>
          <nav aria-label="Main navigation" className="flex flex-col gap-1">
            {tabs.map(t => <button key={t.key} onClick={() => openTab(t.key)} aria-current={tab === t.key ? 'page' : undefined}
              className={cn('flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-left cursor-pointer', tab === t.key ? 'bg-iron-650 text-gold-bright' : 'text-sand-400 hover:bg-iron-750 hover:text-sand-100')}>
              <span className="flex-1">{t.label}</span>
              {t.count != null && <span className="font-mono text-xs text-sand-500">{t.count}</span>}
              {t.alert && <span className="text-rust-soft" aria-label="Needs attention">!</span>}
            </button>)}
          </nav>
          <div className="mt-auto pt-4 text-xs text-sand-500 leading-relaxed">Read-only settlement insights<br />Built on <a href="https://github.com/claygorman/bellwright-companion" className="underline hover:text-sand-100">Bellwright Companion</a></div>
        </aside>
        <div className="flex flex-col flex-1 min-w-0 min-h-0">
      {/* tabs + toolbar */}
      <div className="flex items-center gap-1.5 py-2.25 px-4.5 border-b border-line bg-iron-800 flex-none flex-wrap relative z-[15]">
        <label className="sm:hidden flex items-center gap-2 w-full text-xs text-sand-400">View
          <select aria-label="View section" value={tab} onChange={e => openTab(e.target.value as TabKey)} className="flex-1 min-w-0 rounded-lg border border-line-3 bg-ink p-2 text-sand-100 text-sm">
            {tabs.map(t => <option key={t.key} value={t.key}>{t.label}{t.count != null ? ` (${t.count})` : ''}</option>)}
          </select>
        </label>
        <nav aria-label="Settlement sections" className="bw-scroll hidden sm:flex lg:hidden gap-0.75 bg-ink p-0.75 rounded-[9px] border border-[#344D40] max-w-full overflow-x-auto flex-nowrap">
          {tabs.map(t => {
            const active = tab === t.key;
            return (
              <button aria-current={tab === t.key ? 'page' : undefined} key={t.key} onClick={() => openTab(t.key)} className={cn(
                'inline-flex items-center gap-1.75 py-1.5 px-3.25 border-none flex-none whitespace-nowrap rounded-[7px] cursor-pointer font-sans text-[13px] md:text-[14.5px] transition-all duration-100',
                active ? 'font-semibold text-[#101D19] bg-gold' : 'font-medium text-sand-400 bg-transparent hover:text-[#F1EBDD] hover:bg-white/[.03]',
              )}>
                <span>{t.label}</span>
                {t.count != null && (
                  <span className={cn(
                    'font-mono text-[11px] md:text-[12px] py-px px-1.5 rounded-[5px]',
                    active ? 'text-[#101D19] bg-black/[.16]' : 'text-sand-400 bg-[#21362C]',
                  )}>{t.count}</span>
                )}
                {t.alert && (
                  <span data-tip={t.key === 'storage' ? 'Storage needs attention'
                    : t.key === 'insights' ? `${homeless} villager${homeless === 1 ? '' : 's'} without a bed`
                    : 'Villagers need attention'}
                    className="inline-flex items-center justify-center w-3.75 h-3.75 rounded-full bg-rust text-white text-[10px] md:text-[11px] font-bold leading-none">!</span>
                )}
              </button>
            );
          })}
        </nav>
        <div className="flex-1" />
        {isTable && (
          <div className="flex items-center gap-2.5 flex-wrap w-full sm:w-auto">
            <div className="flex items-center gap-2 px-2.75 h-8.5 bg-ink border border-line-3 rounded-lg flex-1 min-w-0 sm:flex-none">
              <SearchIcon />
              <input value={q} onChange={e => setQ(e.target.value)}
                placeholder="Filter by name or archetype…"
                className="bg-transparent border-none outline-none text-sand-200 text-[13px] md:text-[14.5px] w-full sm:w-47.5 min-w-0 font-sans" />
            </div>
            <BwSelect label="View" value={rosterView} align="start"
              options={[{ value: 'skills', label: 'Skills' }, { value: 'gear', label: 'Gear & inventory' }]}
              onChange={v => setRosterView(v as RosterView)} />
            <button onClick={() => setCompareMode(m => { if (m) setCompareSet([]); return !m; })} className={cn(
              'inline-flex items-center gap-1.75 h-8.5 px-3.25 rounded-lg cursor-pointer font-sans text-[12.5px] md:text-[14px] font-medium border transition-colors',
              compareMode ? 'border-gold bg-gold/[.14] text-gold-bright' : 'border-line-3 bg-ink text-sand-400 hover:border-[#516C5B]',
            )}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                <path d="M9 3v18M15 3v18M3 9h18" />
              </svg>
              <span>Compare</span>
            </button>
            <span className="hidden sm:inline text-xs md:text-[13px] text-sand-600 tabular-nums whitespace-nowrap">
              {rows.length} {tab === 'npcs' ? 'recruits' : 'villagers'}
            </span>
          </div>
        )}
      </div>

      {isTable && (
        <div className="bw-scroll flex items-center gap-1.5 py-1.75 px-4.5 flex-nowrap overflow-x-auto sm:flex-wrap sm:overflow-visible border-b border-line bg-[#101D19] text-[11.5px] md:text-[12.5px]">
          {villages.length > 0 && (<>
            <FilterChips label="Village" all={villages} sel={villageFilter}
              onToggle={v => setVillageFilter(s2 => { const n = new Set(s2); n.has(v) ? n.delete(v) : n.add(v); return n; })} />
            <span className="w-px h-4 bg-line-3 flex-none" />
          </>)}
          {specialistChips.length > 0 && (<>
            <FilterChips label="Specialist (liberated)" all={specialistChips} sel={profFilter}
              onToggle={v => setProfFilter(s2 => { const n = new Set(s2); n.has(v) ? n.delete(v) : n.add(v); return n; })} />
            <span className="w-px h-4 bg-line-3 flex-none" />
          </>)}
          {commonChips.length > 0 && (<>
            <FilterChips label="Common" all={commonChips} sel={profFilter}
              onToggle={v => setProfFilter(s2 => { const n = new Set(s2); n.has(v) ? n.delete(v) : n.add(v); return n; })} />
            <span className="w-px h-4 bg-line-3 flex-none" />
          </>)}
          {ranks.length > 0 && (<>
            <FilterChips label="Trust" all={ranks} sel={rankFilter}
              onToggle={v => setRankFilter(s2 => { const n = new Set(s2); n.has(v) ? n.delete(v) : n.add(v); return n; })} />
            <span className="w-px h-4 bg-line-3 flex-none" />
          </>)}
          <FilterChips label="Role" all={['Fighter', 'Worker', 'Balanced']} sel={roleFilter as Set<string>}
            colors={ROLE_COLORS as Record<string, string>}
            onToggle={v => setRoleFilter(s2 => { const n = new Set(s2); n.has(v as RecruitRole) ? n.delete(v as RecruitRole) : n.add(v as RecruitRole); return n; })} />
        </div>
      )}

      <div className="flex-none px-4 sm:px-6 py-3 border-b border-line">
        <div className="text-[11px] uppercase tracking-widest text-sand-500">{world.meta.character ?? 'Player'} / {world.meta.region ?? 'Settlement'}</div>
        <h1 className="font-serif text-xl sm:text-2xl text-sand-100 mt-1">{tab === 'villagers' ? 'People of your settlement' : tab === 'npcs' ? 'Find your next villager' : tabs.find(t => t.key === tab)?.label ?? 'Events'}</h1>
        {isTable && <p className="text-xs text-sand-400 mt-1">Save-backed roster · Portraits use saved appearance tokens until real screenshots are supplied.</p>}
      </div>
      {/* main */}
      {/* fixed content pane — each tab body is the single scroll region (h-full
          overflow-y-auto) so scrolling is consistent and never double-scrolls */}
      <main className="flex-1 overflow-hidden relative min-h-0">
        {tab === 'me' && world.player && (
          <MeTab player={world.player} meta={world.meta} villagerCount={villagers.length}
            presetName={null} />
        )}
        {isTable && (
          <Roster rows={rows} npcCol={tab === 'npcs'} archCol playtime={world.meta.playtimeSeconds} ingestedAt={world.ingested_at}
            isMobile={isMobile} view={rosterView} carried={world.carried}
            compareMode={compareMode} compareSet={compareSet} sort={sort} onSort={onSort}
            onOpen={setSelGuid} onToggleCompare={toggleCompare} />
        )}
        {tab === 'trends' && <Trends />}
        {tab === 'insights' && <Insights cards={insights} upgrades={upgrades} villages={world.villages ?? []} housing={world.housing} villagerCount={villagers.length} snapshotId={world.snapshot_id} onOpen={setSelGuid} />}
        {tab === 'storage' && <StorageTab containers={containers} />}
        {tab === 'map' && (
          <MapTab world={world} realtime={realtime}
            region={`${world.meta.map?.startsWith('Karvenia') ? 'Karvenia' : world.meta.map ?? ''} — ${world.meta.region ?? ''}`}
            onOpenProfile={setSelGuid} />
        )}
        {tab === 'events' && (realtime
          ? <EventsTab />
          : (
            <div className="flex flex-col items-center justify-center h-full text-center px-6 gap-3">
              <div className="font-serif text-lg md:text-[20px] text-sand-200">Realtime monitor is off</div>
              <p className="text-[13px] md:text-[14.5px] text-sand-600 max-w-105">
                The Events feed and live map need the realtime monitor, which reads from the
                companion daemon. Enable it to view reader status and the live event log.
              </p>
              <button onClick={() => setRealtime(true)}
                className="mt-1 inline-flex items-center gap-2 h-8.5 px-3.5 rounded-lg border border-gold bg-gold/[.14] text-gold-bright text-[12.5px] md:text-[14px] font-medium cursor-pointer">
                Enable realtime monitor
              </button>
            </div>
          ))}
      </main>
        </div>
      </div>

      {compareMode && compareNpcs.length > 0 && (
        <CompareTray npcs={compareNpcs} onRemove={toggleCompare}
          onClear={() => setCompareSet([])} onOpen={() => setShowCompare(true)} />
      )}
      {sel && <Drawer v={sel} playtime={world.meta.playtimeSeconds} ingestedAt={world.ingested_at} squads={squadsOfSel(sel.guid)}
        presetName={sel.gear_preset ? (world.gear_presets ?? []).find(p => p.key === sel.gear_preset)?.name ?? null : null}
        onOpen={setSelGuid} isMobile={isMobile} onClose={() => setSelGuid(null)} />}
      {showCompare && <CompareModal npcs={compareNpcs} onClose={() => setShowCompare(false)} />}
    </div>
  );
};

const FilterChips = ({ label, all, sel, onToggle, colors }: {
  label: string; all: string[]; sel: Set<string>;
  onToggle: (v: string) => void; colors?: Record<string, string>;
}) => (
  <span className="inline-flex items-center gap-1.25 flex-none flex-nowrap sm:flex-wrap">
    <span className="text-[9.5px] md:text-[10.5px] tracking-[.5px] uppercase text-sand-600 whitespace-nowrap">{label}</span>
    {all.map(v => {
      const on = sel.has(v);
      const color = colors?.[v] ?? '#C8D2C7';
      return (
        <button key={v} onClick={() => onToggle(v)}
          className={cn(
            'py-0.5 px-2.25 rounded-full cursor-pointer font-sans text-[11px] md:text-[12px] border flex-none whitespace-nowrap',
            on ? 'text-[#EDE4D2]' : 'border-line-3 bg-transparent text-sand-400',
          )}
          // active chip tint comes from the role/village color in data
          style={on ? { borderColor: color, background: `${color}22` } : undefined}>{v}</button>
      );
    })}
  </span>
);

// Header settings menu — houses the per-browser Realtime monitor toggle.
const SettingsMenu = ({ realtime, onRealtime }: { realtime: boolean; onRealtime: (v: boolean) => void }) => (
  <DropdownMenu>
    <DropdownMenuTrigger
      aria-label="Settings"
      data-tip="Settings"
      className="inline-flex items-center justify-center w-8.5 h-8.5 rounded-lg border border-line-3 bg-ink text-sand-400 hover:border-[#516C5B] cursor-pointer outline-none flex-none">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.8}>
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end" className="w-70 p-2.5">
      <button onClick={() => onRealtime(!realtime)}
        className="flex items-start gap-2.5 w-full text-left rounded-md p-2 hover:bg-white/[.04] cursor-pointer">
        <span className={cn('mt-0.5 w-8.5 h-4.75 rounded-full flex-none relative transition-colors',
          realtime ? 'bg-gold' : 'bg-[#3C5145]')}>
          <span className={cn('absolute top-0.5 w-3.75 h-3.75 rounded-full bg-white transition-all',
            realtime ? 'left-4.25' : 'left-0.5')} />
        </span>
        <span className="min-w-0">
          <span className="block text-[12.5px] md:text-[14px] text-sand-100 font-medium">Realtime monitor</span>
          <span className="block text-[11px] md:text-[12px] text-sand-500 leading-snug">
            Live map pins, raid alerts &amp; the Events tab — needs the companion reader daemon running.
          </span>
        </span>
      </button>
      <div className="mt-1 pt-2 border-t border-line-2 px-2 flex items-center justify-between">
        <span className="text-[10.5px] md:text-[11.5px] text-sand-600">Steward · based on Companion</span>
        <span className="font-mono text-[10.5px] md:text-[11.5px] text-sand-500">v{process.env.NEXT_PUBLIC_APP_VERSION ?? '—'}</span>
      </div>
    </DropdownMenuContent>
  </DropdownMenu>
);

const Meta = ({ k, v, mono = false }: { k: string; v: string; mono?: boolean }) => (
  <span className="inline-flex gap-1.5 text-sand-400 whitespace-nowrap">
    <span className="text-sand-700">{k}</span>
    <span className={mono ? 'font-mono' : undefined}>{v}</span>
  </span>
);
