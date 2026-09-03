# Split deploy: web on Vercel, relay on a VPS

The repo-root `docker-compose.yml` runs web and relay together under one domain.
This is the alternative: the Next.js app on Vercel, the Go relay on a VPS
(Coolify), talking to one shared Postgres.

## Why it needs a shared parent domain

The relay authenticates browser traffic by reading the better-auth session
cookie straight off the request. That cookie is host-only by default, so a
relay on a different host never receives it. The fix is to widen the cookie to
a shared parent domain:

```
app.example.com     -> Vercel    Next.js: viewer, auth, /api/internal/session/verify
relay.example.com   -> Coolify   Go relay: /v1/*, /healthz, WebSockets
                       both      -> one Postgres
cookie Domain=example.com  -> browser sends the session to both hosts
```

Both hosts must be subdomains of a domain you control. A bare `*.vercel.app`
cannot share a cookie with the VPS.

`COOP_COOKIE_DOMAIN=.example.com` turns this on (`apps/web/lib/auth/auth.ts`).
Unset, the cookie stays host-only and only the co-hosted compose works.

The relay accepts both `better-auth.session_token` and the
`__Secure-`-prefixed name better-auth issues over HTTPS.

## One-time setup

1. **DNS** — `app` → Vercel, `relay` → VPS IP.
2. **Postgres** — one database reachable from both (Coolify service, Neon,
   Supabase, …), `sslmode=require`.
3. **Web schema** — run once against the production database, and again after
   any schema change (Vercel does not run migrations):
   ```
   cd apps/web && DATABASE_URL='...' bunx drizzle-kit push
   ```
   The relay migrates its own tables on boot.
4. **GitHub OAuth app** — Homepage `https://app.example.com`, callback
   `https://app.example.com/api/auth/callback/github`. One client id/secret,
   used by both services.
5. **Secrets** — `COOP_INTERNAL_SECRET` (shared web <-> relay) and
   `BETTER_AUTH_SECRET` (`openssl rand -base64 32`).

## Relay on Coolify

New resource → Docker Compose → `apps/relay/docker-compose.yml`. Environment:

| var | value |
| --- | --- |
| `DATABASE_URL` | shared Postgres |
| `COOP_WEB_ORIGINS` | `https://app.example.com` |
| `COOP_WEB_INTERNAL_URL` | `https://app.example.com` |
| `COOP_INTERNAL_SECRET` | shared with Vercel |
| `COOP_GITHUB_CLIENT_ID` / `COOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app |
| `COOP_RELAY_DOMAIN` | `relay.example.com` |

Single replica only — the relay holds in-memory state and must never be scaled.
Confirm the Traefik `entrypoints` / `certresolver` label names match your
Coolify instance (Servers → Proxy → Configuration).

## Web on Vercel

Root Directory `apps/web`. Build Command:

```
cd ../.. && bunx turbo run build --filter=@coop/web
```

Environment (Production):

| var | value |
| --- | --- |
| `DATABASE_URL` | shared Postgres |
| `BETTER_AUTH_SECRET` | `openssl rand -base64 32` |
| `BETTER_AUTH_URL` | `https://app.example.com` |
| `NEXT_PUBLIC_BETTER_AUTH_URL` | `https://app.example.com` |
| `NEXT_PUBLIC_COOP_RELAY_URL` | `https://relay.example.com` |
| `COOP_COOKIE_DOMAIN` | `.example.com` |
| `COOP_INTERNAL_SECRET` | shared with the relay |
| `COOP_GITHUB_CLIENT_ID` / `COOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app |

`NEXT_PUBLIC_*` vars are read at build time — redeploy after changing them.

## Order

1. Deploy the relay, check `https://relay.example.com/healthz`.
2. Deploy the web app.
3. Sign in at `https://app.example.com`, open a session, and confirm the
   session WebSocket to `relay.example.com` connects without a 401.

## CLI

Users point the CLI at the relay:

```
export COOP_RELAY_URL=https://relay.example.com
coop attach
```

Binaries come from GitHub Releases on every `vX.Y.Z` tag.
