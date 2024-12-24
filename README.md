### Instructions

#### Generating gRPC deps:
1. `make build-proto-builder` to build protoc container with all deps
2. `make generate-protoc` to generate all proto files


#### Misc
calling grpc:
```
grpcurl -plaintext -d '{"transaction": {"from_addr": "sender", "to_addr": "recipient", "amount": 100}}' localhost:8181 transaction.TransactionService/CreateTransaction

grpcurl -plaintext -d '{"id": "abc"}' localhost:8181 transaction.TransactionService/DeleteTransaction
```
