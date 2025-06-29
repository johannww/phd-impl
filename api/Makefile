install-openapi-codegen:
	go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

generate-api:
	go generate ./api/generate.go

docs:
	redocly build-docs ./api/openapi.yaml -o docs.html

run:
	go run ./cmd/api/main.go
