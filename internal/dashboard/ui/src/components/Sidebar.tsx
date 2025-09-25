import { h } from "preact";
import { Button } from "./ui/button";
import { cn } from "../lib/utils";

interface SidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  stats: {
    total: number;
    successRate: number;
    avgDuration: number;
  };
  onClearAll: () => void;
}

export function Sidebar({
  activeTab,
  onTabChange,
  stats,
  onClearAll,
}: SidebarProps) {
  const tabs = [
    { id: "dashboard", label: "Dashboard" },
    { id: "requests", label: "Requests" },
    { id: "environment", label: "Environment" },
    { id: "trace", label: "Trace" },
  ];

  return (
    <aside className="w-64 bg-background/95 backdrop-blur-md border-r h-screen flex flex-col shadow-xl">
      {/* Header */}
      <div className="p-6 border-b bg-gradient-to-r from-primary/5 to-primary/10">
        <h1 className="text-2xl font-bold bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
          GoVisual
        </h1>
        <p className="text-xs text-muted-foreground mt-1 font-medium">
          HTTP Request Visualizer
        </p>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4">
        <ul className="space-y-2">
          {tabs.map((tab) => (
            <li key={tab.id}>
              <button
                onClick={() => onTabChange(tab.id)}
                className={cn(
                  "w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all duration-200",
                  activeTab === tab.id
                    ? "bg-primary text-primary-foreground shadow-md"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground hover:translate-x-1"
                )}
              >
                <span>{tab.label}</span>
                {activeTab === tab.id && (
                  <span className="ml-auto w-1.5 h-1.5 bg-primary-foreground rounded-full animate-pulse" />
                )}
              </button>
            </li>
          ))}
        </ul>
      </nav>

      {/* Quick Stats */}
      <div className="p-4 border-t bg-muted/20">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase mb-4 tracking-wider">
          Quick Stats
        </h3>
        <div className="space-y-3">
          <div className="flex justify-between items-baseline p-2 rounded-lg hover:bg-muted/30 transition-colors">
            <p className="text-xs text-muted-foreground">Total</p>
            <p className="text-lg font-bold tabular-nums">{stats.total}</p>
          </div>
          <div className="flex justify-between items-baseline p-2 rounded-lg hover:bg-muted/30 transition-colors">
            <p className="text-xs text-muted-foreground">Success</p>
            <p className="text-lg font-bold tabular-nums">
              {stats.successRate}%
            </p>
          </div>
          <div className="flex justify-between items-baseline p-2 rounded-lg hover:bg-muted/30 transition-colors">
            <p className="text-xs text-muted-foreground">Avg</p>
            <p className="text-lg font-bold tabular-nums">
              {stats.avgDuration}ms
            </p>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="p-4 space-y-3 bg-muted/10">
        <Button
          variant="destructive"
          className="w-full hover:scale-[1.02] transition-transform"
          onClick={onClearAll}
        >
          Clear All Requests
        </Button>
        <div className="text-center">
          <p className="text-[10px] text-muted-foreground font-medium">
            VERSION 0.2.0
          </p>
          <p className="text-[10px] text-muted-foreground mt-1">
            Created by{" "}
            <a
              href="https://github.com/doganarif"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary hover:underline font-medium"
            >
              @doganarif
            </a>
          </p>
        </div>
      </div>
    </aside>
  );
}
