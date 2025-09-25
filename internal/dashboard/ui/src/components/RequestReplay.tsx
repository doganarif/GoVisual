import { h } from "preact";
import { useState } from "preact/hooks";
import { api, RequestLog, ReplayResponse } from "../lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Input } from "./ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { cn } from "../lib/utils";

interface RequestReplayProps {
  request: RequestLog;
  onClose: () => void;
}

export function RequestReplay({ request, onClose }: RequestReplayProps) {
  const [replayUrl, setReplayUrl] = useState(
    request.Path + (request.Query ? `?${request.Query}` : "")
  );
  const [replayMethod, setReplayMethod] = useState(request.Method);
  const [replayHeaders, setReplayHeaders] = useState<Record<string, string>>(
    () => {
      const headers: Record<string, string> = {};
      if (request.RequestHeaders) {
        Object.entries(request.RequestHeaders).forEach(([key, values]) => {
          headers[key] = Array.isArray(values) ? values[0] : values;
        });
      }
      return headers;
    }
  );
  const [replayBody, setReplayBody] = useState(request.RequestBody || "");
  const [isReplaying, setIsReplaying] = useState(false);
  const [replayResponse, setReplayResponse] = useState<ReplayResponse | null>(
    null
  );
  const [replayError, setReplayError] = useState<string | null>(null);

  const handleReplay = async () => {
    try {
      setIsReplaying(true);
      setReplayError(null);

      // Build full URL if needed
      let fullUrl = replayUrl;
      if (!fullUrl.startsWith("http")) {
        // Try to extract host from original request headers
        const hostHeader = request.RequestHeaders?.["Host"];
        const host = hostHeader
          ? Array.isArray(hostHeader)
            ? hostHeader[0]
            : hostHeader
          : "localhost";
        const protocol = "http://"; // Default to http, could be made configurable
        fullUrl = protocol + host + fullUrl;
      }

      const response = await api.replayRequest({
        requestId: request.ID,
        url: fullUrl,
        method: replayMethod,
        headers: replayHeaders,
        body: replayBody,
      });

      setReplayResponse(response);
    } catch (error) {
      setReplayError(error.message || "Failed to replay request");
    } finally {
      setIsReplaying(false);
    }
  };

  const handleHeaderChange = (key: string, value: string) => {
    setReplayHeaders((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const addHeader = () => {
    const newKey = prompt("Enter header name:");
    if (newKey) {
      setReplayHeaders((prev) => ({
        ...prev,
        [newKey]: "",
      }));
    }
  };

  const removeHeader = (key: string) => {
    setReplayHeaders((prev) => {
      const updated = { ...prev };
      delete updated[key];
      return updated;
    });
  };

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return "bg-green-100 text-green-800";
    if (status >= 300 && status < 400) return "bg-blue-100 text-blue-800";
    if (status >= 400 && status < 500) return "bg-yellow-100 text-yellow-800";
    if (status >= 500) return "bg-red-100 text-red-800";
    return "bg-gray-100 text-gray-800";
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-2xl font-bold">Replay Request</h2>
        <Button onClick={onClose} variant="outline">
          Close
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Request Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Method</label>
              <select
                value={replayMethod}
                onChange={(e) =>
                  setReplayMethod((e.target as HTMLSelectElement).value)
                }
                className="w-full p-2 border rounded-md"
              >
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="PATCH">PATCH</option>
                <option value="DELETE">DELETE</option>
                <option value="HEAD">HEAD</option>
                <option value="OPTIONS">OPTIONS</option>
              </select>
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">URL</label>
              <Input
                value={replayUrl}
                onChange={(e) =>
                  setReplayUrl((e.target as HTMLInputElement).value)
                }
                placeholder="Enter URL"
              />
            </div>
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium">Headers</label>
              <Button size="sm" variant="outline" onClick={addHeader}>
                Add Header
              </Button>
            </div>
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {Object.entries(replayHeaders).map(([key, value]) => (
                <div key={key} className="flex items-center gap-2">
                  <Input
                    value={key}
                    disabled
                    className="flex-1 font-mono text-sm"
                  />
                  <Input
                    value={value}
                    onChange={(e) =>
                      handleHeaderChange(
                        key,
                        (e.target as HTMLInputElement).value
                      )
                    }
                    placeholder="Value"
                    className="flex-2 font-mono text-sm"
                  />
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => removeHeader(key)}
                    className="text-red-600 hover:text-red-700"
                  >
                    Remove
                  </Button>
                </div>
              ))}
            </div>
          </div>

          {(replayMethod === "POST" ||
            replayMethod === "PUT" ||
            replayMethod === "PATCH") && (
            <div>
              <label className="text-sm font-medium mb-1 block">
                Request Body
              </label>
              <textarea
                value={replayBody}
                onChange={(e) =>
                  setReplayBody((e.target as HTMLTextAreaElement).value)
                }
                className="w-full p-2 border rounded-md font-mono text-sm"
                rows={6}
                placeholder="Enter request body (JSON, XML, etc.)"
              />
            </div>
          )}

          <div className="flex justify-end gap-2">
            <Button onClick={onClose} variant="outline">
              Cancel
            </Button>
            <Button
              onClick={handleReplay}
              disabled={isReplaying}
              className={cn(isReplaying && "opacity-50 cursor-not-allowed")}
            >
              {isReplaying ? "Replaying..." : "Send Request"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {replayError && (
        <Card className="border-red-200 bg-red-50">
          <CardHeader>
            <CardTitle className="text-red-800">Error</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-red-700">{replayError}</p>
          </CardContent>
        </Card>
      )}

      {replayResponse && (
        <Card>
          <CardHeader>
            <CardTitle>Response</CardTitle>
          </CardHeader>
          <CardContent>
            <Tabs defaultValue="overview" className="w-full">
              <TabsList>
                <TabsTrigger value="overview">Overview</TabsTrigger>
                <TabsTrigger value="headers">Headers</TabsTrigger>
                <TabsTrigger value="body">Body</TabsTrigger>
              </TabsList>

              <TabsContent value="overview" className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Status
                    </span>
                    <div className="mt-1">
                      <Badge
                        className={getStatusColor(replayResponse.statusCode)}
                      >
                        {replayResponse.statusCode}
                      </Badge>
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Duration
                    </span>
                    <div className="mt-1 text-lg font-medium">
                      {replayResponse.duration}ms
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Original Request
                    </span>
                    <div className="mt-1 text-sm font-mono">
                      {replayResponse.originalRequest}
                    </div>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="headers">
                <div className="bg-gray-50 dark:bg-gray-900 rounded p-4">
                  <div className="space-y-1 text-sm font-mono">
                    {Object.entries(replayResponse.headers).map(
                      ([key, values]) => (
                        <div key={key}>
                          <span className="text-blue-600">{key}:</span>{" "}
                          <span className="text-gray-700 dark:text-gray-300">
                            {Array.isArray(values) ? values.join(", ") : values}
                          </span>
                        </div>
                      )
                    )}
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="body">
                <div className="bg-gray-50 dark:bg-gray-900 rounded p-4">
                  <pre className="text-sm font-mono overflow-x-auto max-h-96 overflow-y-auto">
                    {replayResponse.body}
                  </pre>
                </div>
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
