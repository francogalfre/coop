import type { ReactNode } from "react";
import { AppHeader } from "@/components/app-header";
import { DocsNav } from "./components/docs-nav";

export default function DocsLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-dvh bg-background">
      <AppHeader />
      <div className="mx-auto flex max-w-5xl gap-10 px-6 py-10">
        <aside className="hidden w-48 shrink-0 md:block">
          <div className="sticky top-24">
            <DocsNav />
          </div>
        </aside>
        <main className="min-w-0 flex-1 pb-24">{children}</main>
      </div>
    </div>
  );
}
