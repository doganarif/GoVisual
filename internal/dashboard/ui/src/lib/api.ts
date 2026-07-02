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
  Logs?: LogEntry[];
  PanicStack?: string;
}

export interface LogEntry {
  time: string;
  level: string;
  message: string;
  attrs?: Record<string, unknown>;
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

// ApiError carries the HTTP status and response body so callers can render
// useful messages instead of a generic "Failed to fetch". A 404 from a gated
// endpoint and a 403 from an SSRF rejection look the same without this.
export class ApiError extends Error {
  status: number;
  body: string;
  constructor(status: number, body: string, message?: string) {
    super(message || body || `request failed (${status})`);
    this.status = status;
    this.body = body;
  }
  get isNotFound() {
    return this.status === 404;
  }
  get isUnauthorized() {
    return this.status === 401 || this.status === 403;
  }
}

async function request<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    const text = await response.text().catch(() => "");
    throw new ApiError(response.status, text);
  }
  // Some endpoints (clear) return no JSON body.
  const contentType = response.headers.get("content-type") || "";
  if (!contentType.includes("application/json")) {
    return undefined as unknown as T;
  }
  return response.json();
}

// LiveEvent describes the two SSE event types the server emits. Callers
// route by `kind`.
export type LiveEvent =
  | { kind: "snapshot"; data: RequestLog[] }
  | { kind: "append"; data: RequestLog[] };

class API {
  // Resolve the API base from wherever the dashboard is mounted so a
  // custom WithDashboardPath keeps working (#31).
  private baseURL = window.location.pathname.replace(/\/$/, "") + "/api";

  getRequests(signal?: AbortSignal): Promise<RequestLog[]> {
    return request<RequestLog[]>(`${this.baseURL}/requests`, { signal });
  }

  async clearRequests(): Promise<void> {
    await request<void>(`${this.baseURL}/clear`, { method: "POST" });
  }

  compareRequests(
    requestIds: string[],
    signal?: AbortSignal
  ): Promise<RequestLog[]> {
    const params = requestIds
      .map((id) => `id=${encodeURIComponent(id)}`)
      .join("&");
    return request<RequestLog[]>(`${this.baseURL}/compare?${params}`, { signal });
  }

  replayRequest(payload: ReplayRequest): Promise<ReplayResponse> {
    return request<ReplayResponse>(`${this.baseURL}/replay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
  }

  getMetrics(
    requestId: string,
    signal?: AbortSignal
  ): Promise<PerformanceMetrics> {
    return request<PerformanceMetrics>(
      `${this.baseURL}/metrics?id=${encodeURIComponent(requestId)}`,
      { signal }
    );
  }

  getFlameGraph(
    requestId: string,
    signal?: AbortSignal
  ): Promise<FlameGraphNode> {
    return request<FlameGraphNode>(
      `${this.baseURL}/flamegraph?id=${encodeURIComponent(requestId)}`,
      { signal }
    );
  }

  getBottlenecks(signal?: AbortSignal): Promise<any[]> {
    return request<any[]>(`${this.baseURL}/bottlenecks`, { signal });
  }

  getSystemInfo(signal?: AbortSignal): Promise<SystemInfo> {
    return request<SystemInfo>(`${this.baseURL}/system-info`, { signal });
  }

  // subscribeToEvents wires both named SSE events the server emits:
  //   "snapshot": full state, replaces the client list (initial connect and
  //                resync after the store is cleared)
  //   "append":   one or more new requests, prepended to the client list
  // A default onmessage handler would receive neither — named events go
  // exclusively to addEventListener.
  subscribeToEvents(
    onEvent: (event: LiveEvent) => void,
    onError?: (err: Event) => void
  ): EventSource {
    const eventSource = new EventSource(`${this.baseURL}/events`);

    const handle = (kind: LiveEvent["kind"]) => (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        onEvent({ kind, data });
      } catch (err) {
        console.error(`Failed to parse ${kind} event:`, err);
      }
    };

    eventSource.addEventListener(
      "snapshot",
      handle("snapshot") as EventListener
    );
    eventSource.addEventListener("append", handle("append") as EventListener);

    if (onError) {
      eventSource.onerror = onError;
    }
    return eventSource;
  }

  exportRequests(requests: RequestLog[]): string {
    return JSON.stringify(requests, null, 2);
  }

  importRequests(jsonString: string): RequestLog[] {
    const data = JSON.parse(jsonString);
    if (!Array.isArray(data)) {
      throw new Error("Invalid format: expected an array of requests");
    }
    return data as RequestLog[];
  }
}

export const api = new API();
