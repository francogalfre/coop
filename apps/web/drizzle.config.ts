import { defineConfig } from "drizzle-kit";

export default defineConfig({
  dialect: "postgresql",
  schema: "./lib/auth/schema.ts",
  out: "./lib/auth/migrations",
  dbCredentials: {
    url: process.env.DATABASE_URL as string,
  },
});
