import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { api, ApiError, SystemInfo } from "../lib/api";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "./ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./ui/table";

type State =
  | { kind: "loading" }
  | { kind: "ready"; info: SystemInfo }
  | { kind: "disabled" } // endpoint gated off by server config
  | { kind: "error"; message: string };

export function EnvironmentInfo() {
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    const controller = new AbortController();
    api
      .getSystemInfo(controller.signal)
      .then((info) => setState({ kind: "ready", info }))
      .catch((err) => {
        if (err?.name === "AbortError") return;
        if (err instanceof ApiError && err.isNotFound) {
          setState({ kind: "disabled" });
          return;
        }
        setState({
          kind: "error",
          message: err instanceof Error ? err.message : "Failed to load",
        });
      });
    return () => controller.abort();
  }, []);

  if (state.kind === "loading") {
    return (
      <div className="text-sm text-muted-foreground">
        Loading system information...
      </div>
    );
  }

  if (state.kind === "disabled") {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System info is disabled</CardTitle>
          <CardDescription>
            The <code>/__viz/api/system-info</code> endpoint is off by default.
            Enable it on the server with{" "}
            <code>govisual.WithSystemInfo(...)</code>, passing the env var
            allowlist you want exposed.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (state.kind === "error") {
    return (
      <Card className="border-destructive/50 bg-destructive/5">
        <CardHeader>
          <CardTitle>Failed to load system info</CardTitle>
          <CardDescription className="text-destructive">
            {state.message}
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  const info = state.info;
  const memoryPercentage =
    info.memoryTotal > 0 ? (info.memoryUsed / info.memoryTotal) * 100 : 0;
  const envEntries = Object.entries(info.envVars);

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Go Environment</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <Row label="Version" value={info.goVersion} />
            <Row label="GOOS" value={info.goos} />
            <Row label="GOARCH" value={info.goarch} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <Row label="Hostname" value={info.hostname} />
            <Row label="OS" value={info.goos} />
            <Row label="CPU Cores" value={String(info.cpuCores)} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Memory Usage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-primary h-2 rounded-full transition-all duration-300"
                  style={{ width: `${Math.min(100, memoryPercentage)}%` }}
                />
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">
                  {info.memoryUsed}MB / {info.memoryTotal}MB
                </span>
                <span className="font-medium">
                  {memoryPercentage.toFixed(1)}%
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Environment Variables</CardTitle>
          <CardDescription>
            Only variables explicitly allowlisted on the server are shown.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {envEntries.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No environment variables are exposed. Pass names to
              <code className="mx-1">WithSystemInfo(...)</code>
              on the server to surface them here.
            </p>
          ) : (
            <div className="max-h-96 overflow-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Value</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {envEntries.map(([key, value]) => (
                    <TableRow key={key}>
                      <TableCell className="font-mono text-sm">{key}</TableCell>
                      <TableCell className="font-mono text-sm text-muted-foreground break-all">
                        {value}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-sm text-muted-foreground">{label}:</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}
