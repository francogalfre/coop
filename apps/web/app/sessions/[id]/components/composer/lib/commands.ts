import { HARNESS_COMMANDS, type Capabilities, type HarnessCommand } from "@coop/protocol";
import type { SendTarget } from "../target-toggle";

export type CommandScope = "coop" | "harness";

export type CommandContext = {
  args: string;
  isOwner: boolean;
  heldByMe: boolean;
  takeoverHeldBy?: string;
  setTarget: (target: SendTarget) => void;
  ask: (question: string) => Promise<void>;
  setTakeover: (active: boolean) => Promise<void>;
  setMode: (mode: "auto" | "restricted") => Promise<void>;
  runHarnessCommand: (command: HarnessCommand, args?: string) => Promise<void>;
};

export type Command = {
  name: string;
  scope: CommandScope;
  ownerOnly: boolean;
  description: string;
  args?: "text" | "enum";
  enumValues?: readonly string[];
  available: (caps: Capabilities | undefined, isOwner: boolean) => true | string;
  run: (ctx: CommandContext) => Promise<void>;
};

const HARNESS_UNAVAILABLE = "needs coop run — attach can't type into your TUI";

const HARNESS_DESCRIPTIONS: Record<HarnessCommand, string> = {
  model: "switch the harness's active model",
  compact: "compact the conversation history",
  clear: "clear the conversation",
  context: "show context usage",
  cost: "show session cost",
  status: "show harness status",
};

function ownerOnlyAvailable(_caps: Capabilities | undefined, isOwner: boolean): true | string {
  return isOwner ? true : "owner only";
}

export const COMMANDS: Command[] = [
  {
    name: "team",
    scope: "coop",
    ownerOnly: false,
    description: "switch this message to the team channel",
    available: () => true,
    run: async (ctx) => ctx.setTarget("team"),
  },
  {
    name: "agent",
    scope: "coop",
    ownerOnly: false,
    description: "switch this message to the agent",
    available: () => true,
    run: async (ctx) => {
      if (ctx.takeoverHeldBy) {
        throw new Error(`${ctx.takeoverHeldBy} has taken over — the agent is paused`);
      }
      ctx.setTarget("agent");
    },
  },
  {
    name: "mode",
    scope: "coop",
    ownerOnly: true,
    description: "set session mode: auto or restricted",
    args: "enum",
    enumValues: ["auto", "restricted"],
    available: ownerOnlyAvailable,
    run: async (ctx) => {
      const mode = ctx.args.trim();
      if (mode !== "auto" && mode !== "restricted") throw new Error("usage: /mode auto|restricted");
      await ctx.setMode(mode);
    },
  },
  {
    name: "takeover",
    scope: "coop",
    ownerOnly: false,
    description: "take over the session — the agent pauses",
    available: () => true,
    run: async (ctx) => ctx.setTakeover(true),
  },
  {
    name: "release",
    scope: "coop",
    ownerOnly: false,
    description: "release your takeover",
    available: () => true,
    run: async (ctx) => {
      if (!ctx.heldByMe) throw new Error("you don't hold the takeover");
      await ctx.setTakeover(false);
    },
  },
  {
    name: "ask",
    scope: "coop",
    ownerOnly: false,
    description: "ask the agent a question",
    args: "text",
    available: () => true,
    run: async (ctx) => {
      const question = ctx.args.trim();
      if (!question) throw new Error("usage: /ask <question>");
      await ctx.ask(question);
    },
  },
  ...HARNESS_COMMANDS.map<Command>((name) => ({
    name,
    scope: "harness",
    ownerOnly: true,
    description: HARNESS_DESCRIPTIONS[name],
    args: "text",
    available: (caps, isOwner) => {
      if (!isOwner) return "owner only";
      if (!caps?.commands) return HARNESS_UNAVAILABLE;
      return true;
    },
    run: async (ctx) => {
      await ctx.runHarnessCommand(name, ctx.args.trim() || undefined);
    },
  })),
];

export function parseCommandInvocation(raw: string): { name: string; args: string } | null {
  const trimmed = raw.trim();
  const match = /^\/(\S+)\s*([\s\S]*)$/.exec(trimmed);
  if (!match) return null;
  return { name: (match[1] ?? "").toLowerCase(), args: match[2] ?? "" };
}

export function filterCommands(commands: Command[], query: string): Command[] {
  const q = query.trim().toLowerCase();
  if (!q) return commands;
  return commands
    .filter((c) => c.name.includes(q))
    .toSorted((a, b) => Number(b.name.startsWith(q)) - Number(a.name.startsWith(q)));
}
