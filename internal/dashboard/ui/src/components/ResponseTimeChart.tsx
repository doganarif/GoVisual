import { h } from "preact";
import { useState, useEffect, useMemo } from "preact/hooks";
import { RequestLog } from "../lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";

interface ResponseTimeChartProps {
  requests: RequestLog[];
}

interface ChartData {
  labels: string[];
  datasets: {
    label: string;
    data: number[];
    borderColor: string;
    backgroundColor: string;
  }[];
}

interface EndpointStats {
  endpoint: string;
  count: number;
  avgDuration: number;
  minDuration: number;
  maxDuration: number;
  p95Duration: number;
  p99Duration: number;
}

export function ResponseTimeChart({ requests }: ResponseTimeChartProps) {
  const [timeRange, setTimeRange] = useState<"1h" | "6h" | "24h" | "7d">("24h");
  const [viewType, setViewType] = useState<
    "timeline" | "distribution" | "endpoints"
  >("timeline");

  const filteredRequests = useMemo(() => {
    const now = new Date().getTime();
    const ranges = {
      "1h": 60 * 60 * 1000,
      "6h": 6 * 60 * 60 * 1000,
      "24h": 24 * 60 * 60 * 1000,
      "7d": 7 * 24 * 60 * 60 * 1000,
    };

    const cutoff = now - ranges[timeRange];
    return requests.filter((r) => new Date(r.Timestamp).getTime() > cutoff);
  }, [requests, timeRange]);

  const timelineData = useMemo(() => {
    // Group requests by time buckets
    const bucketSize =
      timeRange === "1h"
        ? 60000 // 1 minute
        : timeRange === "6h"
        ? 300000 // 5 minutes
        : timeRange === "24h"
        ? 900000 // 15 minutes
        : 3600000; // 1 hour for 7d

    const buckets = new Map<number, number[]>();

    filteredRequests.forEach((req) => {
      const time = new Date(req.Timestamp).getTime();
      const bucket = Math.floor(time / bucketSize) * bucketSize;

      if (!buckets.has(bucket)) {
        buckets.set(bucket, []);
      }
      buckets.get(bucket)!.push(req.Duration);
    });

    const sortedBuckets = Array.from(buckets.entries()).sort(
      (a, b) => a[0] - b[0]
    );

    return {
      labels: sortedBuckets.map(([time]) =>
        new Date(time).toLocaleTimeString()
      ),
      avgDuration: sortedBuckets.map(
        ([, durations]) =>
          durations.reduce((a, b) => a + b, 0) / durations.length
      ),
      maxDuration: sortedBuckets.map(([, durations]) => Math.max(...durations)),
      minDuration: sortedBuckets.map(([, durations]) => Math.min(...durations)),
      count: sortedBuckets.map(([, durations]) => durations.length),
    };
  }, [filteredRequests, timeRange]);

  const distributionData = useMemo(() => {
    const bins = [0, 100, 200, 500, 1000, 2000, 5000, 10000, Infinity];
    const labels = [
      "<100ms",
      "100-200ms",
      "200-500ms",
      "500ms-1s",
      "1-2s",
      "2-5s",
      "5-10s",
      ">10s",
    ];
    const counts = new Array(bins.length - 1).fill(0);

    filteredRequests.forEach((req) => {
      for (let i = 0; i < bins.length - 1; i++) {
        if (req.Duration >= bins[i] && req.Duration < bins[i + 1]) {
          counts[i]++;
          break;
        }
      }
    });

    return { labels, counts };
  }, [filteredRequests]);

  const endpointStats = useMemo((): EndpointStats[] => {
    const grouped = new Map<string, number[]>();

    filteredRequests.forEach((req) => {
      const endpoint = `${req.Method} ${req.Path}`;
      if (!grouped.has(endpoint)) {
        grouped.set(endpoint, []);
      }
      grouped.get(endpoint)!.push(req.Duration);
    });

    return Array.from(grouped.entries())
      .map(([endpoint, durations]) => {
        const sorted = [...durations].sort((a, b) => a - b);
        const p95Index = Math.floor(sorted.length * 0.95);
        const p99Index = Math.floor(sorted.length * 0.99);

        return {
          endpoint,
          count: durations.length,
          avgDuration: durations.reduce((a, b) => a + b, 0) / durations.length,
          minDuration: Math.min(...durations),
          maxDuration: Math.max(...durations),
          p95Duration: sorted[p95Index] || sorted[sorted.length - 1],
          p99Duration: sorted[p99Index] || sorted[sorted.length - 1],
        };
      })
      .sort((a, b) => b.avgDuration - a.avgDuration);
  }, [filteredRequests]);

  const renderSimpleChart = (
    data: number[],
    labels: string[],
    height = 200,
    color = "#3b82f6"
  ) => {
    if (data.length === 0) return null;

    const max = Math.max(...data);
    const min = Math.min(...data);
    const range = max - min || 1;

    const width = 800;
    const padding = 40;
    const chartWidth = width - 2 * padding;
    const chartHeight = height - 2 * padding;

    const xStep = chartWidth / (data.length - 1 || 1);
    const yScale = chartHeight / range;

    const points = data.map((value, index) => ({
      x: padding + index * xStep,
      y: padding + (max - value) * yScale,
    }));

    const pathData = points.reduce((path, point, index) => {
      return (
        path +
        (index === 0 ? `M ${point.x},${point.y}` : ` L ${point.x},${point.y}`)
      );
    }, "");

    return (
      <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-full">
        {/* Grid lines */}
        {[0, 1, 2, 3, 4].map((i) => {
          const y = padding + (chartHeight * i) / 4;
          const value = max - (range * i) / 4;
          return (
            <g key={i}>
              <line
                x1={padding}
                y1={y}
                x2={width - padding}
                y2={y}
                stroke="#e5e7eb"
                strokeWidth="1"
              />
              <text
                x={padding - 5}
                y={y + 4}
                textAnchor="end"
                className="text-xs fill-gray-500"
              >
                {value.toFixed(0)}ms
              </text>
            </g>
          );
        })}

        {/* Line chart */}
        <path d={pathData} fill="none" stroke={color} strokeWidth="2" />

        {/* Data points */}
        {points.map((point, index) => (
          <circle
            key={index}
            cx={point.x}
            cy={point.y}
            r="3"
            fill={color}
            className="hover:r-5 transition-all"
          >
            <title>{`${labels[index]}: ${data[index].toFixed(0)}ms`}</title>
          </circle>
        ))}
      </svg>
    );
  };

  const renderBarChart = (
    data: number[],
    labels: string[],
    height = 200,
    color = "#3b82f6"
  ) => {
    if (data.length === 0) return null;

    const max = Math.max(...data);
    const width = 800;
    const padding = 40;
    const chartWidth = width - 2 * padding;
    const chartHeight = height - 2 * padding;

    const barWidth = (chartWidth / data.length) * 0.8;
    const barSpacing = (chartWidth / data.length) * 0.2;
    const yScale = chartHeight / (max || 1);

    return (
      <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-full">
        {/* Y-axis labels */}
        {[0, 1, 2, 3, 4].map((i) => {
          const y = padding + (chartHeight * i) / 4;
          const value = max - (max * i) / 4;
          return (
            <g key={i}>
              <line
                x1={padding}
                y1={y}
                x2={width - padding}
                y2={y}
                stroke="#e5e7eb"
                strokeWidth="1"
              />
              <text
                x={padding - 5}
                y={y + 4}
                textAnchor="end"
                className="text-xs fill-gray-500"
              >
                {value.toFixed(0)}
              </text>
            </g>
          );
        })}

        {/* Bars */}
        {data.map((value, index) => {
          const x = padding + index * (barWidth + barSpacing) + barSpacing / 2;
          const barHeight = value * yScale;
          const y = padding + chartHeight - barHeight;

          return (
            <g key={index}>
              <rect
                x={x}
                y={y}
                width={barWidth}
                height={barHeight}
                fill={color}
                className="hover:opacity-80 transition-opacity"
              >
                <title>{`${labels[index]}: ${value}`}</title>
              </rect>
              <text
                x={x + barWidth / 2}
                y={height - 5}
                textAnchor="middle"
                className="text-xs fill-gray-500"
                transform={`rotate(-45, ${x + barWidth / 2}, ${height - 5})`}
              >
                {labels[index]}
              </text>
            </g>
          );
        })}
      </svg>
    );
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Response Time Analysis</CardTitle>
          <div className="flex gap-2">
            {(["1h", "6h", "24h", "7d"] as const).map((range) => (
              <button
                key={range}
                onClick={() => setTimeRange(range)}
                className={`px-3 py-1 text-sm rounded ${
                  timeRange === range
                    ? "bg-primary text-primary-foreground"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                {range}
              </button>
            ))}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {filteredRequests.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            No requests in the selected time range
          </div>
        ) : (
          <Tabs value={viewType} onValueChange={(v) => setViewType(v as any)}>
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="timeline">Timeline</TabsTrigger>
              <TabsTrigger value="distribution">Distribution</TabsTrigger>
              <TabsTrigger value="endpoints">Endpoints</TabsTrigger>
            </TabsList>

            <TabsContent value="timeline" className="space-y-4">
              <div>
                <h4 className="text-sm font-medium mb-2">
                  Average Response Time
                </h4>
                <div className="h-48">
                  {renderSimpleChart(
                    timelineData.avgDuration,
                    timelineData.labels
                  )}
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-medium mb-2">
                    Max Response Time
                  </h4>
                  <div className="h-32">
                    {renderSimpleChart(
                      timelineData.maxDuration,
                      timelineData.labels,
                      128,
                      "#ef4444"
                    )}
                  </div>
                </div>
                <div>
                  <h4 className="text-sm font-medium mb-2">Request Count</h4>
                  <div className="h-32">
                    {renderSimpleChart(
                      timelineData.count,
                      timelineData.labels,
                      128,
                      "#10b981"
                    )}
                  </div>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="distribution">
              <div>
                <h4 className="text-sm font-medium mb-2">
                  Response Time Distribution
                </h4>
                <div className="h-64">
                  {renderBarChart(
                    distributionData.counts,
                    distributionData.labels,
                    256,
                    "#8b5cf6"
                  )}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="endpoints">
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b">
                      <th className="text-left p-2">Endpoint</th>
                      <th className="text-right p-2">Count</th>
                      <th className="text-right p-2">Avg</th>
                      <th className="text-right p-2">Min</th>
                      <th className="text-right p-2">Max</th>
                      <th className="text-right p-2">P95</th>
                      <th className="text-right p-2">P99</th>
                    </tr>
                  </thead>
                  <tbody>
                    {endpointStats.slice(0, 10).map((stat) => (
                      <tr
                        key={stat.endpoint}
                        className="border-b hover:bg-gray-50"
                      >
                        <td className="p-2 font-mono text-xs">
                          {stat.endpoint}
                        </td>
                        <td className="text-right p-2">{stat.count}</td>
                        <td className="text-right p-2">
                          {stat.avgDuration.toFixed(0)}ms
                        </td>
                        <td className="text-right p-2">
                          {stat.minDuration.toFixed(0)}ms
                        </td>
                        <td className="text-right p-2">
                          {stat.maxDuration.toFixed(0)}ms
                        </td>
                        <td className="text-right p-2">
                          {stat.p95Duration.toFixed(0)}ms
                        </td>
                        <td className="text-right p-2">
                          {stat.p99Duration.toFixed(0)}ms
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </TabsContent>
          </Tabs>
        )}
      </CardContent>
    </Card>
  );
}
