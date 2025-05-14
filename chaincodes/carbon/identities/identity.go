package identities

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// TODO: add attributes
type Identity interface {
	String() string
}

type X509Identity struct {
	CertID string `json:"certID"`
}

func (x509identity *X509Identity) String() string {
	return x509identity.CertID
}

func GetID(stub shim.ChaincodeStubInterface) string {
	id, err := cid.GetID(stub)
	if err != nil {
		// idemix identity
		creator, _ := stub.GetCreator()
		if creator == nil {
			panic("idemix identity creator is nil")
		}

		sid := &msp.SerializedIdentity{}
		err := proto.Unmarshal(creator, sid)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal SerializedIdentity: %s", err))
		}
		idemixID := &msp.SerializedIdemixIdentity{}
		err = proto.Unmarshal(sid.IdBytes, idemixID)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal SerializedIdemixIdentity: %s", err))
		}

		return fmt.Sprintf("%x%x", idemixID.NymX[0:8], idemixID.NymY[8:16])
	}

	return id
}
