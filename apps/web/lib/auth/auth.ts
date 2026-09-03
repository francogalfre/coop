import { betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { nextCookies } from "better-auth/next-js";
import { db } from "@/database/client";
import * as schema from "@/database/schema";

export const auth = betterAuth({
    secret: process.env.BETTER_AUTH_SECRET,
    baseURL: process.env.BETTER_AUTH_URL,
    database: drizzleAdapter(db, { provider: "pg", schema }),

    socialProviders: {
        github: {
            clientId: process.env.COOP_GITHUB_CLIENT_ID as string,
            clientSecret: process.env.COOP_GITHUB_CLIENT_SECRET as string,
            verifyIdToken: async (token: string) => {
                const res = await fetch("https://api.github.com/user", {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        "User-Agent": "coop",
                    },
                });
                return res.ok;
            },
        },
    },
    plugins: [nextCookies()],

    // A shared parent domain (".example.com") lets the session cookie reach a
    // relay on a sibling subdomain; unset keeps it host-only for single-origin.
    advanced: process.env.COOP_COOKIE_DOMAIN
        ? { crossSubDomainCookies: { enabled: true, domain: process.env.COOP_COOKIE_DOMAIN } }
        : undefined,
});
