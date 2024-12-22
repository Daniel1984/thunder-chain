FROM alpine:3.18 as builder

# Install zig
RUN apk add --no-cache zig

WORKDIR /app
COPY . .

# Build
RUN zig build mempool

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/zig-out/bin/mempool /app/mempool

EXPOSE 8081
CMD ["/app/mempool"]
