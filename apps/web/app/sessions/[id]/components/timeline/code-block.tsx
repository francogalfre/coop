export function CodeBlock({ label, body }: { label: string; body: string }) {
  return (
    <div className="overflow-hidden rounded-lg border border-border/70 bg-background/60">
      <div className="border-border/70 border-b px-3 py-1.5 font-mono text-3xs text-muted-foreground/70 uppercase tracking-wider">
        {label}
      </div>
      <pre className="max-h-72 overflow-auto p-3 font-mono text-xs text-foreground/85 leading-relaxed">{body}</pre>
    </div>
  );
}
