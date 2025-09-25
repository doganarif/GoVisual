import { h } from "preact";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./ui/table";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { RequestLog } from "../lib/api";
import { cn } from "../lib/utils";

interface RequestTableProps {
  requests: RequestLog[];
  selectedRequest?: RequestLog | null;
  onRequestSelect: (request: RequestLog) => void;
  selectedForComparison?: string[];
  onToggleComparison?: (requestId: string) => void;
  onReplay?: (request: RequestLog) => void;
}

export function RequestTable({
  requests,
  selectedRequest,
  onRequestSelect,
  selectedForComparison = [],
  onToggleComparison,
  onReplay,
}: RequestTableProps) {
  const getStatusVariant = (
    status: number
  ): "default" | "secondary" | "outline" => {
    if (status >= 200 && status < 300) return "default";
    if (status >= 300 && status < 400) return "secondary";
    if (status >= 400) return "outline";
    return "secondary";
  };

  const getMethodClass = (method: string): string => {
    return `method-${method.toLowerCase()}`;
  };

  const formatDuration = (duration: number): string => {
    if (duration < 1) return "<1ms";
    if (duration < 1000) return `${duration}ms`;
    return `${(duration / 1000).toFixed(2)}s`;
  };

  const formatTime = (timestamp: string): string => {
    return new Date(timestamp).toLocaleTimeString();
  };

  const getDurationClass = (duration: number): string => {
    if (duration > 500) return "font-semibold";
    if (duration > 200) return "font-medium";
    return "";
  };

  if (!requests || requests.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
        <svg
          className="w-16 h-16 mb-4 opacity-30"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <p className="text-sm font-medium">No requests logged yet</p>
        <p className="text-xs text-muted-foreground/70 mt-1">
          Waiting for incoming HTTP requests...
        </p>
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {onToggleComparison && (
            <TableHead className="w-12">
              <span className="sr-only">Select</span>
            </TableHead>
          )}
          <TableHead className="w-24">Time</TableHead>
          <TableHead className="w-20">Method</TableHead>
          <TableHead>Path</TableHead>
          <TableHead className="w-20">Status</TableHead>
          <TableHead className="w-24 text-right">Duration</TableHead>
          {onReplay && (
            <TableHead className="w-20">
              <span className="sr-only">Actions</span>
            </TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {requests.map((request) => (
          <TableRow
            key={request.ID}
            onClick={(e) => {
              // Don't select if clicking on checkbox or button
              const target = e.target as HTMLElement;
              if (
                target.tagName !== "INPUT" &&
                target.tagName !== "BUTTON" &&
                !target.closest("button")
              ) {
                onRequestSelect(request);
              }
            }}
            className={cn(
              "cursor-pointer transition-all duration-150 hover:bg-muted/50",
              selectedRequest?.ID === request.ID && "bg-accent/50 shadow-sm"
            )}
            style={{ animationDelay: `${requests.indexOf(request) * 20}ms` }}
          >
            {onToggleComparison && (
              <TableCell>
                <input
                  type="checkbox"
                  checked={selectedForComparison.includes(request.ID)}
                  onChange={(e) => {
                    e.stopPropagation();
                    onToggleComparison(request.ID);
                  }}
                  className="h-4 w-4 rounded border-gray-300"
                  onClick={(e) => e.stopPropagation()}
                />
              </TableCell>
            )}
            <TableCell className="text-xs text-muted-foreground font-medium">
              {formatTime(request.Timestamp)}
            </TableCell>
            <TableCell>
              <span
                className={cn(
                  "px-2 py-1 rounded text-xs",
                  getMethodClass(request.Method)
                )}
              >
                {request.Method}
              </span>
            </TableCell>
            <TableCell className="font-mono text-sm">
              <span className="font-medium">{request.Path}</span>
              {request.Query && (
                <span className="text-muted-foreground text-xs">
                  ?{request.Query}
                </span>
              )}
            </TableCell>
            <TableCell>
              <Badge
                variant={getStatusVariant(request.StatusCode)}
                className="font-mono"
              >
                {request.StatusCode}
              </Badge>
            </TableCell>
            <TableCell
              className={cn(
                "text-right tabular-nums font-mono text-sm",
                getDurationClass(request.Duration)
              )}
            >
              {formatDuration(request.Duration)}
            </TableCell>
            {onReplay && (
              <TableCell>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={(e) => {
                    e.stopPropagation();
                    onReplay(request);
                  }}
                  className="h-7 px-2 text-xs"
                >
                  Replay
                </Button>
              </TableCell>
            )}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
