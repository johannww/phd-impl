module github.com/johannww/phd-impl/experiments/exp-app

go 1.26.1

require (
	github.com/golang/protobuf v1.5.4
	github.com/hyperledger/fabric v1.4.0-rc1.0.20260223074757-57f91e9b60bc
	github.com/hyperledger/fabric-chaincode-go/v2 v2.3.0
	github.com/hyperledger/fabric-gateway v1.7.1
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.7
	golang.org/x/time v0.8.0
	google.golang.org/grpc v1.73.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/hyperledger/fabric-lib-go v1.1.3-0.20240523144151-25edd1eaf5f5 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

require (
	github.com/johannww/phd-impl/chaincodes/carbon v0.0.0-20260213220025-d70147e461cc
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
	google.golang.org/protobuf v1.36.11
)

replace github.com/johannww/phd-impl/chaincodes/common => ../../chaincodes/common
