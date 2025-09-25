import { h } from "preact";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

interface StatsProps {
  requests: any[];
}

export function StatsDashboard({ requests }: StatsProps) {
  const calculateStats = () => {
    const total = requests.length;
    const success = requests.filter(
      (r) => r.StatusCode >= 200 && r.StatusCode < 300
    ).length;
    const redirect = requests.filter(
      (r) => r.StatusCode >= 300 && r.StatusCode < 400
    ).length;
    const clientError = requests.filter(
      (r) => r.StatusCode >= 400 && r.StatusCode < 500
    ).length;
    const serverError = requests.filter((r) => r.StatusCode >= 500).length;
    const avgDuration =
      total > 0
        ? Math.round(requests.reduce((sum, r) => sum + r.Duration, 0) / total)
        : 0;

    return { total, success, redirect, clientError, serverError, avgDuration };
  };

  const stats = calculateStats();

  const statCards = [
    {
      label: "Total Requests",
      value: stats.total,
    },
    {
      label: "Success (2xx)",
      value: stats.success,
    },
    {
      label: "Redirect (3xx)",
      value: stats.redirect,
    },
    {
      label: "Client Error (4xx)",
      value: stats.clientError,
    },
    {
      label: "Server Error (5xx)",
      value: stats.serverError,
    },
    {
      label: "Avg Response Time",
      value: `${stats.avgDuration}ms`,
    },
  ];

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-6">
      {statCards.map((stat, index) => (
        <Card
          key={index}
          className="hover:shadow-lg transition-all duration-300 hover:-translate-y-1 cursor-default"
          style={{ animationDelay: `${index * 50}ms` }}
        >
          <CardContent className="p-6">
            <div className="text-3xl font-bold mb-2 tabular-nums">
              {stat.value}
            </div>
            <p className="text-xs text-muted-foreground font-medium uppercase tracking-wider">
              {stat.label}
            </p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
