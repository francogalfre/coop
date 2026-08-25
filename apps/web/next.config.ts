import type { NextConfig } from "next";

const config: NextConfig = {
  // Workspace packages ship TypeScript source, not a prebuilt bundle.
  transpilePackages: ["@coop/protocol"],
  typedRoutes: true,
};

export default config;
