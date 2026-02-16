# Stage 1: Build Go binary
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o retro-host .

# Stage 2: Download EmulatorJS
FROM alpine:3.20 AS emulatorjs
RUN apk add --no-cache curl p7zip
WORKDIR /tmp/ejs
RUN curl -L -o emulatorjs.7z \
    "https://github.com/EmulatorJS/EmulatorJS/releases/download/v4.2.3/4.2.3.7z" && \
    7z x emulatorjs.7z && \
    rm emulatorjs.7z
# The archive may extract to a versioned directory — find and copy the data/ folder
RUN mkdir -p /emulatorjs && \
    if [ -d "/tmp/ejs/data" ]; then \
        cp -r /tmp/ejs/data/* /emulatorjs/; \
    elif ls -d /tmp/ejs/*/data 2>/dev/null; then \
        cp -r /tmp/ejs/*/data/* /emulatorjs/; \
    else \
        cp -r /tmp/ejs/* /emulatorjs/; \
    fi && \
    ls /emulatorjs/loader.js  # Verify loader.js exists

# Stage 3: Runtime
FROM alpine:3.20

LABEL org.opencontainers.image.source="https://github.com/Infinitely-Iterable/retro-host"
LABEL org.opencontainers.image.description="Self-hosted browser-based retro gaming platform"
LABEL org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /build/retro-host /app/retro-host
COPY --from=builder /build/frontend /app/frontend
COPY --from=emulatorjs /emulatorjs /app/emulatorjs

RUN mkdir -p /data/saves /roms

ENV ROM_DIR=/roms
ENV DATA_DIR=/data
ENV PORT=6969

EXPOSE 6969

ENTRYPOINT ["/app/retro-host"]
