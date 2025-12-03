FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM alpine:3.20

RUN addgroup -g 1000 nonroot && \
    adduser -D -s /bin/sh -u 1000 -G nonroot nonroot

USER nonroot:nonroot

WORKDIR /

COPY --from=builder /app/server /server

EXPOSE 8080

ENV PORT=8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health-check || exit 1

ENTRYPOINT ["/server"]