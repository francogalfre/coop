import { cn } from "@/lib/utils";
import { initials, tintFor } from "@/lib/format";

export function PersonAvatar({
  name,
  avatarUrl,
  className,
}: {
  name: string;
  avatarUrl?: string;
  className?: string;
}) {
  if (avatarUrl) {
    return (
      <img
        src={avatarUrl}
        alt=""
        referrerPolicy="no-referrer"
        className={cn("rounded-full border border-border/60 object-cover", className)}
      />
    );
  }

  return (
    <span
      className={cn("grid place-items-center rounded-full font-medium text-background", className)}
      style={{ background: tintFor(name) }}
    >
      {initials(name)}
    </span>
  );
}
