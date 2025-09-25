import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Drawer, DrawerContent } from "./ui/drawer";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { Badge } from "./ui/badge";
import { Card, CardContent } from "./ui/card";
import { Button } from "./ui/button";
import { FlameGraph } from "./FlameGraph";
import {
  RequestLog,
  api,
  PerformanceMetrics,
  FlameGraphNode,
} from "../lib/api";
import { cn } from "../lib/utils";

interface RequestDrawerProps {
  request: RequestLog | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RequestDrawer({
  request,
  open,
  onOpenChange,
}: RequestDrawerProps) {
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [flameGraphData, setFlameGraphData] = useState<FlameGraphNode | null>(
    null
  );
  const [loadingMetrics, setLoadingMetrics] = useState(false);
  const [activeTab, setActiveTab] = useState("overview");

  useEffect(() => {
    if (open && request?.ID) {
      loadMetrics(request.ID);
      setActiveTab("overview");
    }
  }, [open, request?.ID]);

  const loadMetrics = async (id: string) => {
    setLoadingMetrics(true);
    try {
      const metricsData = await api.getMetrics(id);
      setMetrics(metricsData);
    } catch (error) {
      console.error("Failed to load metrics:", error);
      setMetrics(null);
    } finally {
      setLoadingMetrics(false);
    }
  };

  const loadFlameGraph = async () => {
    if (!request?.ID) return;
    try {
      const data = await api.getFlameGraph(request.ID);
      setFlameGraphData(data);
    } catch (error) {
      console.error("Failed to load flame graph:", error);
    }
  };

  const formatHeaders = (headers: Record<string, string[]>): string => {
    if (!headers) return "No headers";
    return Object.entries(headers)
      .map(([key, values]) => `${key}: ${values.join(", ")}`)
      .join("\n");
  };

  const formatBody = (body?: string): string => {
    if (!body) return "No body";
    try {
      const parsed = JSON.parse(body);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return body;
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1) return "<1ms";
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatDurationNs = (ns?: number): string => {
    if (!ns) return "0ms";
    const ms = ns / 1000000;
    if (ms < 1) return Math.round(ns / 1000) + "μs";
    if (ms < 1000) return ms.toFixed(2) + "ms";
    return (ms / 1000).toFixed(2) + "s";
  };

  const formatBytes = (bytes?: number): string => {
    if (!bytes) return "0 B";
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + " " + sizes[i];
  };

  const getStatusVariant = (
    status: number
  ): "default" | "secondary" | "outline" => {
    if (status >= 200 && status < 300) return "default";
    if (status >= 300 && status < 400) return "secondary";
    return "outline";
  };

  if (!request) return null;

  const hasPerformanceMetrics = !!request.PerformanceMetrics || !!metrics;

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent
        onClose={() => onOpenChange(false)}
        isFullscreen={isFullscreen}
        onToggleFullscreen={() => setIsFullscreen(!isFullscreen)}
      >
        {/* Request Summary Bar */}
        <div className="bg-muted/30 rounded-lg p-4 mb-6 animate-in fade-in-50 duration-300">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Badge className="px-3 py-1" variant="secondary">
                {request.Method}
              </Badge>
              <span className="font-mono text-sm">{request.Path}</span>
              <Badge variant={getStatusVariant(request.StatusCode)}>
                {request.StatusCode}
              </Badge>
            </div>
            <div className="flex items-center gap-6 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <svg
                  className="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                {formatDuration(request.Duration)}
              </span>
              <span>{new Date(request.Timestamp).toLocaleString()}</span>
            </div>
          </div>
        </div>

        {/* Main Content Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-5 mb-6">
            <TabsTrigger
              value="overview"
              className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
            >
              Overview
            </TabsTrigger>
            <TabsTrigger value="headers">Headers</TabsTrigger>
            <TabsTrigger value="body">Body</TabsTrigger>
            <TabsTrigger value="trace">Trace</TabsTrigger>
            {hasPerformanceMetrics && (
              <TabsTrigger value="performance">Performance</TabsTrigger>
            )}
          </TabsList>

          {/* Overview Tab */}
          <TabsContent
            value="overview"
            className="space-y-6 animate-in fade-in-50 duration-300"
          >
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <Card>
                <CardContent className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">
                    Request ID
                  </p>
                  <p className="font-mono text-xs truncate">{request.ID}</p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">Duration</p>
                  <p className="text-lg font-semibold">
                    {formatDuration(request.Duration)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">
                    Response Size
                  </p>
                  <p className="text-lg font-semibold">
                    {formatBytes(request.ResponseBody?.length || 0)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardContent className="p-4">
                  <p className="text-xs text-muted-foreground mb-1">
                    Query String
                  </p>
                  <p className="font-mono text-xs truncate">
                    {request.Query || "No query"}
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Quick Timeline */}
            <Card>
              <CardContent className="p-6">
                <h4 className="text-sm font-medium mb-4">Request Timeline</h4>
                <div className="relative h-12 bg-muted rounded-lg overflow-hidden">
                  <div
                    className="absolute h-full bg-gradient-to-r from-primary/20 to-primary/60"
                    style={{ width: "100%" }}
                  >
                    <div className="h-full flex items-center px-3">
                      <span className="text-xs text-foreground/70">
                        {formatDuration(request.Duration)}
                      </span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {request.Error && (
              <Card className="border-destructive/50 bg-destructive/5">
                <CardContent className="p-4">
                  <div className="flex items-start gap-2">
                    <svg
                      className="w-5 h-5 text-destructive mt-0.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    <div className="flex-1">
                      <p className="font-medium text-destructive mb-1">
                        Error Occurred
                      </p>
                      <p className="text-sm text-muted-foreground">
                        {request.Error}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* Headers Tab */}
          <TabsContent
            value="headers"
            className="space-y-6 animate-in fade-in-50 duration-300"
          >
            <Card>
              <CardContent className="p-6">
                <h4 className="text-sm font-medium mb-3">Request Headers</h4>
                <pre className="bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto font-mono">
                  {formatHeaders(request.RequestHeaders)}
                </pre>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-6">
                <h4 className="text-sm font-medium mb-3">Response Headers</h4>
                <pre className="bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto font-mono">
                  {formatHeaders(request.ResponseHeaders)}
                </pre>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Body Tab */}
          <TabsContent
            value="body"
            className="space-y-6 animate-in fade-in-50 duration-300"
          >
            <Card>
              <CardContent className="p-6">
                <h4 className="text-sm font-medium mb-3">Request Body</h4>
                <pre className="bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto max-h-96 font-mono">
                  {formatBody(request.RequestBody)}
                </pre>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-6">
                <h4 className="text-sm font-medium mb-3">Response Body</h4>
                <pre className="bg-muted/50 p-4 rounded-lg text-xs overflow-x-auto max-h-96 font-mono">
                  {formatBody(request.ResponseBody)}
                </pre>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Trace Tab */}
          <TabsContent
            value="trace"
            className="space-y-6 animate-in fade-in-50 duration-300"
          >
            {request.MiddlewareTrace && request.MiddlewareTrace.length > 0 && (
              <Card>
                <CardContent className="p-6">
                  <h4 className="text-sm font-medium mb-4">
                    Middleware Execution
                  </h4>
                  <div className="space-y-2">
                    {request.MiddlewareTrace.map((middleware, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between p-3 bg-muted/30 rounded-lg hover:bg-muted/50 transition-colors"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium">
                            {index + 1}
                          </div>
                          <span className="font-mono text-sm">
                            {middleware.name || `Middleware ${index + 1}`}
                          </span>
                          {middleware.type && (
                            <Badge className="text-xs" variant="secondary">
                              {middleware.type}
                            </Badge>
                          )}
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                          <span>
                            {formatDuration(middleware.duration || 0)}
                          </span>
                          <Badge
                            variant={
                              middleware.status === "completed"
                                ? "default"
                                : "outline"
                            }
                          >
                            {middleware.status || "completed"}
                          </Badge>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}

            {request.PerformanceMetrics?.sql_queries &&
              request.PerformanceMetrics.sql_queries.length > 0 && (
                <Card>
                  <CardContent className="p-6">
                    <h4 className="text-sm font-medium mb-4">
                      SQL Queries (
                      {request.PerformanceMetrics.sql_queries.length})
                    </h4>
                    <div className="space-y-3">
                      {request.PerformanceMetrics.sql_queries
                        .slice(0, 5)
                        .map((query, index) => (
                          <div
                            key={index}
                            className="p-3 bg-muted/30 rounded-lg"
                          >
                            <pre className="text-xs font-mono overflow-x-auto">
                              {query.query}
                            </pre>
                            <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                              <span>
                                Duration: {formatDurationNs(query.duration)}
                              </span>
                              <span>Rows: {query.rows}</span>
                              {query.error && (
                                <span className="text-red-600">
                                  Error: {query.error}
                                </span>
                              )}
                            </div>
                          </div>
                        ))}
                    </div>
                  </CardContent>
                </Card>
              )}

            {request.PerformanceMetrics?.http_calls &&
              request.PerformanceMetrics.http_calls.length > 0 && (
                <Card>
                  <CardContent className="p-6">
                    <h4 className="text-sm font-medium mb-4">
                      HTTP Calls ({request.PerformanceMetrics.http_calls.length}
                      )
                    </h4>
                    <div className="space-y-3">
                      {request.PerformanceMetrics.http_calls
                        .slice(0, 5)
                        .map((call, index) => (
                          <div
                            key={index}
                            className="flex items-center justify-between p-3 bg-muted/30 rounded-lg"
                          >
                            <div className="flex items-center gap-3">
                              <Badge variant="outline">{call.method}</Badge>
                              <span className="text-sm font-mono">
                                {call.url}
                              </span>
                            </div>
                            <div className="flex items-center gap-4 text-sm text-muted-foreground">
                              <span>Status: {call.status}</span>
                              <span>{formatDurationNs(call.duration)}</span>
                            </div>
                          </div>
                        ))}
                    </div>
                  </CardContent>
                </Card>
              )}

            {request.RouteTrace && (
              <Card>
                <CardContent className="p-6">
                  <h4 className="text-sm font-medium mb-4">
                    Route Information
                  </h4>
                  <div className="space-y-3">
                    <div className="flex justify-between py-2">
                      <span className="text-sm text-muted-foreground">
                        Pattern
                      </span>
                      <span className="font-mono text-sm">
                        {request.RouteTrace.pattern || request.Path}
                      </span>
                    </div>
                    <div className="flex justify-between py-2">
                      <span className="text-sm text-muted-foreground">
                        Handler
                      </span>
                      <span className="font-mono text-sm">
                        {request.RouteTrace.handler || "DefaultHandler"}
                      </span>
                    </div>
                    <div className="flex justify-between py-2">
                      <span className="text-sm text-muted-foreground">
                        Match Time
                      </span>
                      <span className="font-mono text-sm">
                        {formatDuration(request.RouteTrace.matchTime || 0)}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </TabsContent>

          {/* Performance Tab */}
          {hasPerformanceMetrics && (
            <TabsContent
              value="performance"
              className="space-y-6 animate-in fade-in-50 duration-300"
              onFocus={() => {
                if (activeTab === "performance" && !flameGraphData) {
                  loadFlameGraph();
                }
              }}
            >
              {loadingMetrics ? (
                <div className="flex items-center justify-center h-64">
                  <div className="text-muted-foreground">
                    Loading performance metrics...
                  </div>
                </div>
              ) : metrics ? (
                <>
                  {/* Performance Summary */}
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <Card>
                      <CardContent className="p-4">
                        <p className="text-xs text-muted-foreground mb-1">
                          CPU Time
                        </p>
                        <p className="text-lg font-semibold">
                          {formatDurationNs(metrics.cpu_time)}
                        </p>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardContent className="p-4">
                        <p className="text-xs text-muted-foreground mb-1">
                          Memory
                        </p>
                        <p className="text-lg font-semibold">
                          {formatBytes(metrics.memory_alloc)}
                        </p>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardContent className="p-4">
                        <p className="text-xs text-muted-foreground mb-1">
                          Goroutines
                        </p>
                        <p className="text-lg font-semibold">
                          {metrics.num_goroutines || 0}
                        </p>
                      </CardContent>
                    </Card>
                    <Card>
                      <CardContent className="p-4">
                        <p className="text-xs text-muted-foreground mb-1">
                          GC Pauses
                        </p>
                        <p className="text-lg font-semibold">
                          {formatDurationNs(metrics.gc_pause_total)}
                        </p>
                      </CardContent>
                    </Card>
                  </div>

                  {/* Bottlenecks */}
                  {metrics.bottlenecks && metrics.bottlenecks.length > 0 && (
                    <Card>
                      <CardContent className="p-6">
                        <h4 className="text-sm font-medium mb-4">
                          Performance Bottlenecks
                        </h4>
                        <div className="space-y-3">
                          {metrics.bottlenecks.map((bottleneck, idx) => (
                            <div
                              key={idx}
                              className="p-4 bg-muted/30 rounded-lg hover:bg-muted/50 transition-all duration-200 hover:shadow-sm"
                            >
                              <div className="flex justify-between items-start">
                                <div className="flex-1">
                                  <div className="flex items-center gap-2 mb-2">
                                    <Badge variant="secondary">
                                      {bottleneck.type.toUpperCase()}
                                    </Badge>
                                    <span className="font-medium text-sm">
                                      {bottleneck.description}
                                    </span>
                                  </div>
                                  <p className="text-xs text-muted-foreground">
                                    {bottleneck.suggestion}
                                  </p>
                                </div>
                                <div className="text-right ml-4">
                                  <div className="text-lg font-bold">
                                    {(bottleneck.impact * 100).toFixed(1)}%
                                  </div>
                                  <div className="text-xs text-muted-foreground">
                                    {formatDurationNs(bottleneck.duration)}
                                  </div>
                                </div>
                              </div>
                            </div>
                          ))}
                        </div>
                      </CardContent>
                    </Card>
                  )}

                  {/* Flame Graph */}
                  {flameGraphData && (
                    <Card>
                      <CardContent className="p-6">
                        <h4 className="text-sm font-medium mb-4">
                          CPU Profile Flame Graph
                        </h4>
                        <div className="w-full overflow-x-auto">
                          <FlameGraph
                            data={flameGraphData}
                            width={900}
                            height={400}
                          />
                        </div>
                      </CardContent>
                    </Card>
                  )}
                </>
              ) : null}
            </TabsContent>
          )}
        </Tabs>
      </DrawerContent>
    </Drawer>
  );
}
