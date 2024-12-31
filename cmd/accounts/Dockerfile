FROM golang:1.23.4-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o main ./cmd/accounts


FROM scratch
# Copy the statically linked binary from the builder
COPY --from=builder /app/main .

# Specify the user if you want to avoid running as root (optional)
USER 1001:1001

CMD ["/main"]
