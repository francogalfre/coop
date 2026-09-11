# Deployment

coop in production is two long-running services — the **web** viewer
(Next.js) and the **relay** (Go) — sharing one external **Postgres** database
and one **GitHub OAuth app**. The CLI is not deployed; it is built from
`packages/cli` and pointed at the relay with `COOP_RELAY_URL`.

## Required environment variables

From `.env.production.example`:

| Variable | Who reads it | Meaning |
| --- | --- | --- |
| `COOP_DOMAIN` | compose file only | the domain everything is served from, no scheme, no trailing slash |
| `DATABASE_URL` | relay + web | the Postgres connection string (`sslmode=require`). Same database for both. |
| `COOP_GITHUB_CLIENT_ID` / `COOP_GITHUB_CLIENT_SECRET` | relay + web | GitHub OAuth app credentials. Callback URL: `https://<COOP_DOMAIN>/api/auth/callback/github`. |
| `COOP_INTERNAL_SECRET` | relay + web | shared secret for relay → web server-to-server calls (session verify). Generate with `openssl rand -base64 32`. Must match on both services. |
| `BETTER_AUTH_SECRET` | web | better-auth signing secret. Generate with `openssl rand -base64 32`. |

The relay additionally reads `COOP_RELAY_ADDR` (default `:8787`) and
`COOP_WEB_ORIGINS` (CORS / WebSocket origins). In the compose topology below
these are set for you; they matter for the split topology.

## Topology: one domain, one host

`docker-compose.yml` at the repo root runs both services on a single
`COOP_DOMAIN`, routed by path: `/v1/*` and `/healthz` go to the relay,
everything else to the web app. There is no bundled proxy — the file carries
Traefik labels and expects the platform's proxy (Coolify's managed Traefik,
or your own) to read them.

Why one domain and not two subdomains: the relay authenticates browsers by
reading the better-auth session cookie, and that cookie is host-only — it is
only sent to the exact origin the browser signed in on. Path-routing under
one host is what makes cookie sharing work without extra machinery.

Notes the compose file encodes:

- **The relay is strictly single-instance.** It keeps the event store,
  steering mailbox, presence, takeover, pty, steer-request, and question
  state in memory. Never raise it beyond one replica.
- `NEXT_PUBLIC_COOP_RELAY_URL` and `NEXT_PUBLIC_BETTER_AUTH_URL` are build
  args of the web image (`NEXT_PUBLIC_*` values are baked in at build time —
  rebuild after changing them).

### On Coolify

1. Add this repo as a **Docker Compose** resource. Coolify builds `web` and
   `relay` from `docker-compose.yml` and proxies both through its own
   Traefik using the labels already in the file.
2. Point `COOP_DOMAIN`'s DNS A/AAAA record at the server.
3. Set the variables from the table above on the resource.
4. Create the GitHub OAuth app with callback
   `https://<COOP_DOMAIN>/api/auth/callback/github`.
5. Deploy. TLS is provisioned on first request. Verify the Traefik
   `entrypoints` / `certresolver` names against your instance
   (Servers → Proxy → Configuration) if they differ.

### Anywhere else

`docker compose up -d --build` works with any reverse proxy that honors
Traefik labels or that you configure by hand using the same rules
(`Host(<COOP_DOMAIN>)` + `PathPrefix(/v1)` or `Path(/healthz)` → relay port
8787; everything else → web port 3000). Bring your own Postgres; the
compose file runs none.

## Split topology: web and relay on different hosts

The relay can own a whole subdomain instead: `apps/relay/docker-compose.yml`
builds just the relay, with `COOP_RELAY_DOMAIN` as its host. This is what you
use when the web app lives elsewhere (for example a managed Next.js host)
and the relay lives on a VPS.

The constraint is the same cookie from above: for the relay to see the
browser's session cookie, both hosts must be subdomains of a domain you
control, and the web app must widen the cookie with
`COOP_COOKIE_DOMAIN=.example.com` (see `apps/web/lib/auth/auth.ts`). A bare
`*.vercel.app` host cannot share a cookie with your VPS.

Relay side (from `apps/relay/docker-compose.yml`): `DATABASE_URL`,
`COOP_WEB_ORIGINS=https://app.example.com`, `COOP_WEB_INTERNAL_URL` (where
the relay calls the web app's internal session-verify endpoint),
`COOP_INTERNAL_SECRET`, and the GitHub client id/secret.

Web side (from `apps/web/.env.example`): `DATABASE_URL`,
`COOP_GITHUB_CLIENT_ID`/`COOP_GITHUB_CLIENT_SECRET`,
`COOP_INTERNAL_SECRET`, `BETTER_AUTH_SECRET`, `BETTER_AUTH_URL` /
`NEXT_PUBLIC_BETTER_AUTH_URL`, `NEXT_PUBLIC_COOP_RELAY_URL=https://relay.example.com`,
and `COOP_COOKIE_DOMAIN=.example.com`.

## Database migrations

The two services own different tables in the same database:

- **Relay (Ent):** migrates itself on every boot (`db.Pool.Migrate`). Nothing
  to run by hand.
- **Web (Drizzle):** does not migrate on boot. Run once after deploying, and
  again after any schema change:

  ```bash
  cd apps/web
  DATABASE_URL='postgres://…' bun run db:push
  ```

## Before going live

- [ ] `https://<COOP_DOMAIN>/healthz` returns 200 (it pings Postgres).
- [ ] The GitHub OAuth app's callback is exactly
      `https://<COOP_DOMAIN>/api/auth/callback/github`.
- [ ] `DATABASE_URL` uses `sslmode=require` and both services point at the
      same database.
- [ ] `COOP_INTERNAL_SECRET` and `BETTER_AUTH_SECRET` are freshly generated
      random values, and `COOP_INTERNAL_SECRET` matches on both services.
- [ ] TLS works on the domain (the relay's WebSocket needs `wss` behind
      HTTPS).
- [ ] The relay runs as exactly one replica.
- [ ] The web app's drizzle schema has been pushed.
- [ ] Sign in, create a project, attach a session from a machine with the
      CLI, and confirm the session WebSocket connects and steering reaches
      the agent.

## The CLI in production

Release binaries are published to GitHub Releases on every `vX.Y.Z` tag
(`.github/workflows/release.yml` builds linux, darwin, and windows, amd64 and
arm64). Users download one and point it at the relay:

```bash
export COOP_RELAY_URL=https://<COOP_DOMAIN>
export COOP_GITHUB_CLIENT_ID=<public client id of the OAuth app>
coop login
```

Building from source also works:

```bash
cd packages/cli && go build -o coop ./cmd/coop
```
