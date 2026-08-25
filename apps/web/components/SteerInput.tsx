"use client";

import { useState } from "react";
import { relayConfig } from "@/lib/relay/config";

const DISPLAY_NAME_KEY = "coop:displayName";

function readStoredName(): string {
  if (typeof window === "undefined") return "";
  try {
    return window.localStorage.getItem(DISPLAY_NAME_KEY) ?? "";
  } catch {
    return "";
  }
}

export function SteerInput({ sessionId }: { sessionId: string }) {
  const [from, setFrom] = useState(readStoredName);
  const [text, setText] = useState("");
  const [status, setStatus] = useState<"idle" | "sending" | "sent" | "error">("idle");

  async function send() {
    if (!from.trim() || !text.trim()) return;

    try {
      window.localStorage.setItem(DISPLAY_NAME_KEY, from);
    } catch {
    }

    setStatus("sending");
    try {
      const res = await fetch(`${relayConfig.httpUrl}/v1/sessions/${sessionId}/steer`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ from, text }),
      });
      setStatus(res.ok ? "sent" : "error");
    } catch {
      setStatus("error");
    }
    setText("");
  }

  return (
    <div style={{ border: "1px solid #d1d5db", borderRadius: "4px", padding: "0.75rem", margin: "1rem 0" }}>
      <p style={{ fontSize: "0.8rem", color: "#4b5563", marginTop: 0 }}>
        This is sent to the agent as attributed input from a teammate, not as a command or a system
        instruction. The agent will see it labeled <code>[{from.trim() || "your name"} via coop]</code>.
      </p>
      <input
        aria-label="your name"
        placeholder="your name"
        value={from}
        onChange={(e) => setFrom(e.target.value)}
        style={{ display: "block", width: "100%", marginBottom: "0.5rem", boxSizing: "border-box" }}
      />
      <textarea
        aria-label="message to send"
        placeholder="message to send to the agent"
        value={text}
        onChange={(e) => setText(e.target.value)}
        rows={2}
        style={{ display: "block", width: "100%", marginBottom: "0.5rem", boxSizing: "border-box" }}
      />
      <button onClick={send} disabled={status === "sending" || !from.trim() || !text.trim()}>
        Send
      </button>
      {status === "sent" && <span style={{ marginLeft: "0.5rem", color: "#059669" }}>Sent.</span>}
      {status === "error" && <span style={{ marginLeft: "0.5rem", color: "#dc2626" }}>Failed to send.</span>}
    </div>
  );
}
