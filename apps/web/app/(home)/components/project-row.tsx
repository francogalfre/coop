import Link from "next/link";
import type { Route } from "next";
import { motion } from "motion/react";
import type { Project } from "@/lib/relay/api";
import { IconChevronRight } from "@/components/icons";
import { relativeTime } from "@/lib/format";

export function ProjectRow({ project, index }: { project: Project; index: number }) {
  return (
    <motion.li
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.04, ease: [0.16, 1, 0.3, 1] }}
    >
      <Link
        href={`/projects/${project.slug}` as Route}
        className="group flex items-center gap-4 rounded-xl border border-border bg-card/50 px-5 py-4 transition-all hover:border-border/80 hover:bg-card"
      >
        <div className="min-w-0 flex-1">
          <p className="truncate font-display font-medium text-md text-foreground">
            {project.name}
          </p>
          <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
            {project.slug} · created {relativeTime(project.created_at)}
          </p>
        </div>
        <IconChevronRight
          size={16}
          className="shrink-0 text-muted-foreground/50 transition-transform group-hover:translate-x-0.5 group-hover:text-foreground"
        />
      </Link>
    </motion.li>
  );
}
