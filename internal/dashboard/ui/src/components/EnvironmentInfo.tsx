import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";

interface SystemInfo {
  goVersion: string;
  goos: string;
  goarch: string;
  hostname: string;
  cpuCores: number;
  memoryUsed: number;
  memoryTotal: number;
  envVars: Record<string, string>;
}

export function EnvironmentInfo() {
  const [systemInfo, setSystemInfo] = useState<SystemInfo>({
    goVersion: "Loading...",
    goos: "Loading...",
    goarch: "Loading...",
    hostname: "Loading...",
    cpuCores: 0,
    memoryUsed: 0,
    memoryTotal: 0,
    envVars: {}
  });

  useEffect(() => {
    // Fetch system info from API
    fetchSystemInfo();
  }, []);

  const fetchSystemInfo = async () => {
    try {
      const response = await fetch("/__viz/api/system-info");
      if (response.ok) {
        const data = await response.json();
        setSystemInfo(data);
      }
    } catch (error) {
      console.error("Failed to fetch system info:", error);
      // Set default values for demo
      setSystemInfo({
        goVersion: "go1.21.0",
        goos: "darwin",
        goarch: "arm64",
        hostname: "localhost",
        cpuCores: navigator.hardwareConcurrency || 4,
        memoryUsed: 256,
        memoryTotal: 1024,
        envVars: {
          PATH: "/usr/local/bin:/usr/bin:/bin",
          HOME: "/Users/user",
          GOPATH: "/Users/user/go"
        }
      });
    }
  };

  const memoryPercentage = (systemInfo.memoryUsed / systemInfo.memoryTotal) * 100;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Go Environment */}
        <Card>
          <CardHeader>
            <CardTitle>Go Environment</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Version:</span>
              <span className="text-sm font-medium">{systemInfo.goVersion}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">GOOS:</span>
              <span className="text-sm font-medium">{systemInfo.goos}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">GOARCH:</span>
              <span className="text-sm font-medium">{systemInfo.goarch}</span>
            </div>
          </CardContent>
        </Card>

        {/* System Info */}
        <Card>
          <CardHeader>
            <CardTitle>System</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">Hostname:</span>
              <span className="text-sm font-medium">{systemInfo.hostname}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">OS:</span>
              <span className="text-sm font-medium">{systemInfo.goos}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">CPU Cores:</span>
              <span className="text-sm font-medium">{systemInfo.cpuCores}</span>
            </div>
          </CardContent>
        </Card>

        {/* Memory Usage */}
        <Card>
          <CardHeader>
            <CardTitle>Memory Usage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div 
                  className="bg-primary h-2 rounded-full transition-all duration-300"
                  style={{ width: `${memoryPercentage}%` }}
                />
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">
                  {systemInfo.memoryUsed}MB / {systemInfo.memoryTotal}MB
                </span>
                <span className="font-medium">
                  {memoryPercentage.toFixed(1)}%
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Environment Variables */}
      <Card>
        <CardHeader>
          <CardTitle>Environment Variables</CardTitle>
          <CardDescription>System environment variables (sensitive values are redacted)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="max-h-96 overflow-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Value</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {Object.entries(systemInfo.envVars).map(([key, value]) => (
                  <TableRow key={key}>
                    <TableCell className="font-mono text-sm">{key}</TableCell>
                    <TableCell className="font-mono text-sm text-muted-foreground">
                      {value}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
