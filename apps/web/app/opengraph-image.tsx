import { ImageResponse } from "next/og";

export const alt = "coop — multiplayer coding agents";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

// oklch() isn't parsed by satori's bundled color parser — converted from
// globals.css tokens (--background, --foreground, --muted-foreground, --live).
const background = "#0a0a0a";
const foreground = "#fafafa";
const muted = "#919191";
const border = "rgba(255, 255, 255, 0.12)";
const live = "#39c35f";
const agent = "#63a0fa";
const human = "#dcad1b";

export default function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          background,
          padding: 80,
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            fontSize: 22,
            color: muted,
            border: `1px solid ${border}`,
            borderRadius: 999,
            padding: "8px 20px",
            alignSelf: "flex-start",
          }}
        >
          <div style={{ width: 9, height: 9, borderRadius: 999, background: live }} />
          Multiplayer coding agents
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 22 }}>
            <div style={{ display: "flex", position: "relative", width: 88, height: 88 }}>
              <div
                style={{
                  position: "absolute",
                  width: 44,
                  height: 44,
                  borderRadius: "50%",
                  background: agent,
                  left: 6,
                  top: 12,
                }}
              />
              <div
                style={{
                  position: "absolute",
                  width: 52,
                  height: 52,
                  borderRadius: "50%",
                  background,
                  left: 34,
                  top: 4,
                }}
              />
              <div
                style={{
                  position: "absolute",
                  width: 44,
                  height: 44,
                  borderRadius: "50%",
                  background: human,
                  left: 38,
                  top: 12,
                }}
              />
              <div
                style={{
                  position: "absolute",
                  width: 52,
                  height: 52,
                  borderRadius: "50%",
                  background,
                  left: 20,
                  top: 38,
                }}
              />
              <div
                style={{
                  position: "absolute",
                  width: 44,
                  height: 44,
                  borderRadius: "50%",
                  background: live,
                  left: 24,
                  top: 44,
                }}
              />
            </div>
            <div style={{ display: "flex", fontSize: 96, fontWeight: 600, color: foreground }}>
              coop
            </div>
          </div>
          <div style={{ display: "flex", fontSize: 34, color: muted, maxWidth: 900 }}>
            Watch a live coding agent, redirect it mid-task, and hand off control — together.
          </div>
        </div>
      </div>
    ),
    size,
  );
}
