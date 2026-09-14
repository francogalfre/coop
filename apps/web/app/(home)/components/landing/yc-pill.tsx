function IconYCombinator({ size = 14 }: { size?: number }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      aria-hidden
      fill="currentColor"
    >
      <path d="M11.06 12.99 6.44 4.5h2.2l2.86 5.63c.06.13.12.26.18.4l.16.4.16-.4c.06-.14.12-.27.19-.4L15.02 4.5h2.08l-4.62 8.56v6.44h-1.42v-6.51Z" />
    </svg>
  );
}

export function YCPill() {
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-border/70 bg-card/60 py-1 pr-3.5 pl-1 backdrop-blur-sm">
      <span className="flex size-5 items-center justify-center rounded-full bg-[#ff6600] text-white">
        <IconYCombinator />
      </span>
      <span className="font-mono text-2xs text-muted-foreground tracking-tight">
        Not backed by Y Combinator
        <span className="text-foreground/90"> yet</span>
      </span>
    </span>
  );
}
