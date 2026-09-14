import type { ReactNode } from "react";
import { cn } from "@/lib/cn";
import { Reveal } from "@/components/ui/reveal";

export function SectionHeading({
    eyebrow,
    title,
    lede,
    className,
}: {
    eyebrow: string;
    title: ReactNode;
    lede?: ReactNode;
    className?: string;
}) {
    return (
        <Reveal className={cn("max-w-2xl", className)}>
            <p className="font-mono text-2xs text-muted-foreground/70 uppercase tracking-[0.18em]">
                {eyebrow}
            </p>
            <h2 className="mt-4 text-balance text-editorial text-[32px] text-foreground leading-[1.04] sm:text-[44px]">
                {title}
            </h2>
            {lede && (
                <p className="mt-4 max-w-xl text-balance text-[15px] text-muted-foreground leading-relaxed">
                    {lede}
                </p>
            )}
        </Reveal>
    );
}
