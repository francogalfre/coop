import type { Metadata } from "next";
import type { ReactNode } from "react";
import { AuthStatus } from "@/components/AuthStatus";

export const metadata: Metadata = {
  title: "coop",
  description: "Multiplayer sessions for coding agents",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header style={{ display: "flex", justifyContent: "flex-end", padding: "0.75rem 1rem" }}>
          <AuthStatus />
        </header>
        {children}
      </body>
    </html>
  );
}
