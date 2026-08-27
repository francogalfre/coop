import type { Metadata } from "next";
import type { ReactNode } from "react";
import { Bricolage_Grotesque, Geist, Geist_Mono } from "next/font/google";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Toaster } from "@/components/ui/sonner";
import "./globals.css";

const geistSans = Geist({
  subsets: ["latin"],
  variable: "--font-geist-sans",
  display: "swap",
});

const geistMono = Geist_Mono({
  subsets: ["latin"],
  variable: "--font-geist-mono",
  display: "swap",
});

const bricolage = Bricolage_Grotesque({
  subsets: ["latin"],
  variable: "--font-bricolage",
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_BETTER_AUTH_URL ?? "http://localhost:3000"),
  title: "coop — multiplayer coding agents",
  description:
    "One person runs a coding agent. Teammates open a link and watch it work, redirect it mid-task, and hand off control — live, together.",
  openGraph: {
    title: "coop — multiplayer coding agents",
    description:
      "Watch a live coding agent, redirect it mid-task, and hand off control — together.",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "coop — multiplayer coding agents",
    description:
      "Watch a live coding agent, redirect it mid-task, and hand off control — together.",
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className={`dark ${geistSans.variable} ${geistMono.variable} ${bricolage.variable}`}
      suppressHydrationWarning
    >
      <body className="min-h-dvh antialiased">
        <TooltipProvider delayDuration={200}>{children}</TooltipProvider>
        <Toaster position="bottom-right" />
      </body>
    </html>
  );
}
