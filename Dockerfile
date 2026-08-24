# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o nanoflux ./cmd/api

# Final stage - distroless (non-root, minimal)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/nanoflux .

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["./nanoflux"]