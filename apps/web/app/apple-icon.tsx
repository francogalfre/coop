import { ImageResponse } from "next/og";

export const size = { width: 180, height: 180 };
export const contentType = "image/png";

export default function AppleIcon() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          background: "#0a0a0a",
          borderRadius: 40,
          position: "relative",
        }}
      >
        <div
          style={{
            position: "absolute",
            width: 78,
            height: 78,
            borderRadius: "50%",
            background: "#63a0fa",
            left: 33,
            top: 39,
          }}
        />
        <div
          style={{
            position: "absolute",
            width: 90,
            height: 90,
            borderRadius: "50%",
            background: "#0a0a0a",
            left: 84,
            top: 33,
          }}
        />
        <div
          style={{
            position: "absolute",
            width: 78,
            height: 78,
            borderRadius: "50%",
            background: "#dcad1b",
            left: 90,
            top: 39,
          }}
        />
        <div
          style={{
            position: "absolute",
            width: 90,
            height: 90,
            borderRadius: "50%",
            background: "#0a0a0a",
            left: 60,
            top: 75,
          }}
        />
        <div
          style={{
            position: "absolute",
            width: 78,
            height: 78,
            borderRadius: "50%",
            background: "#39c35f",
            left: 66,
            top: 81,
          }}
        />
      </div>
    ),
    { ...size },
  );
}
