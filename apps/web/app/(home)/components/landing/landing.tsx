"use client";

import { toast } from "sonner";
import { signIn } from "@/lib/auth/auth-client";
import { Container } from "@/components/layout/container";
import { SmoothScroll } from "@/components/providers/smooth-scroll";
import { LandingHeader } from "./landing-header";
import { Hero } from "./hero";
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
