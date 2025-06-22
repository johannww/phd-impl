cc:
	go build -o /dev/null ./cmd/carbon/main.go
	
cc-docker:
	docker build -t ghcr.io/johannww/phd-impl/carbon:latest .

app:
	go build -o /dev/null ./cmd/application/main.go

test:
	go test ./tests

test-no-cache:
	go test -count=1 ./tests
