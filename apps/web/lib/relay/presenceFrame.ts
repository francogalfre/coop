export type PresenceActor = { name: string };

export type PresenceFrame =
    | { kind: "presence"; type: "presence.typing"; actor: PresenceActor; active: boolean }
    | { kind: "presence"; type: "human.join"; actor: PresenceActor }
    | { kind: "presence"; type: "human.leave"; actor: PresenceActor };

export function parsePresenceFrame(value: unknown): PresenceFrame | null {
    if (typeof value !== "object" || value === null) return null;

    const record = value as Record<string, unknown>;

    if (record.kind !== "presence") return null;
    if (typeof record.type !== "string") return null;

    const actorValue = record.actor;
    if (typeof actorValue !== "object" || actorValue === null) return null;

    const name = (actorValue as Record<string, unknown>).name;
    if (typeof name !== "string") return null;

    if (record.type === "presence.typing") {
        if (typeof record.active !== "boolean") return null;
        return { kind: "presence", type: "presence.typing", actor: { name }, active: record.active };
    }
    if (record.type === "human.join" || record.type === "human.leave") {
        return { kind: "presence", type: record.type, actor: { name } };
    }

    return null;
}
