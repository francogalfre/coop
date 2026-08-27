# syntax=docker/dockerfile:1
#
# Build context must be the repo root (bun workspace: apps/web depends on
# the workspace package @coop/protocol, which only resolves if the whole
# workspace is present). Build with:
#   docker build -f apps/web/Dockerfile .

FROM oven/bun:1.2.17-alpine AS build
WORKDIR /repo

COPY . .
# --ignore-scripts: root package.json's "prepare" script sets up git hooks
# (`git config core.hooksPath .githooks`), which needs a .git dir that isn't
# present in the build context. Nothing in this workspace needs postinstall.
RUN bun install --frozen-lockfile --ignore-scripts

ARG NEXT_PUBLIC_COOP_RELAY_URL
ARG NEXT_PUBLIC_BETTER_AUTH_URL
ENV NEXT_PUBLIC_COOP_RELAY_URL=${NEXT_PUBLIC_COOP_RELAY_URL}
ENV NEXT_PUBLIC_BETTER_AUTH_URL=${NEXT_PUBLIC_BETTER_AUTH_URL}

RUN cd apps/web && bun run build

# No `output: "standalone"` in next.config.ts, so the production image keeps
# the full workspace node_modules and .next build directory rather than a
# pruned standalone bundle.
FROM oven/bun:1.2.17-alpine AS runtime
WORKDIR /repo

ENV NODE_ENV=production
RUN addgroup -S coop && adduser -S coop -G coop

COPY --from=build /repo /repo
USER coop

WORKDIR /repo/apps/web
EXPOSE 3000
CMD ["bun", "run", "start"]
