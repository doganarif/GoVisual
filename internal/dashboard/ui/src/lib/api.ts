export interface RequestLog {
  ID: string;
  Timestamp: string;
  Method: string;
  Path: string;
  Query: string;
  RequestHeaders: Record<string, string[]>;
  ResponseHeaders: Record<string, string[]>;
  StatusCode: number;
  Duration: number;
  RequestBody?: string;
  ResponseBody?: string;
  Error?: string;
  MiddlewareTrace?: any[];
  RouteTrace?: any;
  PerformanceMetrics?: PerformanceMetrics;
}

export interface PerformanceMetrics {
  request_id: string;
  start_time: string;
  end_time: string;
  duration: number;
  cpu_time: number;
  memory_alloc: number;
  memory_total_alloc: number;
  num_goroutines: number;
  num_gc: number;
  gc_pause_total: number;
  function_timings?: Record<string, number>;
  sql_queries?: SQLQuery[];
  http_calls?: HTTPCall[];
  bottlenecks?: Bottleneck[];
}

export interface SQLQuery {
  query: string;
  duration: number;
  rows: number;
  error?: string;
}

export interface HTTPCall {
  method: string;
  url: string;
  duration: number;
  status: number;
  size: number;
}

export interface Bottleneck {
  type: string;
  description: string;
  impact: number;
  duration: number;
  suggestion: string;
}

export interface FlameGraphNode {
  name: string;
  value: number;
  percentage?: string;
  children?: FlameGraphNode[];
}

export interface SystemInfo {
  goVersion: string;
  goos: string;
  goarch: string;
  hostname: string;
  cpuCores: number;
  memoryUsed: number;
  memoryTotal: number;
  envVars: Record<string, string>;
}

export interface ReplayRequest {
  requestId: string;
  url: string;
  method: string;
  headers: Record<string, string>;
  body: string;
}

export interface ReplayResponse {
  statusCode: number;
  headers: Record<string, string[]>;
  body: string;
  duration: number;
  originalRequest: string;
}

class API {
  private baseURL = "/__viz/api";

  async getRequests(): Promise<RequestLog[]> {
    const response = await fetch(`${this.baseURL}/requests`);
    if (!response.ok) throw new Error("Failed to fetch requests");
    return response.json();
  }

  async clearRequests(): Promise<void> {
    const response = await fetch(`${this.baseURL}/clear`, {
      method: "POST",
    });
    if (!response.ok) throw new Error("Failed to clear requests");
  }

  async compareRequests(requestIds: string[]): Promise<RequestLog[]> {
    const params = requestIds.map((id) => `id=${id}`).join("&");
    const response = await fetch(`${this.baseURL}/compare?${params}`);
    if (!response.ok) throw new Error("Failed to compare requests");
    return response.json();
  }

  async replayRequest(request: ReplayRequest): Promise<ReplayResponse> {
    const response = await fetch(`${this.baseURL}/replay`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!response.ok) throw new Error("Failed to replay request");
    return response.json();
  }

  async getMetrics(requestId: string): Promise<PerformanceMetrics> {
    const response = await fetch(`${this.baseURL}/metrics?id=${requestId}`);
    if (!response.ok) throw new Error("Failed to fetch metrics");
    return response.json();
  }

  async getFlameGraph(requestId: string): Promise<FlameGraphNode> {
    const response = await fetch(`${this.baseURL}/flamegraph?id=${requestId}`);
    if (!response.ok) throw new Error("Failed to fetch flame graph");
    return response.json();
  }

  async getBottlenecks(): Promise<any[]> {
    const response = await fetch(`${this.baseURL}/bottlenecks`);
    if (!response.ok) throw new Error("Failed to fetch bottlenecks");
    return response.json();
  }

  async getSystemInfo(): Promise<SystemInfo> {
    const response = await fetch(`${this.baseURL}/system-info`);
    if (!response.ok) throw new Error("Failed to fetch system info");
    return response.json();
  }

  subscribeToEvents(onMessage: (data: RequestLog[]) => void): EventSource {
    const eventSource = new EventSource(`${this.baseURL}/events`);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (error) {
        console.error("Failed to parse event data:", error);
      }
    };

    eventSource.onerror = (error) => {
      console.error("EventSource error:", error);
    };

    return eventSource;
  }

  // Export requests as JSON
  exportRequests(requests: RequestLog[]): string {
    return JSON.stringify(requests, null, 2);
  }

  // Import requests from JSON
  importRequests(jsonString: string): RequestLog[] {
    try {
      const data = JSON.parse(jsonString);
      if (Array.isArray(data)) {
        return data as RequestLog[];
      }
      throw new Error("Invalid format: expected an array of requests");
    } catch (error) {
      throw new Error(`Failed to import requests: ${error.message}`);
    }
  }
}

export const api = new API();
