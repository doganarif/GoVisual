import { h } from "preact";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { RequestLog } from "../lib/api";

interface RequestDetailsProps {
  request: RequestLog | null;
  onShowPerformance?: () => void;
}

export function RequestDetails({
  request,
  onShowPerformance,
}: RequestDetailsProps) {
  if (!request) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-64 text-muted-foreground">
          Select a request to view details
        </CardContent>
      </Card>
    );
  }

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

  const hasPerformanceMetrics = !!request.PerformanceMetrics;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Request Details</CardTitle>
            <CardDescription className="mt-1">
              {request.Method} {request.Path} •{" "}
              {new Date(request.Timestamp).toLocaleString()}
            </CardDescription>
          </div>
          <div className="flex gap-2">
            {hasPerformanceMetrics && (
              <Button size="sm" variant="outline" onClick={onShowPerformance}>
                View Performance
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="text-sm font-medium text-muted-foreground">
              ID
            </label>
            <p className="font-mono text-sm">{request.ID}</p>
          </div>
          <div>
            <label className="text-sm font-medium text-muted-foreground">
              Status
            </label>
            <div className="flex items-center gap-2">
              <Badge
                variant={
                  request.StatusCode >= 200 && request.StatusCode < 300
                    ? "default"
                    : request.StatusCode >= 300 && request.StatusCode < 400
                    ? "secondary"
                    : "outline"
                }
              >
                {request.StatusCode}
              </Badge>
              <span className="text-sm text-muted-foreground">
                {request.Duration}ms
              </span>
            </div>
          </div>
        </div>

        <Tabs defaultValue="headers" className="w-full">
          <TabsList className="w-full justify-start">
            <TabsTrigger value="headers">Headers</TabsTrigger>
            <TabsTrigger value="request">Request</TabsTrigger>
            <TabsTrigger value="response">Response</TabsTrigger>
            {request.Error && <TabsTrigger value="error">Error</TabsTrigger>}
          </TabsList>

          <TabsContent value="headers" className="mt-4">
            <div className="space-y-4">
              <div>
                <h4 className="text-sm font-medium mb-2">Request Headers</h4>
                <pre className="bg-muted p-3 rounded-md text-xs overflow-x-auto">
                  {formatHeaders(request.RequestHeaders)}
                </pre>
              </div>
              <div>
                <h4 className="text-sm font-medium mb-2">Response Headers</h4>
                <pre className="bg-muted p-3 rounded-md text-xs overflow-x-auto">
                  {formatHeaders(request.ResponseHeaders)}
                </pre>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="request" className="mt-4">
            <pre className="bg-muted p-3 rounded-md text-xs overflow-x-auto max-h-96">
              {formatBody(request.RequestBody)}
            </pre>
          </TabsContent>

          <TabsContent value="response" className="mt-4">
            <pre className="bg-muted p-3 rounded-md text-xs overflow-x-auto max-h-96">
              {formatBody(request.ResponseBody)}
            </pre>
          </TabsContent>

          {request.Error && (
            <TabsContent value="error" className="mt-4">
              <div className="bg-destructive/10 border border-destructive/20 p-3 rounded-md">
                <p className="text-sm text-destructive">{request.Error}</p>
              </div>
            </TabsContent>
          )}
        </Tabs>
      </CardContent>
    </Card>
  );
}
