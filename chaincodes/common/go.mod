module github.com/johannww/phd-impl/chaincodes/common

go 1.26.1

require (
	github.com/golang/protobuf v1.5.4
	github.com/hyperledger/fabric-chaincode-go/v2 v2.3.1-0.20260319210430-56968fdc7833
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.7
	github.com/mmcloughlin/geohash v0.10.0
	github.com/stretchr/testify v1.11.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.79.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hyperledger/fabric-chaincode-go/v2 v2.3.1-0.20260319210430-56968fdc7833 => github.com/johannww/fabric-chaincode-go/v2 v2.0.0-20260707234614-9daa6fdac6ff
