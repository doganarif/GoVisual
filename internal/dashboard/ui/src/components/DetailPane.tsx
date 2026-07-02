import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import {
  RequestLog,
  LogEntry,
  api,
  ApiError,
  PerformanceMetrics,
  FlameGraphNode,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { FlameGraph } from "./FlameGraph";

type Tab = "overview" | "headers" | "body" | "trace" | "logs" | "performance";

interface DetailPaneProps {
  request: RequestLog | null;
  onReplay?: (req: RequestLog) => void;
  onCompareAdd?: (req: RequestLog) => void;
  comparePending?: boolean;
}

export function DetailPane({
  request,
  onReplay,
  onCompareAdd,
  comparePending,
}: DetailPaneProps) {
  const [tab, setTab] = useState<Tab>("overview");
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [flame, setFlame] = useState<FlameGraphNode | null>(null);
  const [loadingMetrics, setLoadingMetrics] = useState(false);

  // Reset tab + metrics when the selected request changes. An AbortController
  // cancels any in-flight metric load so a slow response from a previous
  // selection can't overwrite the current pane.
  useEffect(() => {
    if (!request?.ID) {
      setMetrics(null);
      setFlame(null);
      setTab("overview");
      return;
    }
    const controller = new AbortController();
    setTab("overview");
    setMetrics(null);
    setFlame(null);
    setLoadingMetrics(true);
    api
      .getMetrics(request.ID, controller.signal)
      .then((m) => setMetrics(m))
      .catch((err) => {
        if (err?.name === "AbortError") return;
        // 501 = profiling disabled, 404 = no metrics for this id.
        // Both are normal states, not errors worth logging.
        if (
          err instanceof ApiError &&
          (err.status === 501 || err.status === 404)
        ) {
          setMetrics(null);
          return;
        }
        console.error("Failed to load metrics:", err);
        setMetrics(null);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoadingMetrics(false);
      });
    return () => controller.abort();
  }, [request?.ID]);

  // Lazy-load the flame graph the first time the Performance tab is shown.
  useEffect(() => {
    if (tab !== "performance" || !request?.ID || flame) return;
    const controller = new AbortController();
    api
      .getFlameGraph(request.ID, controller.signal)
      .then(setFlame)
      .catch((err) => {
        if (err?.name === "AbortError") return;
        if (err instanceof ApiError && (err.status === 404 || err.status === 501)) {
          setFlame(null);
          return;
        }
        console.error("Failed to load flame graph:", err);
      });
    return () => controller.abort();
  }, [tab, request?.ID]);

  if (!request) {
    return (
      <main class="flex-1 flex items-center justify-center bg-zinc-50/40">
        <div class="text-center max-w-sm">
          <div class="w-12 h-12 mx-auto rounded-full bg-zinc-100 flex items-center justify-center mb-3 text-zinc-400">
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="9 11 12 14 22 4" />
              <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
            </svg>
          </div>
          <h3 class="text-sm font-medium text-zinc-900">No request selected</h3>
          <p class="text-xs text-zinc-500 mt-1">
            Pick a request from the list to see headers, body, and timing.
          </p>
        </div>
      </main>
    );
  }

  const hasMetrics = !!metrics || !!request.PerformanceMetrics;
  const activeMetrics = metrics || request.PerformanceMetrics || null;
  const hasLogs = !!request.Logs && request.Logs.length > 0;

  return (
    <main class="flex-1 flex flex-col bg-white overflow-hidden">
      {/* Header */}
      <div class="px-6 py-4 border-b border-zinc-200 flex items-start justify-between gap-4">
        <div class="min-w-0">
          <div class="flex items-center gap-2 mb-1 flex-wrap">
            <span class="text-xs font-semibold text-blue-700 px-1.5 py-0.5 rounded bg-blue-50">
              {request.Method}
            </span>
            <h2 class="text-base font-mono truncate">{request.Path}</h2>
            <span class={cn("text-xs font-mono px-1.5 py-0.5 rounded", statusPill(request.StatusCode))}>
              {request.StatusCode}
            </span>
          </div>
          <div class="text-xs text-zinc-500 flex items-center gap-3 flex-wrap">
            <span>{formatDuration(request.Duration)}</span>
            <span>·</span>
            <span>{new Date(request.Timestamp).toLocaleString()}</span>
            <span>·</span>
            <span class="font-mono truncate" title={request.ID}>
              {request.ID.slice(0, 12)}…
            </span>
          </div>
        </div>
        <div class="flex items-center gap-1 shrink-0">
          {onReplay && (
            <button
              onClick={() => onReplay(request)}
              class="text-xs border border-zinc-200 rounded-md px-2.5 py-1.5 hover:bg-zinc-50"
            >
              Replay
            </button>
          )}
          {onCompareAdd && (
            <button
              onClick={() => onCompareAdd(request)}
              class={cn(
                "text-xs border rounded-md px-2.5 py-1.5",
                comparePending
                  ? "bg-zinc-900 text-white border-zinc-900"
                  : "border-zinc-200 hover:bg-zinc-50"
              )}
            >
              {comparePending ? "Selected" : "Compare"}
            </button>
          )}
          <button
            onClick={() => copyAsCurl(request)}
            class="text-xs border border-zinc-200 rounded-md px-2.5 py-1.5 hover:bg-zinc-50"
            title="Copy as curl"
          >
            Copy cURL
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div class="px-6 border-b border-zinc-200">
        <nav class="flex gap-1 -mb-px">
          {(["overview", "headers", "body", "trace"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              class={cn(
                "px-3 py-2.5 text-sm border-b-2",
                tab === t
                  ? "border-zinc-900 text-zinc-900 font-medium"
                  : "border-transparent text-zinc-500 hover:text-zinc-900"
              )}
            >
              {labelFor(t)}
            </button>
          ))}
          {hasLogs && (
            <button
              onClick={() => setTab("logs")}
              class={cn(
                "px-3 py-2.5 text-sm border-b-2",
                tab === "logs"
                  ? "border-zinc-900 text-zinc-900 font-medium"
                  : "border-transparent text-zinc-500 hover:text-zinc-900"
              )}
            >
              Logs · {request.Logs!.length}
            </button>
          )}
          {hasMetrics && (
            <button
              onClick={() => setTab("performance")}
              class={cn(
                "px-3 py-2.5 text-sm border-b-2",
                tab === "performance"
                  ? "border-zinc-900 text-zinc-900 font-medium"
                  : "border-transparent text-zinc-500 hover:text-zinc-900"
              )}
            >
              Performance
            </button>
          )}
        </nav>
      </div>

      <div class="flex-1 overflow-auto p-6 space-y-5">
        {tab === "overview" && <Overview request={request} />}
        {tab === "headers" && <Headers request={request} />}
        {tab === "body" && <Body request={request} />}
        {tab === "trace" && <Trace request={request} metrics={activeMetrics} />}
        {tab === "logs" && <Logs request={request} />}
        {tab === "performance" && (
          <Performance
            metrics={activeMetrics}
            flame={flame}
            loading={loadingMetrics}
          />
        )}
      </div>
    </main>
  );
}

function labelFor(t: Tab): string {
  switch (t) {
    case "overview":
      return "Overview";
    case "headers":
      return "Headers";
    case "body":
      return "Body";
    case "trace":
      return "Trace";
    case "logs":
      return "Logs";
    case "performance":
      return "Performance";
  }
}

function statusPill(status: number): string {
  if (status >= 200 && status < 300) return "bg-emerald-50 text-emerald-700";
  if (status >= 300 && status < 400) return "bg-amber-50 text-amber-700";
  if (status >= 400 && status < 500) return "bg-orange-50 text-orange-700";
  if (status >= 500) return "bg-red-50 text-red-700";
  return "bg-zinc-100 text-zinc-700";
}

function formatDuration(d: number): string {
  if (d < 1) return "<1ms";
  if (d < 1000) return `${d}ms`;
  return `${(d / 1000).toFixed(2)}s`;
}

function formatDurationNs(ns?: number): string {
  if (!ns) return "0ms";
  const ms = ns / 1_000_000;
  if (ms < 1) return Math.round(ns / 1000) + "μs";
  if (ms < 1000) return ms.toFixed(2) + "ms";
  return (ms / 1000).toFixed(2) + "s";
}

function formatBytes(bytes?: number): string {
  if (!bytes) return "0 B";
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + " " + sizes[i];
}

function formatBody(body?: string): string {
  if (!body) return "No body";
  try {
    return JSON.stringify(JSON.parse(body), null, 2);
  } catch {
    return body;
  }
}

function copyAsCurl(req: RequestLog) {
  const host = req.RequestHeaders?.Host?.[0] || "localhost";
  const url = `http://${host}${req.Path}${req.Query ? "?" + req.Query : ""}`;
  const parts = [`curl -X ${req.Method} ${shellQuote(url)}`];
  for (const [k, values] of Object.entries(req.RequestHeaders || {})) {
    for (const v of values) {
      parts.push(`-H ${shellQuote(`${k}: ${v}`)}`);
    }
  }
  if (req.RequestBody) {
    parts.push(`-d ${shellQuote(req.RequestBody)}`);
  }
  const cmd = parts.join(" \\\n  ");
  navigator.clipboard.writeText(cmd).catch(() => {});
}

function shellQuote(s: string): string {
  return `'${s.replace(/'/g, "'\\''")}'`;
}

function Overview({ request }: { request: RequestLog }) {
  return (
    <Fragment>
      <section class="grid grid-cols-4 gap-3">
        <Stat label="Duration" value={formatDuration(request.Duration)} />
        <Stat label="Status" value={String(request.StatusCode)} />
        <Stat
          label="Response size"
          value={formatBytes(request.ResponseBody?.length || 0)}
        />
        <Stat
          label="Query"
          value={request.Query || "—"}
          mono
        />
      </section>

      <section class="border border-zinc-200 rounded-lg">
        <header class="px-4 py-2.5 border-b border-zinc-200 flex items-center justify-between">
          <h3 class="text-sm font-medium">Timeline</h3>
          <span class="text-xs text-zinc-500">total {formatDuration(request.Duration)}</span>
        </header>
        <div class="p-4">
          <div class="flex items-center gap-3 text-xs font-mono">
            <span class="w-24 text-zinc-500">Duration</span>
            <div class="flex-1 h-1.5 bg-zinc-100 rounded-full overflow-hidden">
              <div class="h-1.5 bg-blue-400 rounded-full" style={{ width: "100%" }} />
            </div>
            <span class="w-16 text-right">{formatDuration(request.Duration)}</span>
          </div>
        </div>
      </section>

      {request.Error && (
        <section class="border border-red-200 bg-red-50/50 rounded-lg p-4">
          <h3 class="text-sm font-medium text-red-800 mb-1">Error</h3>
          <pre class="text-xs font-mono text-red-700 whitespace-pre-wrap">
            {request.Error}
          </pre>
          {request.PanicStack && (
            <pre class="text-[11px] font-mono text-red-600/80 whitespace-pre-wrap break-all mt-3 pt-3 border-t border-red-200 max-h-64 overflow-auto">
              {request.PanicStack}
            </pre>
          )}
        </section>
      )}
    </Fragment>
  );
}

function Logs({ request }: { request: RequestLog }) {
  const logs = request.Logs || [];
  if (logs.length === 0) {
    return (
      <div class="text-xs text-zinc-500 text-center py-8">
        No log lines captured. Wrap your slog handler with{" "}
        <code class="font-mono bg-zinc-100 px-1 rounded">
          govisual.SlogHandler(...)
        </code>{" "}
        and log with the request context to capture per-request lines here.
      </div>
    );
  }
  return (
    <section class="border border-zinc-200 rounded-lg">
      <header class="px-4 py-2.5 border-b border-zinc-200 flex items-center justify-between">
        <h3 class="text-sm font-medium">Application logs</h3>
        <span class="text-[11px] text-zinc-500">{logs.length} lines</span>
      </header>
      <div class="divide-y divide-zinc-100">
        {logs.map((entry, i) => (
          <LogRow key={i} entry={entry} />
        ))}
      </div>
    </section>
  );
}

function LogRow({ entry }: { entry: LogEntry }) {
  const attrs = entry.attrs ? Object.entries(entry.attrs) : [];
  const t = new Date(entry.time);
  const tStr = isNaN(t.getTime())
    ? ""
    : t.toLocaleTimeString("en-US", { hour12: false }) +
      "." +
      String(t.getMilliseconds()).padStart(3, "0");
  return (
    <div class="px-4 py-2 text-xs">
      <div class="flex items-baseline gap-2 font-mono">
        {tStr && <span class="text-zinc-400 shrink-0">{tStr}</span>}
        <span class={cn("shrink-0 font-semibold", logLevelColor(entry.level))}>
          {entry.level}
        </span>
        <span class="text-zinc-900 break-all">{entry.message}</span>
      </div>
      {attrs.length > 0 && (
        <div class="mt-1 pl-4 flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] font-mono text-zinc-500">
          {attrs.map(([k, v]) => (
            <span key={k}>
              <span class="text-zinc-400">{k}=</span>
              <span class="text-zinc-700">{formatAttr(v)}</span>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function formatAttr(v: unknown): string {
  if (typeof v === "string") return v;
  try {
    return JSON.stringify(v);
  } catch {
    return String(v);
  }
}

function logLevelColor(level: string): string {
  switch (level.toUpperCase()) {
    case "ERROR":
      return "text-red-600";
    case "WARN":
    case "WARNING":
      return "text-amber-600";
    case "EVENT":
      return "text-blue-600";
    case "DEBUG":
      return "text-zinc-500";
    default:
      return "text-emerald-700";
  }
}

function Stat({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div class="border border-zinc-200 rounded-lg p-3">
      <div class="text-[11px] text-zinc-500 mb-1">{label}</div>
      <div class={cn("font-semibold truncate", mono ? "text-sm font-mono" : "text-lg")}>
        {value}
      </div>
    </div>
  );
}

function Headers({ request }: { request: RequestLog }) {
  return (
    <Fragment>
      <HeadersCard title="Request headers" headers={request.RequestHeaders} />
      <HeadersCard title="Response headers" headers={request.ResponseHeaders} />
    </Fragment>
  );
}

function HeadersCard({
  title,
  headers,
}: {
  title: string;
  headers?: Record<string, string[]>;
}) {
  const entries = Object.entries(headers || {});
  return (
    <section class="border border-zinc-200 rounded-lg">
      <header class="px-4 py-2.5 border-b border-zinc-200">
        <h3 class="text-sm font-medium">{title}</h3>
      </header>
      {entries.length === 0 ? (
        <div class="px-4 py-3 text-xs text-zinc-500">No headers</div>
      ) : (
        <div class="divide-y divide-zinc-100">
          {entries.map(([k, values]) => (
            <div key={k} class="grid grid-cols-[180px_1fr] gap-3 px-4 py-2 text-xs font-mono">
              <div class="text-zinc-500 truncate" title={k}>{k}</div>
              <div class="text-zinc-900 break-all">{values.join(", ")}</div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

function Body({ request }: { request: RequestLog }) {
  return (
    <Fragment>
      <BodyCard title="Request body" body={request.RequestBody} />
      <BodyCard title="Response body" body={request.ResponseBody} />
    </Fragment>
  );
}

function BodyCard({ title, body }: { title: string; body?: string }) {
  return (
    <section class="border border-zinc-200 rounded-lg">
      <header class="px-4 py-2.5 border-b border-zinc-200 flex items-center justify-between">
        <h3 class="text-sm font-medium">{title}</h3>
        <span class="text-[11px] text-zinc-500">{body?.length || 0} bytes</span>
      </header>
      <pre class="p-4 text-xs font-mono overflow-auto max-h-96 bg-zinc-50/50 rounded-b-lg whitespace-pre-wrap break-all">
        {formatBody(body)}
      </pre>
    </section>
  );
}

function Trace({
  request,
  metrics,
}: {
  request: RequestLog;
  metrics: PerformanceMetrics | null;
}) {
  const mw = request.MiddlewareTrace || [];
  const sql = metrics?.sql_queries || [];
  const http = metrics?.http_calls || [];
  return (
    <Fragment>
      {mw.length > 0 && (
        <section class="border border-zinc-200 rounded-lg">
          <header class="px-4 py-2.5 border-b border-zinc-200">
            <h3 class="text-sm font-medium">Middleware</h3>
          </header>
          <div class="divide-y divide-zinc-100">
            {mw.map((m: any, i: number) => (
              <div key={i} class="px-4 py-2 flex items-center justify-between text-xs">
                <span class="font-mono">{m.name || `Middleware ${i + 1}`}</span>
                <span class="text-zinc-500 font-mono">{formatDuration(m.duration || 0)}</span>
              </div>
            ))}
          </div>
        </section>
      )}

      {sql.length > 0 && (
        <section class="border border-zinc-200 rounded-lg">
          <header class="px-4 py-2.5 border-b border-zinc-200">
            <h3 class="text-sm font-medium">SQL queries · {sql.length}</h3>
          </header>
          <div class="divide-y divide-zinc-100">
            {sql.slice(0, 10).map((q, i) => (
              <div key={i} class="px-4 py-2 text-xs">
                <pre class="font-mono text-emerald-700 whitespace-pre-wrap break-all">{q.query}</pre>
                <div class="text-zinc-500 mt-1 font-mono">
                  {formatDurationNs(q.duration)} · {q.rows} rows
                  {q.error && <span class="text-red-600 ml-2">· {q.error}</span>}
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {http.length > 0 && (
        <section class="border border-zinc-200 rounded-lg">
          <header class="px-4 py-2.5 border-b border-zinc-200">
            <h3 class="text-sm font-medium">HTTP calls · {http.length}</h3>
          </header>
          <div class="divide-y divide-zinc-100">
            {http.slice(0, 10).map((c, i) => (
              <div key={i} class="px-4 py-2 text-xs flex items-center justify-between">
                <span class="font-mono truncate">
                  <span class="text-blue-700 font-semibold mr-2">{c.method}</span>
                  {c.url}
                </span>
                <span class="text-zinc-500 font-mono shrink-0 ml-2">
                  {c.status} · {formatDurationNs(c.duration)}
                </span>
              </div>
            ))}
          </div>
        </section>
      )}

      {mw.length === 0 && sql.length === 0 && http.length === 0 && (
        <div class="text-xs text-zinc-500 text-center py-8">
          No trace data captured for this request.
          <br />
          Enable profiling with <code class="font-mono bg-zinc-100 px-1 rounded">govisual.WithProfiling(true)</code> to see SQL queries and outbound HTTP calls.
        </div>
      )}
    </Fragment>
  );
}

function Performance({
  metrics,
  flame,
  loading,
}: {
  metrics: PerformanceMetrics | null;
  flame: FlameGraphNode | null;
  loading: boolean;
}) {
  if (loading) {
    return (
      <div class="text-xs text-zinc-500 text-center py-8">Loading metrics…</div>
    );
  }
  if (!metrics) {
    return (
      <div class="text-xs text-zinc-500 text-center py-8">
        Profiling is not enabled. Pass{" "}
        <code class="font-mono bg-zinc-100 px-1 rounded">
          govisual.WithProfiling(true)
        </code>{" "}
        on the server to see CPU and memory metrics here.
      </div>
    );
  }
  return (
    <Fragment>
      <section class="grid grid-cols-4 gap-3">
        <Stat label="CPU time" value={formatDurationNs(metrics.cpu_time)} />
        <Stat label="Memory" value={formatBytes(metrics.memory_alloc)} />
        <Stat label="Goroutines" value={String(metrics.num_goroutines || 0)} />
        <Stat label="GC pause" value={formatDurationNs(metrics.gc_pause_total)} />
      </section>

      {metrics.bottlenecks && metrics.bottlenecks.length > 0 && (
        <section class="border border-zinc-200 rounded-lg">
          <header class="px-4 py-2.5 border-b border-zinc-200">
            <h3 class="text-sm font-medium">Bottlenecks</h3>
          </header>
          <div class="divide-y divide-zinc-100">
            {metrics.bottlenecks.map((b, i) => (
              <div key={i} class="px-4 py-3">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-medium uppercase tracking-wide text-zinc-500">
                    {b.type}
                  </span>
                  <span class="text-xs font-mono text-zinc-700">
                    {(b.impact * 100).toFixed(1)}% · {formatDurationNs(b.duration)}
                  </span>
                </div>
                <div class="text-sm text-zinc-900">{b.description}</div>
                {b.suggestion && (
                  <div class="text-xs text-zinc-500 mt-1">{b.suggestion}</div>
                )}
              </div>
            ))}
          </div>
        </section>
      )}

      {flame && (
        <section class="border border-zinc-200 rounded-lg">
          <header class="px-4 py-2.5 border-b border-zinc-200">
            <h3 class="text-sm font-medium">CPU flame graph</h3>
          </header>
          <div class="p-4 overflow-x-auto">
            <FlameGraph data={flame} width={900} height={400} />
          </div>
        </section>
      )}
    </Fragment>
  );
}
