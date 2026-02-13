module github.com/johannww/phd-impl/chaincodes/interop

go 1.24.4

require (
	github.com/hyperledger/fabric-chaincode-go/v2 v2.0.0
	github.com/hyperledger/fabric-contract-api-go/v2 v2.2.0
	github.com/johannww/phd-impl/chaincodes/carbon v0.0.0-20250831185049-8495b62111bf
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/Microsoft/confidential-sidecar-containers v0.0.0-20250820140128-4814b442cf71 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.6 // indirect
	github.com/johannww/phd-impl/tee_auction/go v0.0.0-20250617200929-e632967cc693 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx v1.2.31 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mmcloughlin/geohash v0.10.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tidwall/geoindex v1.7.0 // indirect
	github.com/tidwall/rtree v1.10.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hyperledger/fabric-chaincode-go/v2 v2.0.0 => github.com/johannww/fabric-chaincode-go/v2 v2.0.0-20250520213033-d967d6ea1875

replace github.com/johannww/phd-impl/chaincodes/carbon v0.0.0-20250831185049-8495b62111bf => ../carbon

replace github.com/johannww/phd-impl/tee_auction/go v0.0.0-20250617200929-e632967cc693 => ../../tee_auction/go
