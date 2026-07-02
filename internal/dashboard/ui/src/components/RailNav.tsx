import { h } from "preact";
import { cn } from "@/lib/utils";
import { useTheme } from "@/hooks/useTheme";

// View enumerates every top-level destination the rail can switch between.
// Keep this in lockstep with the list of items below; App.tsx narrows on it.
export type View =
  | "inbox"
  | "errors"
  | "slow"
  | "analytics"
  | "environment";

interface RailNavProps {
  active: View;
  onChange: (view: View) => void;
  errorCount?: number;
}

interface Item {
  id: View;
  label: string;
  icon: h.JSX.Element;
}

const items: Item[] = [
  {
    id: "inbox",
    label: "Inbox",
    icon: (
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 12h-6l-2 3h-4l-2-3H2" />
        <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" />
      </svg>
    ),
  },
  {
    id: "errors",
    label: "Errors",
    icon: (
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="12" />
        <line x1="12" y1="16" x2="12.01" y2="16" />
      </svg>
    ),
  },
  {
    id: "slow",
    label: "Slow",
    icon: (
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <polyline points="12 6 12 12 16 14" />
      </svg>
    ),
  },
  {
    id: "analytics",
    label: "Analytics",
    icon: (
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="18" y1="20" x2="18" y2="10" />
        <line x1="12" y1="20" x2="12" y2="4" />
        <line x1="6" y1="20" x2="6" y2="14" />
      </svg>
    ),
  },
  {
    id: "environment",
    label: "Environment",
    icon: (
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <rect x="2" y="3" width="20" height="14" rx="2" />
        <line x1="8" y1="21" x2="16" y2="21" />
        <line x1="12" y1="17" x2="12" y2="21" />
      </svg>
    ),
  },
];

export function RailNav({ active, onChange, errorCount = 0 }: RailNavProps) {
  const [theme, toggleTheme] = useTheme();
  return (
    <aside class="w-14 border-r border-zinc-200 bg-white flex flex-col items-center py-3 gap-1 shrink-0">
      <a
        href="https://github.com/doganarif/GoVisual"
        target="_blank"
        rel="noopener noreferrer"
        title="GoVisual on GitHub"
        class="w-8 h-8 rounded bg-zinc-900 text-white flex items-center justify-center text-sm font-bold mb-4"
      >
        G
      </a>
      {items.map((item) => {
        const isActive = active === item.id;
        return (
          <button
            key={item.id}
            onClick={() => onChange(item.id)}
            title={item.label}
            class={cn(
              "w-9 h-9 rounded-md flex items-center justify-center relative",
              isActive
                ? "bg-zinc-100 text-zinc-900"
                : "text-zinc-500 hover:bg-zinc-100 hover:text-zinc-900"
            )}
          >
            {item.icon}
            {item.id === "errors" && errorCount > 0 && (
              <span class="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 rounded-full bg-red-500 text-white text-[10px] font-medium flex items-center justify-center">
                {errorCount > 99 ? "99+" : errorCount}
              </span>
            )}
          </button>
        );
      })}
      <div class="flex-1" />
      <button
        onClick={toggleTheme}
        title={theme === "dark" ? "Switch to light theme" : "Switch to dark theme"}
        class="w-9 h-9 rounded-md hover:bg-zinc-100 flex items-center justify-center text-zinc-500"
      >
        {theme === "dark" ? (
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="5" />
            <line x1="12" y1="1" x2="12" y2="3" />
            <line x1="12" y1="21" x2="12" y2="23" />
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
            <line x1="1" y1="12" x2="3" y2="12" />
            <line x1="21" y1="12" x2="23" y2="12" />
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
          </svg>
        ) : (
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
          </svg>
        )}
      </button>
      <a
        href="https://github.com/doganarif/GoVisual"
        target="_blank"
        rel="noopener noreferrer"
        title="View source"
        class="w-9 h-9 rounded-md hover:bg-zinc-100 flex items-center justify-center text-zinc-500"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.387.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.4 3-.405 1.02.005 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12" />
        </svg>
      </a>
    </aside>
  );
}
