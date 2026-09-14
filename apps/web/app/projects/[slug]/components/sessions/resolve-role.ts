import type { Project } from "@/lib/relay/api";

export type ProjectRole = "owner" | "member";

export function resolveProjectRole(
  projects: Project[],
  slug: string,
  userId: string | undefined,
): ProjectRole {
  if (userId === undefined) return "member";
  const project = projects.find((p) => p.slug === slug);
  return project && project.created_by === userId ? "owner" : "member";
}
