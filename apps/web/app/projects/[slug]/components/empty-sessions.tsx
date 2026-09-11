import { AttachOnboarding } from "./attach-onboarding";
import { type ProjectRole } from "./resolve-role";

export function EmptySessions({
  slug,
  role,
}: {
  slug: string;
  role: ProjectRole;
}) {
  return <AttachOnboarding slug={slug} role={role} />;
}
