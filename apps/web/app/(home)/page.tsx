"use client";

import { useSession } from "@/lib/auth/auth-client";
import { SignedOut } from "./components/signed-out";
import { SignedIn } from "./components/signed-in";

export default function Home() {
  const { data, isPending } = useSession();

  if (isPending) return <div className="min-h-dvh bg-background" />;
  return data?.user ? <SignedIn /> : <SignedOut />;
}
