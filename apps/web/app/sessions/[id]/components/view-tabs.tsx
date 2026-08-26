"use client";

import { IconSparkles, IconTerminal } from "@/components/icons";
import { cn } from "@/lib/utils";

export type SessionViewTab = "timeline" | "terminal";

export function ViewTabs({
  active,
  onChange,
}: {
  active: SessionViewTab;
  onChange: (tab: SessionViewTab) => void;
}) {
  const tabs: { id: SessionViewTab; label: string; icon: typeof IconSparkles }[] = [
    { id: "timeline", label: "Timeline", icon: IconSparkles },
    { id: "terminal", label: "Terminal", icon: IconTerminal },
  ];

  return (
    <div className="mx-auto flex max-w-3xl gap-1 px-4 pb-2 sm:px-6">
      {tabs.map(({ id, label, icon: Icon }) => (
        <button
          key={id}
          type="button"
          aria-pressed={active === id}
          onClick={() => onChange(id)}
          className={cn(
            "inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-[12.5px] transition-colors",
            active === id
              ? "bg-secondary text-foreground"
              : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground",
          )}
        >
          <Icon size={13} />
          {label}
        </button>
      ))}
    </div>
  );
}
