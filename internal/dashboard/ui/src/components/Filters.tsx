import { h } from "preact";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Button } from "./ui/button";

interface FiltersProps {
  onFilterChange: (filters: FilterState) => void;
  onClear: () => void;
}

export interface FilterState {
  method: string;
  statusCode: string;
  path: string;
  minDuration: string;
}

export function Filters({ onFilterChange, onClear }: FiltersProps) {
  const handleFilterChange = () => {
    const filters: FilterState = {
      method: (document.getElementById("method-filter") as HTMLSelectElement)?.value || "",
      statusCode: (document.getElementById("status-filter") as HTMLSelectElement)?.value || "",
      path: (document.getElementById("path-filter") as HTMLInputElement)?.value || "",
      minDuration: (document.getElementById("duration-filter") as HTMLInputElement)?.value || ""
    };
    onFilterChange(filters);
  };

  const handleReset = () => {
    (document.getElementById("method-filter") as HTMLSelectElement).value = "";
    (document.getElementById("status-filter") as HTMLSelectElement).value = "";
    (document.getElementById("path-filter") as HTMLInputElement).value = "";
    (document.getElementById("duration-filter") as HTMLInputElement).value = "";
    onFilterChange({
      method: "",
      statusCode: "",
      path: "",
      minDuration: ""
    });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Filters</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
          {/* HTTP Method */}
          <div>
            <label htmlFor="method-filter" className="block text-sm font-medium mb-1">
              HTTP Method
            </label>
            <select
              id="method-filter"
              className="w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              onChange={handleFilterChange}
            >
              <option value="">All Methods</option>
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
              <option value="PATCH">PATCH</option>
              <option value="HEAD">HEAD</option>
              <option value="OPTIONS">OPTIONS</option>
            </select>
          </div>

          {/* Status Code */}
          <div>
            <label htmlFor="status-filter" className="block text-sm font-medium mb-1">
              Status Code
            </label>
            <select
              id="status-filter"
              className="w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              onChange={handleFilterChange}
            >
              <option value="">All Status Codes</option>
              <option value="2xx">2xx Success</option>
              <option value="3xx">3xx Redirect</option>
              <option value="4xx">4xx Client Error</option>
              <option value="5xx">5xx Server Error</option>
            </select>
          </div>

          {/* Path Contains */}
          <div>
            <label htmlFor="path-filter" className="block text-sm font-medium mb-1">
              Path Contains
            </label>
            <input
              type="text"
              id="path-filter"
              placeholder="Filter by path..."
              className="w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              onInput={handleFilterChange}
            />
          </div>

          {/* Min Duration */}
          <div>
            <label htmlFor="duration-filter" className="block text-sm font-medium mb-1">
              Min Duration (ms)
            </label>
            <input
              type="number"
              id="duration-filter"
              placeholder="Min duration..."
              className="w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              onInput={handleFilterChange}
            />
          </div>
        </div>

        <div className="flex gap-2">
          <Button onClick={handleFilterChange}>Apply Filters</Button>
          <Button variant="secondary" onClick={handleReset}>Reset</Button>
          <Button variant="destructive" onClick={onClear}>Clear All Requests</Button>
        </div>
      </CardContent>
    </Card>
  );
}
