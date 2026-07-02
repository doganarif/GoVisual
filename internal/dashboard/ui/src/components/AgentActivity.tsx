import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import { AgentActivity as Entry, api } from "@/lib/api";
import { cn } from "@/lib/utils";

// AgentActivity shows what a coding agent has done through the MCP endpoint.
// Empty state means either no agent connected yet, or the host didn't wire
// govisual.WithActivityLog + mcp.WithActivityLog through — we show a hint.
export function AgentActivity() {
  const [entries, setEntries] = useState<Entry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    const load = () => {
      api
        .getAgentActivity()
        .then((data) => {
          if (alive) {
            setEntries(data || []);
            setLoading(false);
          }
        })
        .catch(() => {
          if (alive) setLoading(false);
        });
    };
    load();
    const t = setInterval(load, 3000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, []);

  return (
    <main class="flex-1 flex flex-col bg-white overflow-hidden">
      <div class="px-6 py-4 border-b border-zinc-200">
        <h2 class="text-base font-medium">Agent activity</h2>
        <p class="text-xs text-zinc-500 mt-1">
          Recent MCP tool calls, newest first. Refreshes every 3s.
        </p>
      </div>
      <div class="flex-1 overflow-auto p-6">
        {loading ? (
          <div class="text-xs text-zinc-500 text-center py-8">Loading…</div>
        ) : entries.length === 0 ? (
          <EmptyState />
        ) : (
          <ul class="border border-zinc-200 rounded-lg divide-y divide-zinc-100">
            {entries.map((e, i) => (
              <li key={i} class="px-4 py-3 text-xs">
                <div class="flex items-baseline gap-2 flex-wrap">
                  <span class="text-zinc-400 font-mono shrink-0">
                    {formatTime(e.time)}
                  </span>
                  <span
                    class={cn(
                      "px-1.5 py-0.5 rounded font-mono text-[11px]",
                      e.mutating
                        ? "bg-amber-50 text-amber-700"
                        : "bg-blue-50 text-blue-700"
                    )}
                  >
                    {e.tool}
                  </span>
                  <span class="text-zinc-500 font-mono">
                    {formatDuration(e.duration)}
                  </span>
                  {e.error && (
                    <span class="text-red-600 font-mono truncate">
                      {e.error}
                    </span>
                  )}
                </div>
                {e.args && Object.keys(e.args).length > 0 && (
                  <div class="mt-1 pl-4 flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] font-mono text-zinc-500">
                    {Object.entries(e.args).map(([k, v]) => (
                      <span key={k}>
                        <span class="text-zinc-400">{k}=</span>
                        <span class="text-zinc-700 break-all">{v}</span>
                      </span>
                    ))}
                  </div>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </main>
  );
}

function EmptyState() {
  return (
    <Fragment>
      <div class="text-xs text-zinc-500 max-w-lg mx-auto text-center py-8">
        <p class="mb-3">No agent activity yet.</p>
        <p class="mb-2">
          Share a{" "}
          <code class="font-mono bg-zinc-100 px-1 rounded">
            store.NewActivityLog(200)
          </code>{" "}
          between{" "}
          <code class="font-mono bg-zinc-100 px-1 rounded">
            govisual.Wrap
          </code>{" "}
          and the MCP handler:
        </p>
        <pre class="text-left text-[11px] font-mono bg-zinc-50 border border-zinc-200 rounded p-3 overflow-x-auto">
{`log := store.NewActivityLog(200)
app := govisual.Wrap(mux,
    govisual.WithStore(st),
    govisual.WithActivityLog(log),
)
root.Handle("/mcp", gvmcp.Handler(st,
    gvmcp.WithActivityLog(log),
))`}
        </pre>
      </div>
    </Fragment>
  );
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "";
  return (
    d.toLocaleTimeString("en-US", { hour12: false }) +
    "." +
    String(d.getMilliseconds()).padStart(3, "0")
  );
}

function formatDuration(ns: number): string {
  if (!ns) return "0ms";
  const ms = ns / 1_000_000;
  if (ms < 1) return Math.round(ns / 1000) + "μs";
  if (ms < 1000) return ms.toFixed(1) + "ms";
  return (ms / 1000).toFixed(2) + "s";
}
