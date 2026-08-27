"use client";

import "@xterm/xterm/css/xterm.css";
import { useEffect, useRef, useState } from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { IconTerminal } from "@/components/icons";
import { usePtyStream } from "@/lib/relay/usePtyStream";
import { cn } from "@/lib/utils";

function readToken(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}

function buildTheme() {
  return {
    background: readToken("--card"),
    foreground: readToken("--foreground"),
    cursor: readToken("--live"),
    selectionBackground: readToken("--agent"),
  };
}

function EmptyState() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-20 text-center">
      <span className="grid size-11 place-items-center rounded-xl border border-border bg-card text-muted-foreground">
        <IconTerminal size={19} />
      </span>
      <div className="space-y-1">
        <p className="font-display font-medium text-md text-foreground">No live terminal</p>
        <p className="max-w-xs text-sm text-muted-foreground leading-relaxed">
          Waiting for a pty to connect. Output will appear here as soon as one streams in.
        </p>
      </div>
    </div>
  );
}

export function PtyTerminal({
  sessionId,
  heldByMe,
  visible = true,
}: {
  sessionId: string;
  heldByMe: boolean;
  visible?: boolean;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const [hasReceivedOutput, setHasReceivedOutput] = useState(false);
  const { onOutput, onResize, sendInput, sendResize } = usePtyStream(sessionId);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const terminal = new Terminal({
      fontFamily: "var(--font-mono)",
      fontSize: 13,
      cursorBlink: true,
      theme: buildTheme(),
    });
    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(container);
    fitAddon.fit();

    terminalRef.current = terminal;
    fitAddonRef.current = fitAddon;

    return () => {
      terminal.dispose();
      terminalRef.current = null;
      fitAddonRef.current = null;
    };
  }, []);

  useEffect(() => {
    const terminal = terminalRef.current;
    if (terminal) terminal.options.disableStdin = !heldByMe;
  }, [heldByMe]);

  useEffect(() => {
    return onOutput((bytes) => {
      setHasReceivedOutput(true);
      terminalRef.current?.write(bytes);
    });
  }, [onOutput]);

  useEffect(() => {
    return onResize((cols, rows) => {
      setHasReceivedOutput(true);
      terminalRef.current?.resize(cols, rows);
    });
  }, [onResize]);

  useEffect(() => {
    const terminal = terminalRef.current;
    if (!terminal || !heldByMe) return;
    const disposable = terminal.onData((data) => {
      sendInput(new TextEncoder().encode(data));
    });
    return () => disposable.dispose();
  }, [heldByMe, sendInput]);

  useEffect(() => {
    const container = containerRef.current;
    const fitAddon = fitAddonRef.current;
    if (!container || !fitAddon) return;

    const observer = new ResizeObserver(() => {
      fitAddon.fit();
      if (!heldByMe) return;
      const terminal = terminalRef.current;
      if (terminal) sendResize(terminal.cols, terminal.rows);
    });
    observer.observe(container);
    return () => observer.disconnect();
  }, [heldByMe, sendResize]);

  return (
    <div
      id="panel-terminal"
      role="tabpanel"
      aria-labelledby="tab-terminal"
      className={cn("relative flex-1 overflow-hidden", !visible && "hidden")}
    >
      {!hasReceivedOutput && (
        <div className="absolute inset-0 z-10 bg-background">
          <EmptyState />
        </div>
      )}
      <div ref={containerRef} className="mx-auto h-full w-full max-w-3xl px-4 py-3 sm:px-6" />
    </div>
  );
}
