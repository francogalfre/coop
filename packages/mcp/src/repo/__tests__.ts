import { execFileSync } from "node:child_process";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { basename, join } from "node:path";
import { describe, expect, it } from "vitest";
import { detectRepo } from "./identity.js";

function scratch(): string {
  return mkdtempSync(join(tmpdir(), "coop-repoid-"));
}

function git(dir: string, ...args: string[]): void {
  execFileSync("git", ["-C", dir, ...args], { stdio: "ignore" });
}

describe("detectRepo", () => {
  it("falls back to the directory basename outside a git repo", () => {
    const dir = scratch();
    expect(detectRepo(dir)).toBe(basename(dir));
  });

  it("falls back to the toplevel basename when there is no remote", () => {
    const dir = scratch();
    git(dir, "init");
    expect(detectRepo(dir)).toBe(basename(dir));
  });

  it("normalizes an ssh remote the same way the Go CLI does", () => {
    const dir = scratch();
    git(dir, "init");
    git(dir, "remote", "add", "origin", "git@github.com:francogalfre/coop.git");
    expect(detectRepo(dir)).toBe("github.com/francogalfre/coop");
  });

  it("normalizes an https remote", () => {
    const dir = scratch();
    git(dir, "init");
    git(dir, "remote", "add", "origin", "https://github.com/francogalfre/coop.git");
    expect(detectRepo(dir)).toBe("github.com/francogalfre/coop");
  });
});
