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

**Submitting transaction to node api:**
```sh
curl -X POST http://localhost:8080/transactions \
 -H "Content-Type: application/json" \
 -d '{
   "hash": "41b2f983b9621be56ff5aa4149d989172d80e3aea79679bae06635bf953dc1da",
   "from_addr": "0x820176689e3e24B3731baA88f242094b79421663",
   "to_addr": "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
   "amount": 1000,
   "fee": 10,
   "nonce": 1,
   "signature": "481c177aa5154279583397d3f17525ce471dff3bcdab5542dd8f3c8ab4736fe56c617a1cd0274a670ac39de71a4c727b80a76fb32737ade3317c7797d567d41e01"
 }'
 ```
