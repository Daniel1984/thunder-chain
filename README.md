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
   "hash": "28290ba82d30fefca8b3424e32149b0ad08398362e961a45492ba7aa36e49991",
   "from_addr": "0x1124bb912381CA2774c228F1DF7cc209AcEE576F",
   "to_addr": "0x7217d3eC0A0C357d7Dde4896094B83137c137E42",
   "amount": 1000,
   "fee": 10,
   "nonce": 1,
   "signature": "f3f342d758203da6f879ed9b946a22ac1c80a76375dbbff16a04acdbe9cc4ac0513d7177114e327709489cc35d207d886cd35bd5e37dd05fe23c40443400b6cd00"
 }'
 ```
