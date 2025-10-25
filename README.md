### Instructions

#### Generating gRPC deps:

1. `make build-proto-builder` to build protoc container with all deps
2. `make generate-protoc` to generate all proto files

#### Misc

**Calling grpc:**

```sh
grpcurl -plaintext -d '{"transaction": {"from_addr": "sender", "to_addr": "recipient", "amount": 100}}' localhost:8181 transaction.TransactionService/CreateTransaction

grpcurl -plaintext -d '{"id": "abc"}' localhost:8181 transaction.TransactionService/DeleteTransaction
```

#### Start services on host:

```sh
API_PORT=8181 DB_PATH=./cmd/mempool/data/mempool.db go run ./cmd/mempool
MEMPOOL_API=localhost:8181 STATE_API=localhost:8383 go run ./cmd/miner
API_PORT=8383 DB_PATH=./cmd/state/data/state.db go run ./cmd/state
MEMPOOL_API=localhost:8181 API_PORT=8080 STATE_API=localhost:8383 go run ./cmd/node
```

#### Dev flow:

Sample keys (also in genesis file):

```
Private Key: a6f7fa0885f49b8327376bdcc1da167750ec8004b1331705358c7fb697a74fbb
Address: 0xE07cD67682C4b43bEF6b399bb7C180D975571aaa

Private Key: 3d6d88ea612a2096cab4b8be2d41b3d10ceaf1e024557d26a39d0d62723e663d
Address: 0x76F86614A08683bDFd4a44Df1Ee24E94Bf5c19b2
```

Sign transaction:

```sh
go run ./cmd/cli/main.go sign-tx \
  --from 0xE07cD67682C4b43bEF6b399bb7C180D975571aaa \
  --to 0x76F86614A08683bDFd4a44Df1Ee24E94Bf5c19b2 \
  --amount 999 \
  --fee 5 \
  --nonce 1 \
  --private-key a6f7fa0885f49b8327376bdcc1da167750ec8004b1331705358c7fb697a74fbb
```

Submit signed transaction:

```sh
curl -X POST http://localhost:8080/transactions \
 -H "Content-Type: application/json" \
 -d '{
"hash": "53383375632170df16eba5d0659c2d19d06b6cfaca225cb22cd9b6615b59241a",
"from_addr": "0xE07cD67682C4b43bEF6b399bb7C180D975571aaa",
"to_addr": "0x76F86614A08683bDFd4a44Df1Ee24E94Bf5c19b2",
"signature": "88feb99af78cc63d5ad8e2218ca84d9fda70c2bd739fc156a7ed605c4df9a333143e00a87371c3ad0eb604923f7f4ad98707e8200e12d8129b89f707f8c8c5d801",
"amount": 999,
"fee": 5,
"nonce": 1
}'
```
