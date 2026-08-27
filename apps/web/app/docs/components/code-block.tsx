export function CodeBlock({ children, label }: { children: string; label?: string }) {
  return (
    <div className="my-4 overflow-hidden rounded-lg border border-border">
      {label && (
        <div className="border-border border-b bg-secondary/40 px-3 py-1.5 font-mono text-2xs text-muted-foreground">
          {label}
        </div>
      )}
      <pre className="overflow-x-auto p-3" style={{ background: "oklch(0.12 0 0)" }}>
        <code className="font-mono text-[13px] text-foreground/90 leading-relaxed">{children}</code>
      </pre>
    </div>
  );
}
