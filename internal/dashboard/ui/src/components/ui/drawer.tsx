import { h, ComponentChildren } from "preact";
import { useEffect, useRef } from "preact/hooks";
import { cn } from "../../lib/utils";

interface DrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: ComponentChildren;
  className?: string;
}

interface DrawerContentProps {
  children: ComponentChildren;
  className?: string;
  onClose: () => void;
  isFullscreen: boolean;
  onToggleFullscreen: () => void;
}

export function Drawer({
  open,
  onOpenChange,
  children,
  className,
}: DrawerProps) {
  useEffect(() => {
    if (open) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [open]);

  if (!open) return null;

  return (
    <div className={cn("drawer-container", className)}>
      {/* Backdrop */}
      <div
        className={cn(
          "fixed inset-0 bg-black/50 backdrop-blur-sm z-40",
          "animate-in fade-in duration-200"
        )}
        onClick={() => onOpenChange(false)}
      />
      {/* Content */}
      <div className="fixed inset-0 z-50 pointer-events-none">{children}</div>
    </div>
  );
}

export function DrawerContent({
  children,
  className,
  onClose,
  isFullscreen,
  onToggleFullscreen,
}: DrawerContentProps) {
  const contentRef = useRef<HTMLDivElement>(null);
  const dragHandleRef = useRef<HTMLDivElement>(null);
  const startY = useRef<number>(0);
  const currentY = useRef<number>(0);
  const isDragging = useRef<boolean>(false);

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging.current || isFullscreen) return;
      const deltaY = e.clientY - startY.current;
      currentY.current = Math.max(0, deltaY);

      if (contentRef.current) {
        contentRef.current.style.transform = `translateY(${currentY.current}px)`;
      }
    };

    const handleMouseUp = () => {
      if (!isDragging.current) return;
      isDragging.current = false;

      if (currentY.current > 200) {
        onClose();
      } else if (contentRef.current) {
        contentRef.current.style.transform = "";
        contentRef.current.style.transition = "transform 0.3s ease";
        setTimeout(() => {
          if (contentRef.current) {
            contentRef.current.style.transition = "";
          }
        }, 300);
      }
    };

    const handleMouseDown = (e: MouseEvent) => {
      if (isFullscreen) return;
      isDragging.current = true;
      startY.current = e.clientY;
      currentY.current = 0;
    };

    const handle = dragHandleRef.current;
    if (handle) {
      handle.addEventListener("mousedown", handleMouseDown);
    }

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);

    return () => {
      if (handle) {
        handle.removeEventListener("mousedown", handleMouseDown);
      }
      document.removeEventListener("mousemove", handleMouseMove);
      document.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isFullscreen, onClose]);

  return (
    <div
      ref={contentRef}
      className={cn(
        "fixed bg-background rounded-t-2xl shadow-2xl pointer-events-auto",
        "animate-in slide-in-from-bottom duration-300",
        isFullscreen
          ? "inset-0 rounded-none"
          : "inset-x-0 bottom-0 max-h-[85vh]",
        className
      )}
    >
      {/* Drag Handle */}
      {!isFullscreen && (
        <div
          ref={dragHandleRef}
          className="absolute top-0 left-0 right-0 h-8 cursor-ns-resize flex items-center justify-center"
        >
          <div className="w-12 h-1 bg-muted-foreground/30 rounded-full" />
        </div>
      )}

      {/* Header */}
      <div className="sticky top-0 bg-background/95 backdrop-blur-sm border-b z-10 px-6 py-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">Request Details</h3>
          <div className="flex items-center gap-2">
            <button
              onClick={onToggleFullscreen}
              className="p-2 rounded-lg hover:bg-accent transition-colors"
              aria-label={isFullscreen ? "Exit fullscreen" : "Enter fullscreen"}
            >
              {isFullscreen ? (
                <svg
                  className="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              ) : (
                <svg
                  className="w-4 h-4"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 8V4m0 0h4M4 4l5 5m11-5h-4m4 0v4m0 0l-5-5M4 16v4m0 0h4M4 20l5-5m11 5h-4m4 0v-4m0 0l-5 5"
                  />
                </svg>
              )}
            </button>
            <button
              onClick={onClose}
              className="p-2 rounded-lg hover:bg-accent transition-colors"
              aria-label="Close drawer"
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div
        className={cn(
          "overflow-y-auto",
          isFullscreen ? "h-[calc(100vh-4rem)]" : "max-h-[calc(85vh-4rem)]"
        )}
      >
        <div className="p-6">{children}</div>
      </div>
    </div>
  );
}

export function DrawerHeader({
  children,
  className,
}: {
  children: ComponentChildren;
  className?: string;
}) {
  return <div className={cn("mb-4", className)}>{children}</div>;
}

export function DrawerTitle({
  children,
  className,
}: {
  children: ComponentChildren;
  className?: string;
}) {
  return (
    <h2 className={cn("text-2xl font-semibold", className)}>{children}</h2>
  );
}

export function DrawerDescription({
  children,
  className,
}: {
  children: ComponentChildren;
  className?: string;
}) {
  return (
    <p className={cn("text-muted-foreground mt-1", className)}>{children}</p>
  );
}
