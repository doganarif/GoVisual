import { h, Fragment } from "preact";
import { useMemo } from "preact/hooks";
import { RequestLog } from "@/lib/api";
import { cn } from "@/lib/utils";

interface RequestListProps {
  title: string;
  subtitle?: string;
  requests: RequestLog[];
  selectedId?: string;
  onSelect: (req: RequestLog) => void;
  // statusFilter: which status-class chips are active. Empty set = all.
  statusFilter: Set<"2xx" | "3xx" | "4xx" | "5xx">;
  onStatusFilterChange: (next: Set<"2xx" | "3xx" | "4xx" | "5xx">) => void;
  search: string;
  onSearchChange: (q: string) => void;
  live?: boolean;
}

const methodColor: Record<string, string> = {
  GET: "text-blue-700",
  POST: "text-emerald-700",
  PUT: "text-amber-700",
  PATCH: "text-amber-700",
  DELETE: "text-red-700",
  HEAD: "text-zinc-500",
  OPTIONS: "text-zinc-500",
};

const statusPillClass = (status: number) => {
  if (status >= 200 && status < 300) return "bg-emerald-50 text-emerald-700";
  if (status >= 300 && status < 400) return "bg-amber-50 text-amber-700";
  if (status >= 400 && status < 500) return "bg-orange-50 text-orange-700";
  if (status >= 500) return "bg-red-50 text-red-700";
  return "bg-zinc-100 text-zinc-700";
};

function bucket(ts: string): string {
  // Group by HH:MM in local time. This matches what the user reads in the row,
  // so the headers line up with row timestamps without surprises.
  const d = new Date(ts);
  if (isNaN(d.getTime())) return "";
  const today = new Date();
  const sameDay =
    d.getFullYear() === today.getFullYear() &&
    d.getMonth() === today.getMonth() &&
    d.getDate() === today.getDate();
  const hh = d.getHours().toString().padStart(2, "0");
  const mm = d.getMinutes().toString().padStart(2, "0");
  return sameDay ? `Today, ${hh}:${mm}` : d.toLocaleString();
}

export function RequestList({
  title,
  subtitle,
  requests,
  selectedId,
  onSelect,
  statusFilter,
  onStatusFilterChange,
  search,
  onSearchChange,
  live,
}: RequestListProps) {
  // Pre-group rows by their HH:MM bucket. useMemo so we don't re-bucket on
  // every render — relevant when the list is long and SSE ticks frequently.
  const groups = useMemo(() => {
    const out: { label: string; items: RequestLog[] }[] = [];
    let last = "";
    for (const r of requests) {
      const b = bucket(r.Timestamp);
      if (b !== last) {
        out.push({ label: b, items: [r] });
        last = b;
      } else {
        out[out.length - 1].items.push(r);
      }
    }
    return out;
  }, [requests]);

  const toggle = (k: "2xx" | "3xx" | "4xx" | "5xx") => {
    const next = new Set(statusFilter);
    if (next.has(k)) next.delete(k);
    else next.add(k);
    onStatusFilterChange(next);
  };

  return (
    <aside class="w-[340px] border-r border-zinc-200 bg-white flex flex-col shrink-0">
      <div class="px-4 py-3 border-b border-zinc-200">
        <div class="flex items-center justify-between mb-2">
          <h2 class="text-sm font-semibold tracking-tight">{title}</h2>
          <span class="text-[11px] text-zinc-500 font-mono">{requests.length}</span>
        </div>
        {subtitle && (
          <p class="text-[11px] text-zinc-500 mb-2 -mt-1">{subtitle}</p>
        )}
        <input
          value={search}
          onInput={(e) => onSearchChange((e.target as HTMLInputElement).value)}
          placeholder="Filter by path..."
          class="w-full text-sm px-2.5 py-1.5 bg-zinc-50 border border-zinc-200 rounded-md focus:outline-none focus:ring-2 focus:ring-zinc-900/10 placeholder:text-zinc-400"
        />
        <div class="flex items-center gap-1.5 flex-wrap mt-2">
          {(["2xx", "3xx", "4xx", "5xx"] as const).map((k) => {
            const on = statusFilter.has(k);
            const cls =
              k === "2xx"
                ? "bg-emerald-50 text-emerald-700 ring-emerald-600/10"
                : k === "3xx"
                ? "bg-amber-50 text-amber-700 ring-amber-600/10"
                : k === "4xx"
                ? "bg-orange-50 text-orange-700 ring-orange-600/10"
                : "bg-red-50 text-red-700 ring-red-600/10";
            return (
              <button
                key={k}
                onClick={() => toggle(k)}
                class={cn(
                  "text-[11px] px-2 py-0.5 rounded-full ring-1",
                  on ? cls : "bg-zinc-50 text-zinc-500 ring-zinc-200"
                )}
              >
                {k}
              </button>
            );
          })}
        </div>
      </div>

      <div class="flex-1 overflow-auto">
        {groups.length === 0 ? (
          <div class="px-4 py-10 text-center text-xs text-zinc-500">
            No matching requests yet.
          </div>
        ) : (
          groups.map((g) => (
            <Fragment key={g.label}>
              <div class="px-4 py-1.5 text-[10px] uppercase tracking-wide text-zinc-500 bg-zinc-50/50 sticky top-0">
                {g.label}
              </div>
              {g.items.map((r) => {
                const isActive = r.ID === selectedId;
                return (
                  <button
                    key={r.ID}
                    onClick={() => onSelect(r)}
                    class={cn(
                      "w-full text-left px-4 py-2.5 border-b border-zinc-100",
                      isActive
                        ? "bg-zinc-50 border-l-2 border-l-zinc-900"
                        : "hover:bg-zinc-50 border-l-2 border-l-transparent"
                    )}
                  >
                    <div class="flex items-center justify-between mb-0.5">
                      <span class="flex items-center gap-2 min-w-0">
                        <span
                          class={cn(
                            "text-[10px] font-semibold shrink-0",
                            methodColor[r.Method] || "text-zinc-700"
                          )}
                        >
                          {r.Method}
                        </span>
                        <span class="text-[10px] text-zinc-400">·</span>
                        <span
                          class={cn(
                            "text-[10px] font-mono px-1.5 py-0.5 rounded shrink-0",
                            statusPillClass(r.StatusCode)
                          )}
                        >
                          {r.StatusCode}
                        </span>
                      </span>
                      <span class="text-[11px] text-zinc-500 font-mono">
                        {formatDuration(r.Duration)}
                      </span>
                    </div>
                    <div class="text-sm font-mono truncate text-zinc-900">
                      {r.Path}
                    </div>
                  </button>
                );
              })}
            </Fragment>
          ))
        )}
      </div>

      <div class="border-t border-zinc-200 px-3 py-2 text-[11px] text-zinc-500 flex items-center justify-between">
        <span>{requests.length} requests</span>
        <span class="flex items-center gap-1.5">
          <span
            class={cn(
              "w-1.5 h-1.5 rounded-full",
              live ? "bg-emerald-500 animate-pulse" : "bg-zinc-300"
            )}
          />
          {live ? "Live" : "Idle"}
        </span>
      </div>
    </aside>
  );
}

function formatDuration(d: number): string {
  if (d < 1) return "<1ms";
  if (d < 1000) return `${d}ms`;
  return `${(d / 1000).toFixed(2)}s`;
}
