FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Cache dependencies first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download
# Copy the rest of the application code
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o main ./cmd/miner


FROM scratch
COPY --from=builder /app/main .
USER 1001:1001
CMD ["/main"]
