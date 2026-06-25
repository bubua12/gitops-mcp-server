# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/gitops-mcp ./cmd/gitops-mcp

# ---- Runtime Stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1000 app

COPY --from=builder /bin/gitops-mcp /usr/local/bin/gitops-mcp

USER app
EXPOSE 18080

ENTRYPOINT ["gitops-mcp"]
