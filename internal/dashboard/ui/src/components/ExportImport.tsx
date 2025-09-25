import { h } from "preact";
import { useState, useRef } from "preact/hooks";
import { api, RequestLog } from "../lib/api";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";

interface ExportImportProps {
  requests: RequestLog[];
  onImport: (requests: RequestLog[]) => void;
}

export function ExportImport({ requests, onImport }: ExportImportProps) {
  const [importError, setImportError] = useState<string | null>(null);
  const [importSuccess, setImportSuccess] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleExport = () => {
    try {
      const jsonData = api.exportRequests(requests);
      const blob = new Blob([jsonData], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `govisual-requests-${Date.now()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Export failed:", error);
    }
  };

  const handleExportCSV = () => {
    try {
      // Convert to CSV format
      const headers = [
        "ID",
        "Timestamp",
        "Method",
        "Path",
        "Status",
        "Duration (ms)",
        "Error",
      ];
      const rows = requests.map((req) => [
        req.ID,
        req.Timestamp,
        req.Method,
        req.Path,
        req.StatusCode,
        req.Duration,
        req.Error || "",
      ]);

      const csvContent = [
        headers.join(","),
        ...rows.map((row) => row.map((cell) => `"${cell}"`).join(",")),
      ].join("\n");

      const blob = new Blob([csvContent], { type: "text/csv" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `govisual-requests-${Date.now()}.csv`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("CSV export failed:", error);
    }
  };

  const handleImport = (event: Event) => {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];

    if (!file) return;

    setImportError(null);
    setImportSuccess(false);

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const content = e.target?.result as string;
        const importedRequests = api.importRequests(content);
        onImport(importedRequests);
        setImportSuccess(true);
        setTimeout(() => setImportSuccess(false), 3000);
      } catch (error) {
        setImportError(error.message || "Failed to import requests");
        setTimeout(() => setImportError(null), 5000);
      }
    };

    reader.readAsText(file);

    // Reset file input
    target.value = "";
  };

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Export & Import</CardTitle>
        <CardDescription>
          Export request logs for analysis or import previously saved logs
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <h4 className="text-sm font-medium mb-2">Export Data</h4>
            <div className="flex flex-col sm:flex-row gap-2">
              <Button
                onClick={handleExport}
                disabled={requests.length === 0}
                className="flex-1"
              >
                Export as JSON
              </Button>
              <Button
                onClick={handleExportCSV}
                disabled={requests.length === 0}
                variant="outline"
                className="flex-1"
              >
                Export as CSV
              </Button>
            </div>
            {requests.length === 0 && (
              <p className="text-xs text-muted-foreground mt-2">
                No requests to export
              </p>
            )}
            {requests.length > 0 && (
              <p className="text-xs text-muted-foreground mt-2">
                {requests.length} request{requests.length !== 1 ? "s" : ""} will
                be exported
              </p>
            )}
          </div>

          <div className="flex-1">
            <h4 className="text-sm font-medium mb-2">Import Data</h4>
            <input
              ref={fileInputRef}
              type="file"
              accept=".json"
              onChange={handleImport}
              className="hidden"
            />
            <Button
              onClick={handleImportClick}
              variant="outline"
              className="w-full"
            >
              Import JSON
            </Button>
            <p className="text-xs text-muted-foreground mt-2">
              Import previously exported request logs
            </p>
          </div>
        </div>

        {importError && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-sm text-red-800">{importError}</p>
          </div>
        )}

        {importSuccess && (
          <div className="p-3 bg-green-50 border border-green-200 rounded-md">
            <p className="text-sm text-green-800">
              Requests imported successfully!
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
