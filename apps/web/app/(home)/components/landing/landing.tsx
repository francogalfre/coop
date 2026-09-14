"use client";

import { toast } from "sonner";
import { signIn } from "@/lib/auth/auth-client";
import { Container } from "@/components/layout/container";
import { SmoothScroll } from "@/components/providers/smooth-scroll";
import { Reveal } from "@/components/ui/reveal";
import { LandingHeader } from "./landing-header";
import { Hero } from "./hero";
import { LiveSession } from "./live-session";
import { HowItWorks } from "./how-it-works";
import { Capabilities } from "./capabilities";
import { Harnesses } from "./harnesses";
import { Trust } from "./trust";
import { CallToAction } from "./call-to-action";
import { LandingFooter } from "./landing-footer";

function signInWithGithub() {
  void signIn
    .social({ provider: "github", callbackURL: "/" })
    .catch(() => toast.error("GitHub sign-in did not go through. Try again."));
}

export function Landing() {
  return (
    <>
      <SmoothScroll />
      <Container>
        <LandingHeader onSignIn={signInWithGithub} />
        <main>
          <Hero onSignIn={signInWithGithub} />

          <section className="border-border/70 border-t px-6 py-14 sm:px-10 sm:py-16">
            <Reveal>
              <LiveSession />
            </Reveal>
          </section>

          <HowItWorks />
          <Capabilities />
          <Harnesses />
          <Trust />
          <CallToAction onSignIn={signInWithGithub} />
        </main>
        <LandingFooter />
      </Container>
    </>
  );
}
