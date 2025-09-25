import { h, ComponentChildren } from "preact";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

interface SimpleSidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  stats: {
    total: number;
    successRate: number;
    avgDuration: number;
  };
  onClearAll: () => void;
}

const menuItems = [
  { id: "dashboard", label: "Dashboard" },
  { id: "requests", label: "Requests" },
  { id: "analytics", label: "Analytics" },
  { id: "environment", label: "Environment" },
  { id: "trace", label: "Trace" },
];

export function SimpleSidebar({
  activeTab,
  onTabChange,
  stats,
  onClearAll,
}: SimpleSidebarProps) {
  return (
    <aside className="w-64 bg-white border-r h-screen flex flex-col shadow-sm">
      {/* Header */}
      <div className="px-6 py-5 border-b">
        <h1 className="text-xl font-bold">GoVisual</h1>
        <p className="text-xs text-muted-foreground mt-1">
          HTTP Request Visualizer
        </p>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-4 py-6">
        <div className="space-y-1">
          {menuItems.map((item) => (
            <button
              key={item.id}
              onClick={() => onTabChange(item.id)}
              className={cn(
                "w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all duration-200",
                activeTab === item.id
                  ? "bg-slate-900 text-white shadow-sm"
                  : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
              )}
            >
              <span>{item.label}</span>
              {activeTab === item.id && (
                <span className="ml-auto w-1.5 h-1.5 bg-white rounded-full" />
              )}
            </button>
          ))}
        </div>
      </nav>

      {/* Quick Stats */}
      <div className="px-4 py-4 border-t bg-slate-50/50">
        <h3 className="text-xs font-semibold text-slate-500 uppercase mb-3 tracking-wider">
          Quick Stats
        </h3>
        <div className="space-y-2">
          <div className="flex justify-between items-baseline px-3 py-2 rounded-lg hover:bg-white transition-colors">
            <span className="text-xs text-slate-500">Total</span>
            <span className="text-sm font-bold tabular-nums">
              {stats.total}
            </span>
          </div>
          <div className="flex justify-between items-baseline px-3 py-2 rounded-lg hover:bg-white transition-colors">
            <span className="text-xs text-slate-500">Success</span>
            <span className="text-sm font-bold tabular-nums">
              {stats.successRate}%
            </span>
          </div>
          <div className="flex justify-between items-baseline px-3 py-2 rounded-lg hover:bg-white transition-colors">
            <span className="text-xs text-slate-500">Avg</span>
            <span className="text-sm font-bold tabular-nums">
              {stats.avgDuration}ms
            </span>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="p-4 border-t space-y-3">
        <Button variant="destructive" className="w-full" onClick={onClearAll}>
          Clear All Requests
        </Button>
        <div className="text-center">
          <p className="text-[10px] text-slate-500 font-medium">
            VERSION 0.2.0
          </p>
          <p className="text-[10px] text-slate-500 mt-1">
            Created by{" "}
            <a
              href="https://github.com/doganarif"
              target="_blank"
              rel="noopener noreferrer"
              className="text-slate-900 hover:underline font-medium"
            >
              @doganarif
            </a>
          </p>
        </div>
      </div>
    </aside>
  );
}
