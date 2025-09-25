// React compatibility layer for Preact
import { h, ComponentChildren, VNode } from "preact";
import { forwardRef as preactForwardRef } from "preact/compat";

// Re-export Preact compat as React for Radix UI components
export * from "preact/compat";

// Type compatibility helpers
export type ReactNode = ComponentChildren;
export type ReactElement = VNode;

// Helper to make Radix UI components work with Preact
export function createPreactComponent<P extends object>(
  Component: any,
  displayName?: string
) {
  const PreactComponent = preactForwardRef<any, P>((props, ref) => {
    return h(Component, { ...props, ref });
  });

  if (displayName) {
    PreactComponent.displayName = displayName;
  }

  return PreactComponent;
}
