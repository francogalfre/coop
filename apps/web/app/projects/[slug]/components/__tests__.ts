import { describe, expect, it } from "vitest";
import { buildAttachCommand } from "./attach-command";
import { resolveProjectRole } from "./resolve-role";

describe("buildAttachCommand", () => {
  it("builds the attach command for a project slug", () => {
    expect(buildAttachCommand("my-app")).toBe("coop attach --project=my-app");
  });

  it("keeps slugs with hyphens intact", () => {
    expect(buildAttachCommand("billing-api")).toBe(
      "coop attach --project=billing-api",
    );
  });
});

describe("resolveProjectRole", () => {
  const ownerId = "u_owner";
  const memberId = "u_member";

  const projects = [
    {
      id: 1,
      name: "My App",
      slug: "my-app",
      created_by: ownerId,
      created_at: "2026-01-01T00:00:00Z",
    },
  ];

  it("returns owner when the user created the project", () => {
    expect(resolveProjectRole(projects, "my-app", ownerId)).toBe("owner");
  });

  it("returns member when the user did not create the project", () => {
    expect(resolveProjectRole(projects, "my-app", memberId)).toBe("member");
  });

  it("returns member while the auth session is still pending", () => {
    expect(resolveProjectRole(projects, "my-app", undefined)).toBe("member");
  });

  it("returns member when the project is not found", () => {
    expect(resolveProjectRole(projects, "missing", ownerId)).toBe("member");
  });
});
