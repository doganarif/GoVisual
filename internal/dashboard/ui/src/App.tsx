import { h } from "preact";
import { useEffect, useState } from "preact/hooks";
import { api, RequestLog } from "./lib/api";
import { SimpleSidebar } from "./components/SimpleSidebar";
import { StatsDashboard } from "./components/StatsDashboard";
import { RequestTable } from "./components/RequestTable";
import { RequestDrawer } from "./components/RequestDrawer";
import { EnvironmentInfo } from "./components/EnvironmentInfo";
import { Filters, FilterState } from "./components/Filters";
import { RequestComparison } from "./components/RequestComparison";
import { RequestReplay } from "./components/RequestReplay";
import { RequestTrace } from "./components/RequestTrace";
import { ExportImport } from "./components/ExportImport";
import { ResponseTimeChart } from "./components/ResponseTimeChart";
import { Button } from "./components/ui/button";
import { cn } from "./lib/utils";

export function App() {
  const [requests, setRequests] = useState<RequestLog[]>([]);
  const [filteredRequests, setFilteredRequests] = useState<RequestLog[]>([]);
  const [selectedRequest, setSelectedRequest] = useState<RequestLog | null>(
    null
  );
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [activeTab, setActiveTab] = useState("dashboard");
  const [filters, setFilters] = useState<FilterState>({
    method: "",
    statusCode: "",
    path: "",
    minDuration: "",
  });
  const [selectedForComparison, setSelectedForComparison] = useState<string[]>(
    []
  );
  const [showComparison, setShowComparison] = useState(false);
  const [showReplay, setShowReplay] = useState(false);
  const [replayRequest, setReplayRequest] = useState<RequestLog | null>(null);

  useEffect(() => {
    // Load initial requests
    api
      .getRequests()
      .then((data) => {
        setRequests(data);
        setFilteredRequests(data);
      })
      .catch(console.error);

    // Subscribe to live updates
    const eventSource = api.subscribeToEvents((data) => {
      setRequests(data);
      applyFilters(data, filters);
    });

    return () => {
      eventSource.close();
    };
  }, []);

  const applyFilters = (
    requestList: RequestLog[],
    filterState: FilterState
  ) => {
    let filtered = [...requestList];

    // Filter by method
    if (filterState.method) {
      filtered = filtered.filter((r) => r.Method === filterState.method);
    }

    // Filter by status code
    if (filterState.statusCode) {
      const statusPrefix = filterState.statusCode.charAt(0);
      filtered = filtered.filter((r) => {
        const statusStr = r.StatusCode.toString();
        return statusStr.charAt(0) === statusPrefix;
      });
    }

    // Filter by path
    if (filterState.path) {
      const searchPath = filterState.path.toLowerCase();
      filtered = filtered.filter((r) =>
        r.Path.toLowerCase().includes(searchPath)
      );
    }

    // Filter by minimum duration
    if (filterState.minDuration) {
      const minDur = parseInt(filterState.minDuration);
      if (!isNaN(minDur)) {
        filtered = filtered.filter((r) => r.Duration >= minDur);
      }
    }

    setFilteredRequests(filtered);
  };

  const handleFilterChange = (newFilters: FilterState) => {
    setFilters(newFilters);
    applyFilters(requests, newFilters);
  };

  const handleClearRequests = async () => {
    try {
      await api.clearRequests();
      setRequests([]);
      setFilteredRequests([]);
      setSelectedRequest(null);
      setSelectedForComparison([]);
    } catch (error) {
      console.error("Failed to clear requests:", error);
    }
  };

  const handleRequestSelect = (request: RequestLog) => {
    setSelectedRequest(request);
    setDrawerOpen(true);
  };

  const handleToggleComparison = (requestId: string) => {
    setSelectedForComparison((prev) => {
      if (prev.includes(requestId)) {
        return prev.filter((id) => id !== requestId);
      }
      return [...prev, requestId];
    });
  };

  const handleStartComparison = () => {
    if (selectedForComparison.length >= 2) {
      setShowComparison(true);
    }
  };

  const handleStartReplay = (request: RequestLog) => {
    setReplayRequest(request);
    setShowReplay(true);
  };

  const handleImportRequests = (importedRequests: RequestLog[]) => {
    const combined = [...requests, ...importedRequests];
    const uniqueRequests = Array.from(
      new Map(combined.map((r) => [r.ID, r])).values()
    );
    setRequests(uniqueRequests);
    applyFilters(uniqueRequests, filters);
  };

  // Calculate stats
  const calculateStats = () => {
    const total = requests.length;
    const success = requests.filter(
      (r) => r.StatusCode >= 200 && r.StatusCode < 300
    ).length;
    const successRate = total > 0 ? Math.round((success / total) * 100) : 0;
    const avgDuration =
      total > 0
        ? Math.round(requests.reduce((sum, r) => sum + r.Duration, 0) / total)
        : 0;

    return { total, successRate, avgDuration };
  };

  const stats = calculateStats();

  const renderContent = () => {
    switch (activeTab) {
      case "dashboard":
        return (
          <div className="space-y-8 animate-in fade-in-50 duration-500">
            <div className="mb-2">
              <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
              <p className="text-muted-foreground mt-1">
                Monitor and analyze HTTP requests in real-time
              </p>
            </div>

            <StatsDashboard requests={requests} />

            <Filters
              onFilterChange={handleFilterChange}
              onClear={handleClearRequests}
            />

            <div className="bg-background rounded-xl shadow-sm border overflow-hidden">
              <div className="px-6 py-4 border-b bg-muted/30">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">Recent Requests</h3>
                  <span className="text-sm text-muted-foreground">
                    {filteredRequests.length} total requests
                  </span>
                </div>
              </div>
              <div className="overflow-auto max-h-[500px]">
                <RequestTable
                  requests={filteredRequests.slice(0, 50)}
                  selectedRequest={selectedRequest}
                  onRequestSelect={handleRequestSelect}
                />
              </div>
            </div>
          </div>
        );

      case "requests":
        return (
          <div className="space-y-8 animate-in fade-in-50 duration-500">
            <div className="mb-2">
              <h1 className="text-3xl font-bold tracking-tight">
                All Requests
              </h1>
              <p className="text-muted-foreground mt-1">
                View and filter all captured HTTP requests
              </p>
            </div>

            <Filters
              onFilterChange={handleFilterChange}
              onClear={handleClearRequests}
            />

            <div className="bg-background rounded-xl shadow-sm border overflow-hidden">
              <div className="px-6 py-4 border-b bg-muted/30">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">Request Log</h3>
                  <div className="flex items-center gap-4">
                    {selectedForComparison.length > 0 && (
                      <Button
                        onClick={handleStartComparison}
                        disabled={selectedForComparison.length < 2}
                        size="sm"
                      >
                        Compare {selectedForComparison.length} Selected
                      </Button>
                    )}
                    <span className="text-sm text-muted-foreground">
                      Showing {filteredRequests.length} of {requests.length}{" "}
                      requests
                    </span>
                  </div>
                </div>
              </div>
              <div
                className="overflow-auto"
                style={{ height: "calc(100vh - 380px)" }}
              >
                <RequestTable
                  requests={filteredRequests}
                  selectedRequest={selectedRequest}
                  onRequestSelect={handleRequestSelect}
                  selectedForComparison={selectedForComparison}
                  onToggleComparison={handleToggleComparison}
                  onReplay={handleStartReplay}
                />
              </div>
            </div>
          </div>
        );

      case "environment":
        return (
          <div className="space-y-8 animate-in fade-in-50 duration-500">
            <div className="mb-2">
              <h1 className="text-3xl font-bold tracking-tight">Environment</h1>
              <p className="text-muted-foreground mt-1">
                System information and environment variables
              </p>
            </div>
            <EnvironmentInfo />
          </div>
        );

      case "trace":
        return (
          <div className="space-y-8 animate-in fade-in-50 duration-500">
            <div className="mb-2">
              <h1 className="text-3xl font-bold tracking-tight">
                Request Trace
              </h1>
              <p className="text-muted-foreground mt-1">
                Analyze request execution flow, middleware chain, SQL queries,
                and HTTP calls
              </p>
            </div>

            {selectedRequest ? (
              <RequestTrace request={selectedRequest} />
            ) : (
              <div>
                <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                  <p className="text-sm text-blue-800">
                    Select a request from the table below to view its execution
                    trace, including middleware execution, SQL queries, and
                    external HTTP calls.
                  </p>
                </div>
                <div className="bg-background rounded-xl shadow-sm border overflow-hidden">
                  <div className="px-6 py-4 border-b bg-muted/30">
                    <h3 className="text-lg font-semibold">
                      Select a Request to Trace
                    </h3>
                  </div>
                  <div
                    className="overflow-auto"
                    style={{ height: "calc(100vh - 380px)" }}
                  >
                    <RequestTable
                      requests={filteredRequests.slice(0, 100)}
                      selectedRequest={selectedRequest}
                      onRequestSelect={handleRequestSelect}
                      selectedForComparison={selectedForComparison}
                      onToggleComparison={handleToggleComparison}
                      onReplay={handleStartReplay}
                    />
                  </div>
                </div>
              </div>
            )}
          </div>
        );

      case "analytics":
        return (
          <div className="space-y-8 animate-in fade-in-50 duration-500">
            <div className="mb-2">
              <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
              <p className="text-muted-foreground mt-1">
                Performance metrics and request analysis
              </p>
            </div>

            <ResponseTimeChart requests={requests} />

            <ExportImport
              requests={filteredRequests}
              onImport={handleImportRequests}
            />
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="flex h-screen bg-gradient-to-br from-background to-muted/20">
      {/* Sidebar */}
      <SimpleSidebar
        activeTab={activeTab}
        onTabChange={setActiveTab}
        stats={stats}
        onClearAll={handleClearRequests}
      />

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        <div className="p-8">{renderContent()}</div>
      </main>

      {/* Request Drawer */}
      <RequestDrawer
        request={selectedRequest}
        open={drawerOpen}
        onOpenChange={setDrawerOpen}
      />

      {/* Comparison Modal */}
      {showComparison && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background rounded-lg p-6 max-w-7xl max-h-[90vh] overflow-auto">
            <RequestComparison
              requestIds={selectedForComparison}
              allRequests={requests}
              onClose={() => {
                setShowComparison(false);
                setSelectedForComparison([]);
              }}
            />
          </div>
        </div>
      )}

      {/* Replay Modal */}
      {showReplay && replayRequest && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-background rounded-lg p-6 max-w-4xl max-h-[90vh] overflow-auto">
            <RequestReplay
              request={replayRequest}
              onClose={() => {
                setShowReplay(false);
                setReplayRequest(null);
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
