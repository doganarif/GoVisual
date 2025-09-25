import { h } from "preact";
import { useEffect, useState } from "preact/hooks";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { Card, CardContent } from "./ui/card";
import { Badge } from "./ui/badge";
import { FlameGraph } from "./FlameGraph";
import { api, PerformanceMetrics, FlameGraphNode } from "../lib/api";

interface PerformanceProfilerProps {
  requestId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function PerformanceProfiler({
  requestId,
  open,
  onOpenChange,
}: PerformanceProfilerProps) {
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [flameGraphData, setFlameGraphData] = useState<FlameGraphNode | null>(
    null
  );
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (open && requestId) {
      loadMetrics(requestId);
    }
  }, [open, requestId]);

  const loadMetrics = async (id: string) => {
    setLoading(true);
    try {
      const metricsData = await api.getMetrics(id);
      setMetrics(metricsData);
    } catch (error) {
      console.error("Failed to load metrics:", error);
    } finally {
      setLoading(false);
    }
  };

  const loadFlameGraph = async () => {
    if (!requestId) return;
    try {
      const data = await api.getFlameGraph(requestId);
      setFlameGraphData(data);
    } catch (error) {
      console.error("Failed to load flame graph:", error);
    }
  };

  const formatDuration = (ns?: number): string => {
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

  if (loading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="max-w-6xl">
          <div className="flex items-center justify-center h-64">
            Loading performance metrics...
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  if (!metrics) {
    return null;
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Performance Profile</DialogTitle>
          <DialogDescription>
            Detailed performance analysis for request {requestId}
          </DialogDescription>
        </DialogHeader>

        {/* Metrics Summary Cards */}
        <div className="grid grid-cols-4 gap-4 mt-4">
          <Card>
            <CardContent className="p-4">
              <div className="text-sm text-muted-foreground">CPU Time</div>
              <div className="text-2xl font-bold">
                {formatDuration(metrics.cpu_time)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-sm text-muted-foreground">Memory</div>
              <div className="text-2xl font-bold">
                {formatBytes(metrics.memory_alloc)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-sm text-muted-foreground">Goroutines</div>
              <div className="text-2xl font-bold">
                {metrics.num_goroutines || 0}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-sm text-muted-foreground">GC Pauses</div>
              <div className="text-2xl font-bold">
                {formatDuration(metrics.gc_pause_total)}
              </div>
            </CardContent>
          </Card>
        </div>

        <Tabs
          defaultValue="bottlenecks"
          className="mt-6"
          onValueChange={(value) => {
            if (value === "flamegraph") {
              loadFlameGraph();
            }
          }}
        >
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="bottlenecks">Bottlenecks</TabsTrigger>
            <TabsTrigger value="flamegraph">Flame Graph</TabsTrigger>
            <TabsTrigger value="sql">SQL Queries</TabsTrigger>
            <TabsTrigger value="http">HTTP Calls</TabsTrigger>
            <TabsTrigger value="functions">Functions</TabsTrigger>
          </TabsList>

          <TabsContent value="bottlenecks" className="space-y-3">
            {metrics.bottlenecks && metrics.bottlenecks.length > 0 ? (
              metrics.bottlenecks.map((bottleneck, idx) => (
                <Card key={idx}>
                  <CardContent className="p-4">
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-1">
                          <Badge
                            variant={
                              bottleneck.impact > 0.5
                                ? "outline"
                                : bottleneck.impact > 0.3
                                ? "secondary"
                                : "default"
                            }
                          >
                            {bottleneck.type.toUpperCase()}
                          </Badge>
                          <span className="font-medium">
                            {bottleneck.description}
                          </span>
                        </div>
                        <p className="text-sm text-muted-foreground">
                          {bottleneck.suggestion}
                        </p>
                      </div>
                      <div className="text-right">
                        <div className="text-lg font-bold">
                          {(bottleneck.impact * 100).toFixed(1)}%
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {formatDuration(bottleneck.duration)}
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))
            ) : (
              <p className="text-center text-muted-foreground">
                No bottlenecks detected
              </p>
            )}
          </TabsContent>

          <TabsContent value="flamegraph">
            <div className="w-full overflow-x-auto">
              <FlameGraph data={flameGraphData} width={900} height={400} />
            </div>
          </TabsContent>

          <TabsContent value="sql">
            {metrics.sql_queries && metrics.sql_queries.length > 0 ? (
              <div className="rounded-md border">
                <table className="w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="p-2 text-left">Query</th>
                      <th className="p-2 text-left">Duration</th>
                      <th className="p-2 text-left">Rows</th>
                    </tr>
                  </thead>
                  <tbody>
                    {metrics.sql_queries.map((query, idx) => (
                      <tr key={idx} className="border-b">
                        <td className="p-2 font-mono text-sm">{query.query}</td>
                        <td className="p-2">
                          {formatDuration(query.duration)}
                        </td>
                        <td className="p-2">{query.rows}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-center text-muted-foreground">
                No SQL queries recorded
              </p>
            )}
          </TabsContent>

          <TabsContent value="http">
            {metrics.http_calls && metrics.http_calls.length > 0 ? (
              <div className="rounded-md border">
                <table className="w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="p-2 text-left">Method</th>
                      <th className="p-2 text-left">URL</th>
                      <th className="p-2 text-left">Status</th>
                      <th className="p-2 text-left">Duration</th>
                      <th className="p-2 text-left">Size</th>
                    </tr>
                  </thead>
                  <tbody>
                    {metrics.http_calls.map((call, idx) => (
                      <tr key={idx} className="border-b">
                        <td className="p-2">{call.method}</td>
                        <td className="p-2 font-mono text-sm">{call.url}</td>
                        <td className="p-2">
                          <Badge
                            variant={
                              call.status >= 200 && call.status < 300
                                ? "default"
                                : "outline"
                            }
                          >
                            {call.status}
                          </Badge>
                        </td>
                        <td className="p-2">{formatDuration(call.duration)}</td>
                        <td className="p-2">{formatBytes(call.size)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-center text-muted-foreground">
                No HTTP calls recorded
              </p>
            )}
          </TabsContent>

          <TabsContent value="functions">
            {metrics.function_timings &&
            Object.keys(metrics.function_timings).length > 0 ? (
              <div className="rounded-md border">
                <table className="w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="p-2 text-left">Function</th>
                      <th className="p-2 text-left">Duration</th>
                      <th className="p-2 text-left">Percentage</th>
                    </tr>
                  </thead>
                  <tbody>
                    {Object.entries(metrics.function_timings)
                      .sort(([, a], [, b]) => b - a)
                      .map(([name, duration], idx) => {
                        const percentage = (
                          (duration / metrics.duration) *
                          100
                        ).toFixed(2);
                        return (
                          <tr key={idx} className="border-b">
                            <td className="p-2 font-mono text-sm">{name}</td>
                            <td className="p-2">{formatDuration(duration)}</td>
                            <td className="p-2">
                              <div className="flex items-center gap-2">
                                <div className="flex-1 bg-gray-200 rounded-full h-2">
                                  <div
                                    className="bg-primary h-2 rounded-full"
                                    style={{ width: `${percentage}%` }}
                                  />
                                </div>
                                <span className="text-sm">{percentage}%</span>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-center text-muted-foreground">
                No function timings recorded
              </p>
            )}
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

function cn(...classes: (string | boolean | undefined)[]): string {
  return classes.filter(Boolean).join(" ");
}
