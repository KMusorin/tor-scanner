# Stage 1: build tor-scanner
FROM golang:1.27.0-alpine3.24 AS scanner-builder

WORKDIR /src

COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/

RUN go mod download \
    && CGO_ENABLED=0 go build -o tor-scanner ./cmd/main.go

# Stage 2: build lyrebird (obfs4 pluggable transport)
# Pin lyrebird to a specific commit for reproducible builds.
FROM golang:1.27.0-alpine3.24 AS lyrebird-builder

ARG LYBIRD_VERSION=ca49fa715c8cd78d7263d790fbb79a88657748a2

WORKDIR /lyrebird

RUN wget -q -O lyrebird.tar.gz \
        "https://github.com/gh4rib/lyrebird/archive/${LYBIRD_VERSION}.tar.gz" \
    && tar xzf lyrebird.tar.gz --strip-components=1

# proxy.golang.org intermittently returns an unreachable 302/403 for
# github.com/klauspost/compress@v1.17.11.zip. Prefer a mirror proxy first,
# then fall back to the official proxy (and direct VCS) for everything else.
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct

RUN go mod download \
    && CGO_ENABLED=0 go build -o lyrebird ./cmd/lyrebird

# Stage 3: runtime
FROM alpine:3.24

# Optional mirror override for environments where the default
# dl-cdn.alpinelinux.org is unreachable (e.g. blanket CDN blocks).
ARG ALPINE_MIRROR=https://dl-cdn.alpinelinux.org/alpine

RUN if [ -n "$ALPINE_MIRROR" ]; then \
        printf '%s/v3.24/main\n%s/v3.24/community\n' "$ALPINE_MIRROR" "$ALPINE_MIRROR" \
            > /etc/apk/repositories; \
    fi \
    && apk add --no-cache \
       ca-certificates \
       curl \
       tor \
    && mkdir -p /etc/tor \
       /etc/tor/torrc.d \
       /var/lib/tor \
       /var/run/tor \
    && chown -R tor:tor /etc/tor /var/lib/tor /var/run/tor

COPY --from=scanner-builder /src/tor-scanner /usr/local/bin/tor-scanner
COPY --from=lyrebird-builder /lyrebird/lyrebird /usr/local/bin/lyrebird

COPY docker/entrypoint.sh /entrypoint.sh
COPY docker/torrc /etc/tor/torrc
RUN chmod +x /entrypoint.sh

RUN chown tor:tor /usr/local/bin/tor-scanner /usr/local/bin/lyrebird /entrypoint.sh

USER tor

WORKDIR /home/tor

EXPOSE 9050

HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -s -x socks5h://127.0.0.1:9050 'https://check.torproject.org/api/ip' | grep -qm1 -E '"IsTor"\s*:\s*true'

ENTRYPOINT ["/entrypoint.sh"]
CMD []
