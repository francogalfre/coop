# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src

COPY apps/relay/go.mod apps/relay/go.sum ./
RUN go mod download

COPY apps/relay/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/relay ./cmd/relay

# distroless/static has no shell or wget, which rules out a Docker/Compose
# HEALTHCHECK (exec-only, no external prober) against GET /healthz — alpine
# trades a slightly larger image for that.
FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget && \
    addgroup -S coop && adduser -S coop -G coop
WORKDIR /app
COPY --from=build /out/relay ./relay

EXPOSE 8787
USER coop:coop
ENTRYPOINT ["/app/relay"]
