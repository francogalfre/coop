import { cn } from "@/lib/cn";

const WIDTH = 1200;
const HEIGHT = 460;
const SAMPLES = 240;

type Peak = { center: number; spread: number; height: number };

type Ridge = {
  id: string;
  peaks: Peak[];
  ripple: number;
  opacity: number;
  fade: number;
};

// Gaussian peaks read as mountains where a plain sine sum reads as waves; the
// ripple term only roughens the slopes so the silhouette never looks extruded.
const RIDGES: Ridge[] = [
  {
    id: "far",
    peaks: [
      { center: 0.22, spread: 0.16, height: 268 },
      { center: 0.54, spread: 0.13, height: 196 },
      { center: 0.82, spread: 0.18, height: 312 },
    ],
    ripple: 9,
    opacity: 0.5,
    fade: 0.76,
  },
  {
    id: "mid",
    peaks: [
      { center: 0.36, spread: 0.14, height: 356 },
      { center: 0.68, spread: 0.11, height: 262 },
      { center: 0.94, spread: 0.15, height: 214 },
    ],
    ripple: 13,
    opacity: 0.78,
    fade: 0.86,
  },
  {
    id: "near",
    peaks: [
      { center: 0.12, spread: 0.12, height: 196 },
      { center: 0.47, spread: 0.1, height: 152 },
      { center: 0.75, spread: 0.13, height: 224 },
    ],
    ripple: 16,
    opacity: 1,
    fade: 0.94,
  },
];

function elevation(t: number, ridge: Ridge): number {
  let sum = 0;

  for (const peak of ridge.peaks) {
    const offset = (t - peak.center) / peak.spread;
    sum += peak.height * Math.exp(-offset * offset);
  }

  sum += ridge.ripple * Math.sin(t * 27.3 + ridge.peaks.length);
  sum += ridge.ripple * 0.6 * Math.sin(t * 51.7 + 1.7);

  return Math.max(sum, 0);
}

function ridgePath(ridge: Ridge): string {
  const points: string[] = [];

  for (let i = 0; i <= SAMPLES; i++) {
    const t = i / SAMPLES;
    const x = (t * WIDTH).toFixed(1);
    const y = (HEIGHT - elevation(t, ridge)).toFixed(1);
    points.push(`${x} ${y}`);
  }

  return `M0 ${HEIGHT} L${points.join(" L")} L${WIDTH} ${HEIGHT} Z`;
}

export function DottedMountain({ className }: { className?: string }) {
  return (
    <svg
      viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
      preserveAspectRatio="xMidYMax slice"
      aria-hidden
      className={cn("pointer-events-none select-none", className)}
    >
      <defs>
        <pattern
          id="mountain-dots"
          width="5"
          height="5"
          patternUnits="userSpaceOnUse"
        >
          <circle cx="1.3" cy="1.3" r="1.05" fill="currentColor" />
        </pattern>

        {RIDGES.map((ridge) => (
          <linearGradient
            key={ridge.id}
            id={`mountain-fade-${ridge.id}`}
            x1="0"
            y1="0"
            x2="0"
            y2="1"
          >
            <stop offset="0%" stopColor="white" stopOpacity="1" />
            <stop
              offset={`${ridge.fade * 100}%`}
              stopColor="white"
              stopOpacity="0.45"
            />
            <stop offset="100%" stopColor="white" stopOpacity="0.2" />
          </linearGradient>
        ))}

        <linearGradient id="mountain-edges" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stopColor="white" stopOpacity="0" />
          <stop offset="14%" stopColor="white" stopOpacity="1" />
          <stop offset="86%" stopColor="white" stopOpacity="1" />
          <stop offset="100%" stopColor="white" stopOpacity="0" />
        </linearGradient>

        {RIDGES.map((ridge) => (
          <mask key={ridge.id} id={`mountain-mask-${ridge.id}`}>
            <rect
              width={WIDTH}
              height={HEIGHT}
              fill={`url(#mountain-fade-${ridge.id})`}
            />
            <rect
              width={WIDTH}
              height={HEIGHT}
              fill="url(#mountain-edges)"
              style={{ mixBlendMode: "multiply" }}
            />
          </mask>
        ))}
      </defs>

      {RIDGES.map((ridge) => (
        <path
          key={ridge.id}
          d={ridgePath(ridge)}
          fill="url(#mountain-dots)"
          opacity={ridge.opacity}
          mask={`url(#mountain-mask-${ridge.id})`}
        />
      ))}
    </svg>
  );
}
