import type { ReactNode } from "react";
import { AppHeader } from "@/components/layout/app-header";
import { Container } from "@/components/layout/container";
import { DocsNav } from "./components/docs-nav";

export default function DocsLayout({ children }: { children: ReactNode }) {
  return (
    <Container>
      <AppHeader />
      <div className="flex gap-10 px-6 py-10 sm:px-10">
        <aside className="hidden w-48 shrink-0 md:block">
          <div className="sticky top-24">
            <DocsNav />
          </div>
        </aside>
        <main className="min-w-0 flex-1 pb-24">{children}</main>
      </div>
    </Container>
  );
}
