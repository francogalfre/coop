# Deployment

coop is two long-running services — the **web** viewer and the **relay** — plus
a **Postgres** database and a **GitHub OAuth app**. The CLI is not deployed;
users download a binary from GitHub Releases and point it at the relay with
`COOP_RELAY_URL`.

There are two ways to run the services. Pick one.

---

## Option A — one domain, Docker Compose

Everything on one host, web and relay path-routed under a single domain. This
is the simplest option and what `docker-compose.yml` is built for.

### With Coolify

Add this repo as a **Docker Compose** resource. Coolify builds `web` and
`relay` from `docker-compose.yml` and proxies both through its own Traefik
instance using the `traefik.*` labels already in the file — no separate proxy
to run.

1. Point `COOP_DOMAIN`'s DNS A/AAAA record at your Coolify server.
2. Set the variables from `.env.production.example` on the resource:
   `COOP_DOMAIN`, `DATABASE_URL`, `COOP_GITHUB_CLIENT_ID` /
   `COOP_GITHUB_CLIENT_SECRET`, `COOP_INTERNAL_SECRET`, `BETTER_AUTH_SECRET`.
3. Create the GitHub OAuth app with callback
   `https://<COOP_DOMAIN>/api/auth/callback/github`.
4. Deploy. Coolify provisions TLS on the first request.

### Anywhere else

`docker compose up -d --build` works, but you supply your own reverse proxy —
the compose file carries Traefik labels, not a bundled proxy. Postgres is
external (Neon, Supabase, a managed instance); bring your own `DATABASE_URL`.

---

## Option B — web on Vercel, relay on a VPS

The Next.js app on Vercel, the Go relay on a VPS (Coolify), sharing one
Postgres.

### Why it needs a shared parent domain

The relay authenticates browsers by reading the better-auth session cookie
straight off the request. That cookie is host-only by default, so a relay on a
different host never receives it. Widening it to a shared parent domain fixes
that:

```
app.example.com     → Vercel    viewer, auth, session verify
relay.example.com   → VPS       /v1/*, /healthz, WebSockets
                      both       → one Postgres
cookie Domain=example.com  → the browser sends the session to both
```

Both hosts must be subdomains of a domain you control — a bare `*.vercel.app`
can't share a cookie with the VPS. `COOP_COOKIE_DOMAIN=.example.com` turns this
on; unset, only Option A works.

### One-time setup

1. **DNS** — `app` → Vercel, `relay` → VPS IP.
2. **Postgres** — one database reachable from both, `sslmode=require`.
3. **Web schema** — run once, and again after any schema change (Vercel does
   not run migrations): `cd apps/web && DATABASE_URL='…' bunx drizzle-kit push`.
   The relay migrates its own tables on boot.
4. **GitHub OAuth app** — homepage `https://app.example.com`, callback
   `https://app.example.com/api/auth/callback/github`. One client id/secret,
   used by both services.
5. **Secrets** — `COOP_INTERNAL_SECRET` (shared web ↔ relay) and
   `BETTER_AUTH_SECRET` (`openssl rand -base64 32`).

### Relay on Coolify

New resource → Docker Compose → `apps/relay/docker-compose.yml`.

| variable | value |
| --- | --- |
| `DATABASE_URL` | shared Postgres |
| `COOP_WEB_ORIGINS` | `https://app.example.com` |
| `COOP_WEB_INTERNAL_URL` | `https://app.example.com` |
| `COOP_INTERNAL_SECRET` | shared with Vercel |
| `COOP_GITHUB_CLIENT_ID` / `COOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app |
| `COOP_RELAY_DOMAIN` | `relay.example.com` |

Single replica only. Confirm the Traefik `entrypoints` / `certresolver` label
names match your Coolify instance (Servers → Proxy → Configuration).

### Web on Vercel

Root Directory `apps/web`, build command `cd ../.. && bunx turbo run build --filter=@coop/web`.

| variable | value |
| --- | --- |
| `DATABASE_URL` | shared Postgres |
| `BETTER_AUTH_SECRET` | `openssl rand -base64 32` |
| `BETTER_AUTH_URL` / `NEXT_PUBLIC_BETTER_AUTH_URL` | `https://app.example.com` |
| `NEXT_PUBLIC_COOP_RELAY_URL` | `https://relay.example.com` |
| `COOP_COOKIE_DOMAIN` | `.example.com` |
| `COOP_INTERNAL_SECRET` | shared with the relay |
| `COOP_GITHUB_CLIENT_ID` / `COOP_GITHUB_CLIENT_SECRET` | GitHub OAuth app |

`NEXT_PUBLIC_*` vars are read at build time — redeploy after changing them.

### Order

1. Deploy the relay, check `https://relay.example.com/healthz`.
2. Deploy the web app.
3. Sign in at `https://app.example.com`, open a session, and confirm the
   session WebSocket to `relay.example.com` connects without a 401.

---

## The CLI

Release binaries are published to GitHub Releases on every `vX.Y.Z` tag. Users
point them at your relay:

```bash
export COOP_RELAY_URL=https://relay.example.com   # or https://<COOP_DOMAIN> for Option A
coop attach
```
