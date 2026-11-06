# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hosting .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates openssh-client tzdata \
    && adduser -D -h /app hosting \
    && mkdir -p /app/.ssh \
    && chown -R hosting:hosting /app \
    && chmod 700 /app/.ssh

WORKDIR /app
COPY --from=builder /out/hosting /app/hosting
COPY templates /app/templates
COPY static /app/static

USER hosting
EXPOSE 8021

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 \
  CMD wget -q -O /dev/null http://127.0.0.1:8021/healthz || exit 1

ENTRYPOINT ["/app/hosting"]
