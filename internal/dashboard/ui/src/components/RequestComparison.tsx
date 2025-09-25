import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { api, RequestLog } from "../lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { cn } from "../lib/utils";

interface RequestComparisonProps {
  requestIds: string[];
  allRequests: RequestLog[];
  onClose: () => void;
}

export function RequestComparison({
  requestIds,
  allRequests,
  onClose,
}: RequestComparisonProps) {
  const [compareRequests, setCompareRequests] = useState<RequestLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadComparisonData();
  }, [requestIds]);

  const loadComparisonData = async () => {
    try {
      setLoading(true);
      const data = await api.compareRequests(requestIds);
      setCompareRequests(data);
    } catch (error) {
      console.error("Failed to load comparison data:", error);
      // Fallback to local data
      const localData = allRequests.filter((req) =>
        requestIds.includes(req.ID)
      );
      setCompareRequests(localData);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-muted-foreground">Loading comparison...</div>
      </div>
    );
  }

  if (compareRequests.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <div className="text-muted-foreground mb-4">
          No requests found for comparison
        </div>
        <Button onClick={onClose}>Close</Button>
      </div>
    );
  }

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  const formatDate = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return "text-green-600";
    if (status >= 300 && status < 400) return "text-blue-600";
    if (status >= 400 && status < 500) return "text-yellow-600";
    if (status >= 500) return "text-red-600";
    return "text-gray-600";
  };

  const renderComparisonRow = (label: string, values: any[]) => {
    const allSame = values.every(
      (v) => JSON.stringify(v) === JSON.stringify(values[0])
    );

    return (
      <tr>
        <td className="font-medium text-sm p-2 border-b">{label}</td>
        {values.map((value, idx) => (
          <td
            key={idx}
            className={cn(
              "text-sm p-2 border-b",
              !allSame && "bg-yellow-50 dark:bg-yellow-900/10"
            )}
          >
            {typeof value === "object" ? JSON.stringify(value, null, 2) : value}
          </td>
        ))}
      </tr>
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-2xl font-bold">Request Comparison</h2>
        <Button onClick={onClose} variant="outline">
          Close
        </Button>
      </div>

      <Tabs defaultValue="overview" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="headers">Headers</TabsTrigger>
          <TabsTrigger value="body">Body</TabsTrigger>
          <TabsTrigger value="performance">Performance</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Request Details</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr>
                      <th className="text-left p-2 border-b">Property</th>
                      {compareRequests.map((req, idx) => (
                        <th key={idx} className="text-left p-2 border-b">
                          Request {idx + 1}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {renderComparisonRow(
                      "Method",
                      compareRequests.map((r) => r.Method)
                    )}
                    {renderComparisonRow(
                      "Path",
                      compareRequests.map((r) => r.Path)
                    )}
                    {renderComparisonRow(
                      "Query",
                      compareRequests.map((r) => r.Query || "None")
                    )}
                    {renderComparisonRow(
                      "Status",
                      compareRequests.map((r) => (
                        <span className={getStatusColor(r.StatusCode)}>
                          {r.StatusCode}
                        </span>
                      ))
                    )}
                    {renderComparisonRow(
                      "Duration",
                      compareRequests.map((r) => formatDuration(r.Duration))
                    )}
                    {renderComparisonRow(
                      "Timestamp",
                      compareRequests.map((r) => formatDate(r.Timestamp))
                    )}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="headers" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle>Request Headers</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {compareRequests.map((req, idx) => (
                    <div key={idx}>
                      <h4 className="font-medium mb-2">Request {idx + 1}</h4>
                      <div className="bg-gray-50 dark:bg-gray-900 rounded p-2 text-xs font-mono">
                        {Object.entries(req.RequestHeaders || {}).map(
                          ([key, values]) => (
                            <div key={key}>
                              <span className="text-blue-600">{key}:</span>{" "}
                              {Array.isArray(values)
                                ? values.join(", ")
                                : values}
                            </div>
                          )
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Response Headers</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {compareRequests.map((req, idx) => (
                    <div key={idx}>
                      <h4 className="font-medium mb-2">Request {idx + 1}</h4>
                      <div className="bg-gray-50 dark:bg-gray-900 rounded p-2 text-xs font-mono">
                        {Object.entries(req.ResponseHeaders || {}).map(
                          ([key, values]) => (
                            <div key={key}>
                              <span className="text-green-600">{key}:</span>{" "}
                              {Array.isArray(values)
                                ? values.join(", ")
                                : values}
                            </div>
                          )
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="body" className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <Card>
              <CardHeader>
                <CardTitle>Request Body</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {compareRequests.map((req, idx) => (
                    <div key={idx}>
                      <h4 className="font-medium mb-2">Request {idx + 1}</h4>
                      <div className="bg-gray-50 dark:bg-gray-900 rounded p-2">
                        <pre className="text-xs overflow-x-auto">
                          {req.RequestBody || "No request body"}
                        </pre>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Response Body</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {compareRequests.map((req, idx) => (
                    <div key={idx}>
                      <h4 className="font-medium mb-2">Request {idx + 1}</h4>
                      <div className="bg-gray-50 dark:bg-gray-900 rounded p-2">
                        <pre className="text-xs overflow-x-auto max-h-48 overflow-y-auto">
                          {req.ResponseBody || "No response body"}
                        </pre>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="performance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Performance Metrics</CardTitle>
            </CardHeader>
            <CardContent>
              {compareRequests.some((r) => r.PerformanceMetrics) ? (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr>
                        <th className="text-left p-2 border-b">Metric</th>
                        {compareRequests.map((req, idx) => (
                          <th key={idx} className="text-left p-2 border-b">
                            Request {idx + 1}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {renderComparisonRow(
                        "CPU Time",
                        compareRequests.map((r) =>
                          r.PerformanceMetrics
                            ? `${r.PerformanceMetrics.cpu_time}ms`
                            : "N/A"
                        )
                      )}
                      {renderComparisonRow(
                        "Memory Allocated",
                        compareRequests.map((r) =>
                          r.PerformanceMetrics
                            ? `${(
                                r.PerformanceMetrics.memory_alloc /
                                1024 /
                                1024
                              ).toFixed(2)}MB`
                            : "N/A"
                        )
                      )}
                      {renderComparisonRow(
                        "Goroutines",
                        compareRequests.map(
                          (r) => r.PerformanceMetrics?.num_goroutines || "N/A"
                        )
                      )}
                      {renderComparisonRow(
                        "GC Runs",
                        compareRequests.map(
                          (r) => r.PerformanceMetrics?.num_gc || "N/A"
                        )
                      )}
                      {renderComparisonRow(
                        "GC Pause",
                        compareRequests.map((r) =>
                          r.PerformanceMetrics
                            ? `${r.PerformanceMetrics.gc_pause_total}ms`
                            : "N/A"
                        )
                      )}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center text-muted-foreground py-8">
                  No performance metrics available for these requests
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
