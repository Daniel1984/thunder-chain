build-proto-builder:
	docker build -t protoc-builder -f Dockerfile.grpc .

generate-protoc:
	docker run --rm \
		-v $(CURDIR)/proto:/proto \
		protoc-builder \
		--proto_path=/proto \
		--go_out=/proto --go_opt=paths=source_relative \
		--go-grpc_out=/proto --go-grpc_opt=paths=source_relative \
		$(shell find $(CURDIR)/proto -name '*.proto' -exec basename {} \;)
