import { h, Fragment } from "preact";
import { useMemo, useRef, useState } from "preact/hooks";
import { api, RequestLog } from "@/lib/api";
import { cn } from "@/lib/utils";

type Range = "5m" | "15m" | "1h" | "6h" | "24h" | "all";

interface AnalyticsProps {
  requests: RequestLog[];
  onClearAll: () => void;
  onImport: (logs: RequestLog[]) => void;
}

// Bucket sizes (ms) for the throughput chart at each time range. Tuned so
// each chart ends up with roughly 30-60 buckets — enough resolution to look
// like a chart, few enough that bars don't shimmer at sub-pixel widths.
const bucketMs: Record<Range, number> = {
  "5m": 10_000, // 10s
  "15m": 30_000, // 30s
  "1h": 120_000, // 2m
  "6h": 600_000, // 10m
  "24h": 1_800_000, // 30m
  all: 0, // computed dynamically from data span
};

const rangeMs: Record<Range, number> = {
  "5m": 5 * 60_000,
  "15m": 15 * 60_000,
  "1h": 60 * 60_000,
  "6h": 6 * 60 * 60_000,
  "24h": 24 * 60 * 60_000,
  all: Number.POSITIVE_INFINITY,
};

export function Analytics({ requests, onClearAll, onImport }: AnalyticsProps) {
  const [range, setRange] = useState<Range>("15m");

  const filtered = useMemo(() => {
    if (range === "all") return requests;
    const cutoff = Date.now() - rangeMs[range];
    return requests.filter((r) => new Date(r.Timestamp).getTime() >= cutoff);
  }, [requests, range]);

  const summary = useMemo(() => summarize(filtered), [filtered]);

  return (
    <main class="flex-1 overflow-auto">
      <header class="px-8 pt-6 pb-4 flex items-start justify-between gap-4 border-b border-zinc-200 bg-white">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Analytics</h1>
          <p class="text-sm text-zinc-500 mt-1">
            Throughput, latency distribution, and per-endpoint breakdown.
          </p>
        </div>
        <div class="flex items-center gap-3">
          <div class="flex items-center gap-1 bg-zinc-100 rounded-md p-0.5">
            {(["5m", "15m", "1h", "6h", "24h", "all"] as const).map((r) => (
              <button
                key={r}
                onClick={() => setRange(r)}
                class={cn(
                  "text-xs px-2.5 py-1 rounded",
                  range === r
                    ? "bg-white text-zinc-900 shadow-sm font-medium"
                    : "text-zinc-500 hover:text-zinc-900"
                )}
              >
                {r}
              </button>
            ))}
          </div>
          <DataActions requests={filtered} onImport={onImport} />
          <button
            onClick={onClearAll}
            class="text-xs text-red-700 border border-red-200 rounded-md px-2.5 py-1.5 hover:bg-red-50"
          >
            Clear all
          </button>
        </div>
      </header>

      <div class="px-8 py-6 space-y-6">
        {filtered.length === 0 ? (
          <EmptyState range={range} />
        ) : (
          <Fragment>
            <HeroStats s={summary} />
            <div class="grid grid-cols-3 gap-4">
              <ThroughputCard requests={filtered} range={range} />
              <StatusBreakdown s={summary} />
            </div>
            <LatencyHistogram requests={filtered} />
            <EndpointsTable requests={filtered} />
          </Fragment>
        )}
      </div>
    </main>
  );
}

// ────────────────────────────────────────────────────────────────────────
// Summary
// ────────────────────────────────────────────────────────────────────────

interface Summary {
  total: number;
  twoXX: number;
  threeXX: number;
  fourXX: number;
  fiveXX: number;
  errorRate: number;
  p50: number;
  p95: number;
  p99: number;
  max: number;
  rps: number;
  windowSec: number;
}

function summarize(reqs: RequestLog[]): Summary {
  const total = reqs.length;
  if (total === 0) {
    return {
      total: 0, twoXX: 0, threeXX: 0, fourXX: 0, fiveXX: 0, errorRate: 0,
      p50: 0, p95: 0, p99: 0, max: 0, rps: 0, windowSec: 0,
    };
  }
  let twoXX = 0, threeXX = 0, fourXX = 0, fiveXX = 0;
  for (const r of reqs) {
    if (r.StatusCode >= 200 && r.StatusCode < 300) twoXX++;
    else if (r.StatusCode < 400) threeXX++;
    else if (r.StatusCode < 500) fourXX++;
    else fiveXX++;
  }
  const sorted = reqs.map((r) => r.Duration).sort((a, b) => a - b);
  const pct = (p: number) => sorted[Math.min(sorted.length - 1, Math.floor(sorted.length * p))];

  // Throughput is total / window. The window is min(time-since-oldest, 1s) so
  // a single request doesn't blow up to Infinity rps.
  const timestamps = reqs.map((r) => new Date(r.Timestamp).getTime());
  const span = (Math.max(...timestamps) - Math.min(...timestamps)) / 1000;
  const windowSec = Math.max(span, 1);
  return {
    total, twoXX, threeXX, fourXX, fiveXX,
    errorRate: (fourXX + fiveXX) / total,
    p50: pct(0.5),
    p95: pct(0.95),
    p99: pct(0.99),
    max: sorted[sorted.length - 1],
    rps: total / windowSec,
    windowSec,
  };
}

// ────────────────────────────────────────────────────────────────────────
// Hero stats row
// ────────────────────────────────────────────────────────────────────────

function HeroStats({ s }: { s: Summary }) {
  return (
    <section class="grid grid-cols-6 gap-3">
      <Metric label="Total" value={s.total.toLocaleString()} sub={`${s.rps.toFixed(2)} rps`} accent="dark" />
      <Metric label="Error rate" value={`${(s.errorRate * 100).toFixed(1)}%`} sub={`${s.fourXX + s.fiveXX} of ${s.total}`} accent={s.errorRate > 0.05 ? "red" : "default"} />
      <Metric label="p50" value={formatMs(s.p50)} />
      <Metric label="p95" value={formatMs(s.p95)} accent={s.p95 > 500 ? "amber" : "default"} />
      <Metric label="p99" value={formatMs(s.p99)} accent={s.p99 > 1000 ? "amber" : "default"} />
      <Metric label="Max" value={formatMs(s.max)} />
    </section>
  );
}

function Metric({
  label, value, sub, accent,
}: {
  label: string;
  value: string;
  sub?: string;
  accent?: "default" | "dark" | "red" | "amber";
}) {
  const valueColor =
    accent === "dark"
      ? "text-zinc-900"
      : accent === "red"
      ? "text-red-700"
      : accent === "amber"
      ? "text-amber-700"
      : "text-zinc-900";
  return (
    <div class="bg-white border border-zinc-200 rounded-xl p-4">
      <div class="text-[11px] uppercase tracking-wide text-zinc-500 mb-1">{label}</div>
      <div class={cn("text-2xl font-semibold tabular-nums", valueColor)}>{value}</div>
      {sub && <div class="text-[11px] text-zinc-500 mt-1">{sub}</div>}
    </div>
  );
}

function formatMs(n: number): string {
  if (!isFinite(n)) return "—";
  if (n < 1) return "<1ms";
  if (n < 1000) return `${Math.round(n)}ms`;
  return `${(n / 1000).toFixed(2)}s`;
}

// ────────────────────────────────────────────────────────────────────────
// Throughput area chart
// ────────────────────────────────────────────────────────────────────────

function ThroughputCard({
  requests, range,
}: {
  requests: RequestLog[];
  range: Range;
}) {
  const data = useMemo(() => buildThroughput(requests, range), [requests, range]);
  return (
    <section class="bg-white border border-zinc-200 rounded-xl p-5 col-span-2 flex flex-col">
      <header class="flex items-center justify-between mb-4">
        <div>
          <h3 class="text-sm font-semibold">Throughput</h3>
          <p class="text-xs text-zinc-500 mt-0.5">Requests per bucket · errors overlaid</p>
        </div>
        <Legend />
      </header>
      <div class="flex-1 min-h-[240px]">
        <ThroughputSvg data={data} />
      </div>
    </section>
  );
}

interface ThroughputBucket {
  t: number;
  total: number;
  errors: number;
}

function buildThroughput(reqs: RequestLog[], range: Range): ThroughputBucket[] {
  if (reqs.length === 0) return [];
  const times = reqs.map((r) => new Date(r.Timestamp).getTime());
  const minT = Math.min(...times);
  const maxT = Math.max(...times);

  let bw = bucketMs[range];
  if (bw === 0) {
    // "all" — divide observed span into ~40 buckets, min 1s.
    bw = Math.max(1000, Math.ceil((maxT - minT) / 40));
  }
  // Snap start/end to bucket boundaries so the chart starts at 0 rather than
  // at a random partial bucket.
  const start = Math.floor(minT / bw) * bw;
  const end = Math.ceil((maxT + 1) / bw) * bw;
  const count = Math.max(1, Math.min(120, Math.round((end - start) / bw)));
  const buckets: ThroughputBucket[] = Array.from({ length: count }, (_, i) => ({
    t: start + i * bw,
    total: 0,
    errors: 0,
  }));
  for (const r of reqs) {
    const idx = Math.min(count - 1, Math.floor((new Date(r.Timestamp).getTime() - start) / bw));
    if (idx >= 0) {
      buckets[idx].total++;
      if (r.StatusCode >= 400) buckets[idx].errors++;
    }
  }
  return buckets;
}

function ThroughputSvg({ data }: { data: ThroughputBucket[] }) {
  if (data.length === 0) {
    return <div class="text-center text-xs text-zinc-500 py-12">No traffic in the selected range.</div>;
  }
  const w = 800, h = 220, padL = 40, padR = 8, padT = 12, padB = 24;
  const innerW = w - padL - padR;
  const innerH = h - padT - padB;
  const maxY = Math.max(1, ...data.map((d) => d.total));
  const yAt = (v: number) => padT + (1 - v / maxY) * innerH;

  // Discrete time buckets render as bars. Each bar is the bucket's total,
  // with the error portion stacked on top in red. Bars communicate the
  // burst-vs-steady nature of traffic far better than a smoothed area
  // path, which falsely implies continuous data between buckets.
  const slot = innerW / data.length;
  const bw = Math.max(2, Math.min(slot - 1, 24));

  const bars = data.map((d, i) => {
    if (d.total === 0) return null;
    const x = padL + i * slot + (slot - bw) / 2;
    const totalH = padT + innerH - yAt(d.total);
    const errH = padT + innerH - yAt(d.errors);
    const successH = totalH - errH;
    const errorY = yAt(d.total);
    return (
      <g key={i}>
        {successH > 0 && (
          <rect x={x} y={yAt(d.total)} width={bw} height={totalH - errH} fill="#18181b" opacity={0.78}>
            <title>{`${fmtTime(d.t)} · ${d.total} req${d.total === 1 ? "" : "s"}`}</title>
          </rect>
        )}
        {d.errors > 0 && (
          <rect x={x} y={errorY} width={bw} height={errH} fill="#ef4444">
            <title>{`${fmtTime(d.t)} · ${d.errors} error${d.errors === 1 ? "" : "s"}`}</title>
          </rect>
        )}
      </g>
    );
  });

  // 3 y-ticks max so the labels breathe.
  const yTicks = maxY <= 2 ? [0, maxY] : [0, Math.round(maxY / 2), maxY];

  // X labels: 3 evenly spaced. Anchored start/middle/end so they don't drift
  // off the edge of the chart.
  const xCount = Math.min(3, data.length);
  const xLabels = Array.from({ length: xCount }, (_, k) => {
    const i = Math.round((k / Math.max(1, xCount - 1)) * (data.length - 1));
    return {
      x: padL + i * slot + slot / 2,
      label: fmtTime(data[i].t),
      anchor: k === 0 ? "start" : k === xCount - 1 ? "end" : ("middle" as const),
    };
  });

  return (
    <svg viewBox={`0 0 ${w} ${h}`} class="w-full h-full">
      {/* Y-axis grid + labels */}
      {yTicks.map((v, i) => (
        <g key={i}>
          <line x1={padL} y1={yAt(v)} x2={w - padR} y2={yAt(v)} stroke="#f4f4f5" />
          <text x={padL - 6} y={yAt(v) + 3} text-anchor="end" font-size="10" fill="#71717a">
            {v}
          </text>
        </g>
      ))}

      {bars}

      {/* X labels — anchored so they never overrun the chart edges. */}
      {xLabels.map((x, i) => (
        <text key={i} x={x.x} y={h - 8} text-anchor={x.anchor as any} font-size="10" fill="#71717a">
          {x.label}
        </text>
      ))}
    </svg>
  );
}

function fmtTime(ms: number): string {
  const d = new Date(ms);
  const hh = d.getHours().toString().padStart(2, "0");
  const mm = d.getMinutes().toString().padStart(2, "0");
  const ss = d.getSeconds().toString().padStart(2, "0");
  return `${hh}:${mm}:${ss}`;
}

function Legend() {
  return (
    <div class="flex items-center gap-4 text-[11px] text-zinc-500">
      <span class="flex items-center gap-1.5">
        <span class="w-2.5 h-2.5 rounded-sm bg-zinc-900/80" />
        Requests
      </span>
      <span class="flex items-center gap-1.5">
        <span class="w-2.5 h-2.5 rounded-sm bg-red-500" />
        Errors
      </span>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────────
// Status breakdown donut
// ────────────────────────────────────────────────────────────────────────

function StatusBreakdown({ s }: { s: Summary }) {
  const segments = [
    { label: "2xx", value: s.twoXX, color: "#10b981" },
    { label: "3xx", value: s.threeXX, color: "#f59e0b" },
    { label: "4xx", value: s.fourXX, color: "#f97316" },
    { label: "5xx", value: s.fiveXX, color: "#ef4444" },
  ];
  const total = segments.reduce((a, b) => a + b.value, 0);

  return (
    <section class="bg-white border border-zinc-200 rounded-xl p-5 flex flex-col">
      <header class="mb-4">
        <h3 class="text-sm font-semibold">Status</h3>
        <p class="text-xs text-zinc-500 mt-0.5">By response class</p>
      </header>
      <div class="flex items-center gap-5">
        <Donut segments={segments} total={total} />
        <div class="flex-1 space-y-1.5">
          {segments.map((seg) => {
            const pct = total === 0 ? 0 : (seg.value / total) * 100;
            return (
              <div key={seg.label} class="flex items-center gap-2 text-xs">
                <span class="w-2 h-2 rounded-sm shrink-0" style={{ background: seg.color }} />
                <span class="text-zinc-500 w-8">{seg.label}</span>
                <span class="flex-1 text-right font-mono tabular-nums">{seg.value}</span>
                <span class="text-zinc-500 font-mono tabular-nums w-12 text-right">
                  {pct.toFixed(0)}%
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

function Donut({
  segments,
  total,
}: {
  segments: { label: string; value: number; color: string }[];
  total: number;
}) {
  // Donut by stroke-dasharray on a circle. Each segment is drawn as a slice
  // of the perimeter, advancing the rotation. Cleaner than path-arc math and
  // crisp at any size.
  const radius = 36;
  const circumference = 2 * Math.PI * radius;
  let acc = 0;

  return (
    <div class="relative shrink-0">
      <svg width="100" height="100" viewBox="0 0 100 100" class="-rotate-90">
        <circle cx="50" cy="50" r={radius} fill="none" stroke="#f4f4f5" stroke-width="12" />
        {segments.map((seg, i) => {
          if (total === 0 || seg.value === 0) return null;
          const len = (seg.value / total) * circumference;
          const offset = -acc;
          acc += len;
          return (
            <circle
              key={i}
              cx="50"
              cy="50"
              r={radius}
              fill="none"
              stroke={seg.color}
              stroke-width="12"
              stroke-dasharray={`${len} ${circumference - len}`}
              stroke-dashoffset={offset}
            />
          );
        })}
      </svg>
      <div class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
        <span class="text-lg font-semibold tabular-nums">{total}</span>
        <span class="text-[10px] text-zinc-500 uppercase tracking-wide">total</span>
      </div>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────────
// Latency histogram
// ────────────────────────────────────────────────────────────────────────

const LATENCY_BUCKETS: { label: string; from: number; to: number }[] = [
  { label: "<10ms", from: 0, to: 10 },
  { label: "10–50ms", from: 10, to: 50 },
  { label: "50–100ms", from: 50, to: 100 },
  { label: "100–200ms", from: 100, to: 200 },
  { label: "200–500ms", from: 200, to: 500 },
  { label: "500ms–1s", from: 500, to: 1000 },
  { label: "1–2s", from: 1000, to: 2000 },
  { label: "2–5s", from: 2000, to: 5000 },
  { label: ">5s", from: 5000, to: Number.POSITIVE_INFINITY },
];

function LatencyHistogram({ requests }: { requests: RequestLog[] }) {
  const counts = useMemo(() => {
    const c = new Array(LATENCY_BUCKETS.length).fill(0);
    for (const r of requests) {
      const i = LATENCY_BUCKETS.findIndex((b) => r.Duration >= b.from && r.Duration < b.to);
      if (i >= 0) c[i]++;
    }
    return c;
  }, [requests]);
  const max = Math.max(1, ...counts);

  return (
    <section class="bg-white border border-zinc-200 rounded-xl p-5">
      <header class="mb-4">
        <h3 class="text-sm font-semibold">Latency distribution</h3>
        <p class="text-xs text-zinc-500 mt-0.5">Where requests fall on the response-time spectrum</p>
      </header>
      <div class="grid grid-cols-9 gap-2 items-end h-32">
        {counts.map((c, i) => {
          const heightPct = (c / max) * 100;
          return (
            <div key={i} class="flex flex-col items-center gap-1 h-full justify-end">
              <span class="text-[10px] font-mono text-zinc-500 tabular-nums">{c}</span>
              <div
                class={cn(
                  "w-full rounded-t",
                  i < 4 ? "bg-emerald-200" : i < 6 ? "bg-amber-300" : "bg-red-400"
                )}
                style={{ height: `${heightPct}%`, minHeight: c > 0 ? "4px" : "0" }}
              />
            </div>
          );
        })}
      </div>
      <div class="grid grid-cols-9 gap-2 mt-2">
        {LATENCY_BUCKETS.map((b, i) => (
          <div key={i} class="text-[10px] text-zinc-500 text-center font-mono">
            {b.label}
          </div>
        ))}
      </div>
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────────
// Slowest endpoints table
// ────────────────────────────────────────────────────────────────────────

interface EndpointStat {
  key: string;
  method: string;
  path: string;
  count: number;
  p50: number;
  p95: number;
  p99: number;
  max: number;
  errors: number;
}

function EndpointsTable({ requests }: { requests: RequestLog[] }) {
  const stats = useMemo<EndpointStat[]>(() => {
    const grouped = new Map<string, { method: string; path: string; ds: number[]; errs: number }>();
    for (const r of requests) {
      const k = `${r.Method} ${r.Path}`;
      let entry = grouped.get(k);
      if (!entry) {
        entry = { method: r.Method, path: r.Path, ds: [], errs: 0 };
        grouped.set(k, entry);
      }
      entry.ds.push(r.Duration);
      if (r.StatusCode >= 400) entry.errs++;
    }
    return Array.from(grouped.entries())
      .map(([k, v]) => {
        const sorted = [...v.ds].sort((a, b) => a - b);
        const pct = (p: number) =>
          sorted[Math.min(sorted.length - 1, Math.floor(sorted.length * p))];
        return {
          key: k,
          method: v.method,
          path: v.path,
          count: v.ds.length,
          p50: pct(0.5),
          p95: pct(0.95),
          p99: pct(0.99),
          max: sorted[sorted.length - 1],
          errors: v.errs,
        };
      })
      .sort((a, b) => b.p95 - a.p95);
  }, [requests]);

  return (
    <section class="bg-white border border-zinc-200 rounded-xl overflow-hidden">
      <header class="px-5 py-3 border-b border-zinc-200 flex items-center justify-between">
        <div>
          <h3 class="text-sm font-semibold">Endpoints</h3>
          <p class="text-xs text-zinc-500 mt-0.5">Sorted by p95, slowest first</p>
        </div>
        <span class="text-xs text-zinc-500 font-mono">{stats.length}</span>
      </header>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-zinc-50/60 border-b border-zinc-200">
            <tr class="text-left text-[11px] uppercase tracking-wide text-zinc-500">
              <th class="px-5 py-2 font-medium">Endpoint</th>
              <th class="px-3 py-2 font-medium text-right w-16">Count</th>
              <th class="px-3 py-2 font-medium text-right w-16">Err</th>
              <th class="px-3 py-2 font-medium text-right w-20">p50</th>
              <th class="px-3 py-2 font-medium text-right w-20">p95</th>
              <th class="px-3 py-2 font-medium text-right w-20">p99</th>
              <th class="px-3 py-2 font-medium text-right w-20">max</th>
              <th class="px-5 py-2 font-medium w-32">Heat</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100">
            {stats.slice(0, 20).map((s) => {
              const heat = Math.min(1, s.p95 / 1000);
              return (
                <tr key={s.key} class="hover:bg-zinc-50">
                  <td class="px-5 py-2">
                    <span class={cn("text-[10px] font-semibold mr-2", methodColor(s.method))}>
                      {s.method}
                    </span>
                    <span class="font-mono text-xs">{s.path}</span>
                  </td>
                  <td class="px-3 py-2 text-right font-mono text-xs tabular-nums">{s.count}</td>
                  <td class="px-3 py-2 text-right font-mono text-xs tabular-nums">
                    {s.errors > 0 ? (
                      <span class="text-red-700">{s.errors}</span>
                    ) : (
                      <span class="text-zinc-400">0</span>
                    )}
                  </td>
                  <td class="px-3 py-2 text-right font-mono text-xs tabular-nums">{formatMs(s.p50)}</td>
                  <td class="px-3 py-2 text-right font-mono text-xs tabular-nums">{formatMs(s.p95)}</td>
                  <td class="px-3 py-2 text-right font-mono text-xs tabular-nums">{formatMs(s.p99)}</td>
                  <td class="px-3 py-2 text-right font-mono text-xs tabular-nums">{formatMs(s.max)}</td>
                  <td class="px-5 py-2">
                    <div class="h-1.5 bg-zinc-100 rounded-full overflow-hidden">
                      <div
                        class={cn(
                          "h-1.5 rounded-full",
                          heat < 0.3 ? "bg-emerald-400" : heat < 0.7 ? "bg-amber-400" : "bg-red-400"
                        )}
                        style={{ width: `${heat * 100}%` }}
                      />
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function methodColor(method: string): string {
  switch (method) {
    case "GET": return "text-blue-700";
    case "POST": return "text-emerald-700";
    case "PUT":
    case "PATCH": return "text-amber-700";
    case "DELETE": return "text-red-700";
    default: return "text-zinc-700";
  }
}

// ────────────────────────────────────────────────────────────────────────
// Data actions — export / import in the header
// ────────────────────────────────────────────────────────────────────────

function DataActions({
  requests, onImport,
}: {
  requests: RequestLog[];
  onImport: (logs: RequestLog[]) => void;
}) {
  const fileInput = useRef<HTMLInputElement>(null);
  const [open, setOpen] = useState(false);
  const [status, setStatus] = useState<"" | "imported" | "failed">("");

  const exportJson = () => {
    const blob = new Blob([api.exportRequests(requests)], { type: "application/json" });
    download(blob, `govisual-${stamp()}.json`);
    setOpen(false);
  };
  const exportCsv = () => {
    const headers = ["ID", "Timestamp", "Method", "Path", "Status", "Duration (ms)", "Error"];
    const escape = (cell: unknown) =>
      `"${String(cell ?? "").replace(/"/g, '""')}"`;
    const rows = requests.map((r) => [
      r.ID, r.Timestamp, r.Method, r.Path, r.StatusCode, r.Duration, r.Error || "",
    ]);
    const body = [headers.map(escape).join(","), ...rows.map((row) => row.map(escape).join(","))].join("\n");
    download(new Blob([body], { type: "text/csv" }), `govisual-${stamp()}.csv`);
    setOpen(false);
  };
  const onFile = (event: Event) => {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = api.importRequests(e.target?.result as string);
        onImport(data);
        setStatus("imported");
      } catch (err) {
        console.error("Import failed:", err);
        setStatus("failed");
      }
      setTimeout(() => setStatus(""), 2500);
    };
    reader.readAsText(file);
    (event.target as HTMLInputElement).value = "";
  };

  return (
    <div class="relative">
      <button
        onClick={() => setOpen((v) => !v)}
        class="text-xs border border-zinc-200 rounded-md px-2.5 py-1.5 hover:bg-zinc-50 flex items-center gap-1.5"
      >
        Data
        <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      {open && (
        <Fragment>
          <button
            class="fixed inset-0 z-30 cursor-default"
            onClick={() => setOpen(false)}
            aria-label="Close menu"
          />
          <div class="absolute right-0 top-full mt-1 w-44 bg-white border border-zinc-200 rounded-md shadow-md z-40 py-1">
            <MenuItem onClick={exportJson} label="Export JSON" hint={`${requests.length} rows`} />
            <MenuItem onClick={exportCsv} label="Export CSV" hint={`${requests.length} rows`} />
            <div class="h-px bg-zinc-100 my-1" />
            <MenuItem
              onClick={() => { setOpen(false); fileInput.current?.click(); }}
              label="Import JSON"
            />
          </div>
        </Fragment>
      )}
      {status && (
        <span
          class={cn(
            "absolute right-0 -bottom-6 text-[11px]",
            status === "imported" ? "text-emerald-700" : "text-red-700"
          )}
        >
          {status === "imported" ? "Imported ✓" : "Import failed"}
        </span>
      )}
      <input
        ref={fileInput}
        type="file"
        accept=".json,application/json"
        class="hidden"
        onChange={onFile}
      />
    </div>
  );
}

function MenuItem({
  onClick, label, hint,
}: {
  onClick: () => void;
  label: string;
  hint?: string;
}) {
  return (
    <button
      onClick={onClick}
      class="w-full px-3 py-1.5 text-left text-sm hover:bg-zinc-50 flex items-center justify-between"
    >
      <span>{label}</span>
      {hint && <span class="text-[11px] text-zinc-500 font-mono">{hint}</span>}
    </button>
  );
}

function download(blob: Blob, name: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function stamp(): string {
  return new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19);
}

// ────────────────────────────────────────────────────────────────────────
// Empty state
// ────────────────────────────────────────────────────────────────────────

function EmptyState({ range }: { range: Range }) {
  const label = range === "all" ? "yet" : `in the last ${range}`;
  return (
    <div class="bg-white border border-zinc-200 rounded-xl py-16 px-8 text-center">
      <div class="w-12 h-12 mx-auto rounded-full bg-zinc-100 flex items-center justify-center mb-3 text-zinc-400">
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="20" x2="18" y2="10" />
          <line x1="12" y1="20" x2="12" y2="4" />
          <line x1="6" y1="20" x2="6" y2="14" />
        </svg>
      </div>
      <h3 class="text-sm font-medium text-zinc-900">No requests {label}</h3>
      <p class="text-xs text-zinc-500 mt-1">
        Generate some traffic and the charts will populate live.
      </p>
    </div>
  );
}
