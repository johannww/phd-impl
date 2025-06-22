module github.com/johannww/phd-impl/chaincodes/carbon

go 1.24.3

toolchain go1.24.4

require (
	github.com/IBM/idemix v0.0.0-20250313153527-832db18b9478
	github.com/Microsoft/confidential-sidecar-containers v0.0.0-20250610214904-989d452c4c7e
	github.com/golang/protobuf v1.5.4
	github.com/hyperledger/fabric-chaincode-go/v2 v2.0.0
	github.com/hyperledger/fabric-contract-api-go/v2 v2.2.0
	github.com/hyperledger/fabric-gateway v1.7.1
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.6
	github.com/johannww/phd-impl/tee_auction/go v0.0.0-20250617200929-e632967cc693
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/IBM/idemix/bccsp/schemes/weak-bb v0.0.0-20241220065751-dc7206770307 // indirect
	github.com/IBM/idemix/bccsp/types v0.0.0-20241220065751-dc7206770307 // indirect
	github.com/IBM/mathlib v0.0.3-0.20241219051532-81539b287cf5 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.13.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/hyperledger/fabric-amcl v0.0.0-20230602173724-9e02669dceb2 // indirect
	github.com/kilic/bls12-381 v0.1.0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.3 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx v1.2.31 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

require (
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	// google.golang.org/grpc v1.70.0
	// google.golang.org/protobuf v1.36.5
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.72.2
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hyperledger/fabric-chaincode-go/v2 v2.0.0 => github.com/johannww/fabric-chaincode-go/v2 v2.0.0-20250520213033-d967d6ea1875

// replace github.com/hyperledger/fabric-gateway v1.7.1 => /home/johann/prj/dtr/fabric-repos/fabric-gateway/

replace github.com/hyperledger/fabric-gateway v1.7.1 => github.com/johannww/fabric-gateway v0.0.0-20250527175102-91efdaca92d8

// NOTE: This is temporary until microsoft fixes serialization
replace github.com/Microsoft/confidential-sidecar-containers v0.0.0-20250610214904-989d452c4c7e => github.com/johannww/confidential-sidecar-containers v0.0.0-20250603023458-bbd0bf198f91
