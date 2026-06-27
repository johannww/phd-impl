module github.com/johannww/phd-impl/experiments/exp-app

go 1.26.1

require (
	github.com/golang/protobuf v1.5.4
	github.com/hyperledger/fabric-chaincode-go/v2 v2.3.0
	github.com/hyperledger/fabric-gateway v1.7.1
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.7
	github.com/johannww/phd-impl/chaincodes/common v0.0.0-00010101000000-000000000000
	golang.org/x/time v0.8.0
	google.golang.org/grpc v1.73.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-echarts/go-echarts/v2 v2.7.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jung-kurt/gofpdf v1.16.2 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/wcharczuk/go-chart/v2 v2.1.2 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	golang.org/x/image v0.18.0 // indirect
)

require (
	github.com/johannww/phd-impl/chaincodes/carbon v0.0.0-20260213220025-d70147e461cc
	github.com/miekg/pkcs11 v1.1.1 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251111163417-95abcf5c77ba // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/johannww/phd-impl/chaincodes/common => ../../chaincodes/common
