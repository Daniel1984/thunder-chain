FROM golang:1.23.4-alpine AS builder

WORKDIR /app

# Cache dependencies first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o main ./cmd/mempool


FROM scratch
# Copy the statically linked binary from the builder
COPY --from=builder /app/main .

# Specify the user if you want to avoid running as root (optional)
USER 1001:1001

CMD ["/main"]
