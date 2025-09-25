import { h } from "preact";
import { useState, useMemo } from "preact/hooks";
import { RequestLog } from "../lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { cn } from "../lib/utils";

interface TraceEntry {
  name: string;
  type: "middleware" | "handler" | "sql" | "http" | "custom";
  start_time: string;
  end_time: string;
  duration: number;
  status: "running" | "completed" | "error";
  error?: string;
  details?: any;
  children?: TraceEntry[];
}

interface RequestTraceProps {
  request: RequestLog;
}

export function RequestTrace({ request }: RequestTraceProps) {
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [selectedTrace, setSelectedTrace] = useState<TraceEntry | null>(null);

  // Parse middleware traces
  const traces = useMemo(() => {
    if (!request.MiddlewareTrace || request.MiddlewareTrace.length === 0) {
      return [];
    }

    return request.MiddlewareTrace.map((trace) => ({
      name: trace.name || "Unknown",
      type: trace.type || "middleware",
      start_time: trace.start_time || "",
      end_time: trace.end_time || "",
      duration: trace.duration || 0,
      status: trace.status || "completed",
      error: trace.error,
      details: trace.details,
      children: trace.children || [],
    })) as TraceEntry[];
  }, [request.MiddlewareTrace]);

  // Parse SQL queries from performance metrics
  const sqlQueries = useMemo(() => {
    if (!request.PerformanceMetrics?.sql_queries) {
      return [];
    }
    return request.PerformanceMetrics.sql_queries;
  }, [request.PerformanceMetrics]);

  // Parse HTTP calls from performance metrics
  const httpCalls = useMemo(() => {
    if (!request.PerformanceMetrics?.http_calls) {
      return [];
    }
    return request.PerformanceMetrics.http_calls;
  }, [request.PerformanceMetrics]);

  const toggleExpand = (nodeId: string) => {
    setExpandedNodes((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(nodeId)) {
        newSet.delete(nodeId);
      } else {
        newSet.add(nodeId);
      }
      return newSet;
    });
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case "middleware":
        return "text-blue-600 bg-blue-50";
      case "handler":
        return "text-green-600 bg-green-50";
      case "sql":
        return "text-purple-600 bg-purple-50";
      case "http":
        return "text-orange-600 bg-orange-50";
      case "custom":
        return "text-gray-600 bg-gray-50";
      default:
        return "text-gray-600 bg-gray-50";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return "text-green-600";
      case "error":
        return "text-red-600";
      case "running":
        return "text-yellow-600";
      default:
        return "text-gray-600";
    }
  };

  const formatDuration = (ms: number) => {
    if (ms < 1) return "<1ms";
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const renderTraceNode = (trace: TraceEntry, depth = 0, nodeId = "0") => {
    const hasChildren = trace.children && trace.children.length > 0;
    const isExpanded = expandedNodes.has(nodeId);

    return (
      <div key={nodeId} className="border-l-2 border-gray-200">
        <div
          className={cn(
            "flex items-center gap-2 p-2 hover:bg-gray-50 cursor-pointer",
            depth > 0 && "ml-4"
          )}
          onClick={() => {
            setSelectedTrace(trace);
            if (hasChildren) {
              toggleExpand(nodeId);
            }
          }}
        >
          {hasChildren && (
            <span className="text-gray-400 text-sm">
              {isExpanded ? "▼" : "▶"}
            </span>
          )}
          <Badge className={cn("text-xs", getTypeColor(trace.type))}>
            {trace.type}
          </Badge>
          <span className="flex-1 text-sm font-medium">{trace.name}</span>
          <span className={cn("text-xs", getStatusColor(trace.status))}>
            {trace.status}
          </span>
          <span className="text-xs text-gray-500">
            {formatDuration(trace.duration)}
          </span>
        </div>

        {hasChildren && isExpanded && (
          <div className="ml-2">
            {trace.children!.map((child, idx) =>
              renderTraceNode(child, depth + 1, `${nodeId}-${idx}`)
            )}
          </div>
        )}
      </div>
    );
  };

  const renderSQLQuery = (query: any, index: number) => {
    return (
      <div key={index} className="border rounded-lg p-3 mb-2">
        <div className="flex items-center justify-between mb-2">
          <Badge className="text-xs bg-purple-100 text-purple-800">
            SQL Query #{index + 1}
          </Badge>
          <span className="text-xs text-gray-500">
            {formatDuration(query.duration)}
          </span>
        </div>
        <pre className="text-xs bg-gray-50 p-2 rounded overflow-x-auto font-mono">
          {query.query}
        </pre>
        <div className="flex items-center gap-4 mt-2 text-xs text-gray-500">
          <span>Rows: {query.rows || 0}</span>
          {query.error && (
            <span className="text-red-600">Error: {query.error}</span>
          )}
        </div>
      </div>
    );
  };

  const renderHTTPCall = (call: any, index: number) => {
    return (
      <div key={index} className="border rounded-lg p-3 mb-2">
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <Badge className="text-xs bg-orange-100 text-orange-800">
              {call.method}
            </Badge>
            <span className="text-sm font-medium">{call.url}</span>
          </div>
          <span className="text-xs text-gray-500">
            {formatDuration(call.duration)}
          </span>
        </div>
        <div className="flex items-center gap-4 text-xs text-gray-500">
          <span>Status: {call.status}</span>
          <span>Size: {call.size} bytes</span>
        </div>
      </div>
    );
  };

  const renderTimeline = () => {
    if (traces.length === 0) {
      return (
        <div className="text-center py-8 text-gray-500">
          No trace data available for this request
        </div>
      );
    }

    // Calculate timeline positions
    const allEvents: any[] = [];

    // Add middleware traces
    const addTraceEvents = (trace: TraceEntry, parent = "") => {
      allEvents.push({
        type: "trace",
        name: trace.name,
        traceType: trace.type,
        startTime: new Date(trace.start_time).getTime(),
        endTime: new Date(trace.end_time).getTime(),
        duration: trace.duration,
        status: trace.status,
        parent,
      });

      if (trace.children) {
        trace.children.forEach((child) => addTraceEvents(child, trace.name));
      }
    };

    traces.forEach((trace) => addTraceEvents(trace));

    // Sort by start time
    allEvents.sort((a, b) => a.startTime - b.startTime);

    if (allEvents.length === 0) {
      return (
        <div className="text-center py-8 text-gray-500">
          No timeline data available
        </div>
      );
    }

    const minTime = allEvents[0].startTime;
    const maxTime = Math.max(...allEvents.map((e) => e.endTime || e.startTime));
    const totalDuration = maxTime - minTime || 1;

    return (
      <div className="relative">
        {allEvents.map((event, idx) => {
          const leftPercent =
            ((event.startTime - minTime) / totalDuration) * 100;
          const widthPercent = (event.duration / totalDuration) * 100;

          return (
            <div key={idx} className="relative h-8 mb-1">
              <div className="absolute inset-y-0 left-0 w-32 pr-2 text-right">
                <span className="text-xs truncate">{event.name}</span>
              </div>
              <div className="absolute inset-y-0 left-32 right-0">
                <div
                  className={cn(
                    "absolute h-6 top-1 rounded",
                    getTypeColor(event.traceType),
                    event.status === "error" && "border-2 border-red-500"
                  )}
                  style={{
                    left: `${leftPercent}%`,
                    width: `${Math.max(widthPercent, 1)}%`,
                  }}
                  title={`${event.name}: ${formatDuration(event.duration)}`}
                />
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Request Trace</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="trace" className="w-full">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="trace">Trace Tree</TabsTrigger>
              <TabsTrigger value="timeline">Timeline</TabsTrigger>
              <TabsTrigger value="sql">SQL Queries</TabsTrigger>
              <TabsTrigger value="http">HTTP Calls</TabsTrigger>
            </TabsList>

            <TabsContent value="trace" className="space-y-4">
              <div className="border rounded-lg">
                {traces.length > 0 ? (
                  traces.map((trace, idx) =>
                    renderTraceNode(trace, 0, idx.toString())
                  )
                ) : (
                  <div className="p-8 text-center text-gray-500">
                    No middleware traces recorded for this request
                  </div>
                )}
              </div>

              {selectedTrace && (
                <Card className="mt-4">
                  <CardHeader>
                    <CardTitle className="text-sm">Trace Details</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="space-y-2 text-sm">
                      <div>
                        <span className="font-medium">Name:</span>{" "}
                        {selectedTrace.name}
                      </div>
                      <div>
                        <span className="font-medium">Type:</span>{" "}
                        <Badge
                          className={cn(
                            "text-xs",
                            getTypeColor(selectedTrace.type)
                          )}
                        >
                          {selectedTrace.type}
                        </Badge>
                      </div>
                      <div>
                        <span className="font-medium">Duration:</span>{" "}
                        {formatDuration(selectedTrace.duration)}
                      </div>
                      <div>
                        <span className="font-medium">Status:</span>{" "}
                        <span className={getStatusColor(selectedTrace.status)}>
                          {selectedTrace.status}
                        </span>
                      </div>
                      {selectedTrace.error && (
                        <div>
                          <span className="font-medium">Error:</span>{" "}
                          <span className="text-red-600">
                            {selectedTrace.error}
                          </span>
                        </div>
                      )}
                      {selectedTrace.details && (
                        <div>
                          <span className="font-medium">Details:</span>
                          <pre className="mt-2 p-2 bg-gray-50 rounded text-xs overflow-x-auto">
                            {JSON.stringify(selectedTrace.details, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            <TabsContent value="timeline">
              <div className="border rounded-lg p-4 overflow-x-auto">
                {renderTimeline()}
              </div>
            </TabsContent>

            <TabsContent value="sql">
              {sqlQueries.length > 0 ? (
                <div>
                  <div className="mb-4 flex items-center justify-between">
                    <span className="text-sm text-gray-600">
                      Total SQL Queries: {sqlQueries.length}
                    </span>
                    <span className="text-sm text-gray-600">
                      Total Time:{" "}
                      {formatDuration(
                        sqlQueries.reduce((acc, q) => acc + q.duration, 0)
                      )}
                    </span>
                  </div>
                  {sqlQueries.map((query, idx) => renderSQLQuery(query, idx))}
                </div>
              ) : (
                <div className="p-8 text-center text-gray-500">
                  No SQL queries recorded for this request
                </div>
              )}
            </TabsContent>

            <TabsContent value="http">
              {httpCalls.length > 0 ? (
                <div>
                  <div className="mb-4 flex items-center justify-between">
                    <span className="text-sm text-gray-600">
                      Total HTTP Calls: {httpCalls.length}
                    </span>
                    <span className="text-sm text-gray-600">
                      Total Time:{" "}
                      {formatDuration(
                        httpCalls.reduce((acc, c) => acc + c.duration, 0)
                      )}
                    </span>
                  </div>
                  {httpCalls.map((call, idx) => renderHTTPCall(call, idx))}
                </div>
              ) : (
                <div className="p-8 text-center text-gray-500">
                  No HTTP calls recorded for this request
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
