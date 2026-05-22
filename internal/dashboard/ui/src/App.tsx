import { h, Fragment } from "preact";
import { useEffect, useMemo, useRef, useState } from "preact/hooks";
import { api, RequestLog } from "./lib/api";
import { RailNav, View } from "./components/RailNav";
import { RequestList } from "./components/RequestList";
import { DetailPane } from "./components/DetailPane";
import { EnvironmentInfo } from "./components/EnvironmentInfo";
import { RequestComparison } from "./components/RequestComparison";
import { RequestReplay } from "./components/RequestReplay";
import { Analytics } from "./components/Analytics";

type StatusFilter = Set<"2xx" | "3xx" | "4xx" | "5xx">;

// Slow threshold (ms): anything at or above this lands in the Slow view.
// Matches the default profiling threshold on the server (10ms) but bumped to
// 200ms to be useful as a triage tool rather than a profiler.
const SLOW_MS = 200;

export function App() {
  const [requests, setRequests] = useState<RequestLog[]>([]);
  const [view, setView] = useState<View>("inbox");
  const [selected, setSelected] = useState<RequestLog | null>(null);

  // Search + status filter live at the App level so they persist across view
  // changes — switching from Inbox to Errors and back shouldn't drop the
  // user's typed filter.
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>(new Set());

  const [compareIds, setCompareIds] = useState<string[]>([]);
  const [showComparison, setShowComparison] = useState(false);
  const [replayRequest, setReplayRequest] = useState<RequestLog | null>(null);
  const [live, setLive] = useState(false);

  // Initial fetch + live subscription. The SSE callback uses refs to read the
  // current filters/selection without resubscribing on every keystroke.
  const searchRef = useRef(search);
  const statusFilterRef = useRef(statusFilter);
  const viewRef = useRef(view);
  useEffect(() => {
    searchRef.current = search;
  }, [search]);
  useEffect(() => {
    statusFilterRef.current = statusFilter;
  }, [statusFilter]);
  useEffect(() => {
    viewRef.current = view;
  }, [view]);

  useEffect(() => {
    const controller = new AbortController();
    api
      .getRequests(controller.signal)
      .then((data) => setRequests(data))
      .catch((err) => {
        if (err?.name !== "AbortError") console.error(err);
      });

    const es = api.subscribeToEvents(
      (event) => {
        setLive(true);
        if (event.kind === "snapshot") {
          setRequests(event.data);
          return;
        }
        setRequests((prev) => [...event.data, ...prev]);
      },
      () => setLive(false)
    );

    return () => {
      controller.abort();
      es.close();
    };
  }, []);

  // Derive the request list for the current view + filters. Memoized so the
  // expensive filter pass doesn't run on every render.
  const filtered = useMemo(() => {
    let out = requests;
    if (view === "errors") {
      out = out.filter((r) => r.StatusCode >= 400);
    } else if (view === "slow") {
      out = out.filter((r) => r.Duration >= SLOW_MS);
    }
    if (statusFilter.size > 0) {
      out = out.filter((r) => {
        const k = statusKey(r.StatusCode);
        return k ? statusFilter.has(k) : false;
      });
    }
    if (search.trim()) {
      const q = search.trim().toLowerCase();
      out = out.filter((r) => r.Path.toLowerCase().includes(q));
    }
    return out;
  }, [requests, view, statusFilter, search]);

  const errorCount = useMemo(
    () => requests.filter((r) => r.StatusCode >= 400).length,
    [requests]
  );

  // Keep the selected request in sync with the visible list. If the user
  // switches views and the selected request is filtered out, drop the
  // selection so the right pane shows the empty state rather than a row
  // that doesn't appear anywhere.
  useEffect(() => {
    if (!selected) return;
    if (!filtered.some((r) => r.ID === selected.ID)) {
      setSelected(null);
    }
  }, [filtered, selected?.ID]);

  const handleClearAll = async () => {
    try {
      await api.clearRequests();
      setRequests([]);
      setSelected(null);
      setCompareIds([]);
    } catch (err) {
      console.error("Failed to clear requests:", err);
    }
  };

  const handleCompareToggle = (req: RequestLog) => {
    setCompareIds((prev) =>
      prev.includes(req.ID) ? prev.filter((id) => id !== req.ID) : [...prev, req.ID]
    );
  };

  const handleImport = (incoming: RequestLog[]) => {
    setRequests((prev) => {
      const merged = [...prev];
      const seen = new Set(merged.map((r) => r.ID));
      for (const r of incoming) {
        if (!seen.has(r.ID)) merged.push(r);
      }
      return merged;
    });
  };

  // The Inbox/Errors/Slow views share the three-pane layout. Analytics and
  // Environment use the full content area (rail + page) instead.
  const isListView =
    view === "inbox" || view === "errors" || view === "slow";

  return (
    <div class="h-screen bg-zinc-50 text-zinc-950 flex overflow-hidden">
      <RailNav active={view} onChange={setView} errorCount={errorCount} />

      {isListView ? (
        <Fragment>
          <RequestList
            title={titleFor(view)}
            subtitle={subtitleFor(view)}
            requests={filtered}
            selectedId={selected?.ID}
            onSelect={setSelected}
            statusFilter={statusFilter}
            onStatusFilterChange={setStatusFilter}
            search={search}
            onSearchChange={setSearch}
            live={live}
          />
          <DetailPane
            request={selected}
            onReplay={(r) => setReplayRequest(r)}
            onCompareAdd={handleCompareToggle}
            comparePending={selected ? compareIds.includes(selected.ID) : false}
          />
        </Fragment>
      ) : view === "analytics" ? (
        <Analytics
          requests={requests}
          onClearAll={handleClearAll}
          onImport={handleImport}
        />
      ) : (
        <EnvironmentView />
      )}

      {/* Floating bar for batch compare. Only appears when 2+ items are checked. */}
      {compareIds.length >= 2 && (
        <div class="fixed left-1/2 -translate-x-1/2 bottom-6 bg-zinc-900 text-white rounded-full shadow-xl px-5 py-2.5 flex items-center gap-4 text-sm z-40">
          <span class="font-medium">{compareIds.length} selected</span>
          <div class="w-px h-4 bg-white/20" />
          <button
            onClick={() => setShowComparison(true)}
            class="hover:text-zinc-200"
          >
            Compare
          </button>
          <button
            onClick={() => setCompareIds([])}
            class="text-zinc-400 hover:text-white text-xs"
          >
            Clear
          </button>
        </div>
      )}

      {showComparison && (
        <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-6">
          <div class="bg-white rounded-lg p-6 max-w-7xl w-full max-h-[90vh] overflow-auto">
            <RequestComparison
              requestIds={compareIds}
              allRequests={requests}
              onClose={() => {
                setShowComparison(false);
                setCompareIds([]);
              }}
            />
          </div>
        </div>
      )}

      {replayRequest && (
        <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-6">
          <div class="bg-white rounded-lg p-6 max-w-4xl w-full max-h-[90vh] overflow-auto">
            <RequestReplay
              request={replayRequest}
              onClose={() => setReplayRequest(null)}
            />
          </div>
        </div>
      )}
    </div>
  );
}

function statusKey(s: number): "2xx" | "3xx" | "4xx" | "5xx" | null {
  if (s >= 200 && s < 300) return "2xx";
  if (s >= 300 && s < 400) return "3xx";
  if (s >= 400 && s < 500) return "4xx";
  if (s >= 500) return "5xx";
  return null;
}

function titleFor(v: View): string {
  switch (v) {
    case "inbox":
      return "Inbox";
    case "errors":
      return "Errors";
    case "slow":
      return "Slow";
    case "analytics":
      return "Analytics";
    case "environment":
      return "Environment";
  }
}

function subtitleFor(v: View): string | undefined {
  switch (v) {
    case "errors":
      return "Status 4xx and 5xx";
    case "slow":
      return `Duration ≥ ${SLOW_MS}ms`;
    default:
      return undefined;
  }
}

function EnvironmentView() {
  return (
    <main class="flex-1 overflow-auto">
      <header class="px-8 pt-6 pb-4">
        <h1 class="text-2xl font-semibold tracking-tight">Environment</h1>
        <p class="text-sm text-zinc-500 mt-1">
          Server runtime and explicitly allowlisted environment variables.
        </p>
      </header>
      <div class="px-8 pb-8">
        <EnvironmentInfo />
      </div>
    </main>
  );
}
