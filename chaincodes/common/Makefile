.PHONY: protos
protos:
	@echo "generating go protobufs into $(shell pwd)/internal/pb"
	# Ensure output directory exists
	mkdir -p pb
	protoc -I protos --go_out=pb --go_opt=paths=source_relative \
		protos/*.proto

.PHONY: clean-protos
clean-protos:
	@echo "cleaning generated protos in internal/pb"
	rm -rf pb/*.pb.go

unit-test:
	go test ./...
