export function Mark({ size = 20, rounded = true }: { size?: number; rounded?: boolean }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      {rounded && <rect width="32" height="32" rx="8" fill="var(--card)" />}
      <circle cx="12" cy="13" r="6.5" fill="var(--agent)" />
      <circle cx="20" cy="13" r="8" fill={rounded ? "var(--card)" : "var(--background)"} />
      <circle cx="20" cy="13" r="6.5" fill="var(--human)" />
      <circle cx="16" cy="20" r="8" fill={rounded ? "var(--card)" : "var(--background)"} />
      <circle cx="16" cy="20" r="6.5" fill="var(--live)" />
    </svg>
  );
}
