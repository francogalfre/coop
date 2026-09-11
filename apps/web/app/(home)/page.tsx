"use client";

import { useSession } from "@/lib/auth/auth-client";
import { Landing } from "./components/landing/landing";
import { SignedIn } from "./components/signed-in";

export default function Home() {
  const { data, isPending } = useSession();

  if (isPending) return <div className="min-h-dvh bg-background" />;
  return data?.user ? <SignedIn /> : <Landing />;
}
