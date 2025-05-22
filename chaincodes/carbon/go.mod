module github.com/johannww/phd-impl/chaincodes/carbon

go 1.23.1

require (
	github.com/IBM/idemix v0.0.0-20250313153527-832db18b9478
	github.com/golang/protobuf v1.5.4
	github.com/hyperledger/fabric-chaincode-go/v2 v2.0.0
	github.com/hyperledger/fabric-contract-api-go/v2 v2.2.0
	github.com/hyperledger/fabric-gateway v1.7.1
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.6
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.7.0
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/IBM/idemix/bccsp/schemes/weak-bb v0.0.0-20241220065751-dc7206770307 // indirect
	github.com/IBM/idemix/bccsp/types v0.0.0-20241220065751-dc7206770307 // indirect
	github.com/IBM/mathlib v0.0.3-0.20241219051532-81539b287cf5 // indirect
	github.com/bits-and-blooms/bitset v1.13.0 // indirect
	github.com/consensys/bavard v0.1.13 // indirect
	github.com/consensys/gnark-crypto v0.13.0 // indirect
	github.com/hyperledger/fabric-amcl v0.0.0-20230602173724-9e02669dceb2 // indirect
	github.com/kilic/bls12-381 v0.1.0 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	// google.golang.org/grpc v1.70.0
	// google.golang.org/protobuf v1.36.5
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/grpc v1.70.0
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hyperledger/fabric-chaincode-go/v2 v2.0.0 => github.com/johannww/fabric-chaincode-go/v2 v2.0.0-20250520213033-d967d6ea1875

replace github.com/hyperledger/fabric-gateway v1.7.1 => github.com/johannww/fabric-gateway v0.0.0-20250520231227-838140d386e0
