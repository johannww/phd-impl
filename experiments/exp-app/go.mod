module github.com/johannww/phd-impl/experiments/exp-app

go 1.26.1

require (
	github.com/hyperledger/fabric-gateway v1.7.1
	golang.org/x/time v0.5.0
	google.golang.org/grpc v1.73.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.7 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/johannww/phd-impl/chaincodes/common => ../../chaincodes/common
