import { h } from "preact";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";

interface AppSidebarProps {
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
  { id: "environment", label: "Environment" },
  { id: "trace", label: "Trace" },
];

export function AppSidebar({
  activeTab,
  onTabChange,
  stats,
  onClearAll,
}: AppSidebarProps) {
  return (
    <Sidebar>
      <SidebarHeader className="border-b px-4 py-5">
        <h1 className="text-xl font-bold">GoVisual</h1>
        <p className="text-xs text-muted-foreground">HTTP Request Visualizer</p>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {menuItems.map((item) => (
                <SidebarMenuItem key={item.id}>
                  <SidebarMenuButton
                    onClick={() => onTabChange(item.id)}
                    isActive={activeTab === item.id}
                    className="w-full"
                  >
                    <span>{item.label}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>Quick Stats</SidebarGroupLabel>
          <SidebarGroupContent>
            <div className="space-y-3 px-2">
              <div className="flex justify-between items-baseline p-2 rounded-lg hover:bg-accent transition-colors">
                <span className="text-xs text-muted-foreground">Total</span>
                <span className="text-sm font-bold tabular-nums">
                  {stats.total}
                </span>
              </div>
              <div className="flex justify-between items-baseline p-2 rounded-lg hover:bg-accent transition-colors">
                <span className="text-xs text-muted-foreground">Success</span>
                <span className="text-sm font-bold tabular-nums">
                  {stats.successRate}%
                </span>
              </div>
              <div className="flex justify-between items-baseline p-2 rounded-lg hover:bg-accent transition-colors">
                <span className="text-xs text-muted-foreground">Avg</span>
                <span className="text-sm font-bold tabular-nums">
                  {stats.avgDuration}ms
                </span>
              </div>
            </div>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t p-4">
        <Button variant="destructive" className="w-full" onClick={onClearAll}>
          Clear All Requests
        </Button>
        <div className="text-center mt-4">
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
      </SidebarFooter>
    </Sidebar>
  );
}
