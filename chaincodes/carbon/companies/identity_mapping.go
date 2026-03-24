package companies

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"google.golang.org/protobuf/proto"
)

const PSEUDONYM_TO_COMPANY_ID_PREFIX = "pseudonymToCompanyID"

type PseudonymToCompanyID struct {
	Pseudonym string `json:"pseudonym"` // Pseudonym of the company (provided by identities.GetID(stub))
	CompanyID string `json:"companyID"` // ID of the company, e.g., CNPJ in Brazil
}

var _ state.WorldStateManager = (*PseudonymToCompanyID)(nil)

func (p *PseudonymToCompanyID) ToProto() proto.Message {
	return &pb.PseudonymToCompanyID{
		Pseudonym: p.Pseudonym,
		CompanyID: p.CompanyID,
	}
}

func (p *PseudonymToCompanyID) FromProto(m proto.Message) error {
	pp, ok := m.(*pb.PseudonymToCompanyID)
	if !ok {
		return fmt.Errorf("unexpected proto message type for PseudonymToCompanyID")
	}
	p.Pseudonym = pp.Pseudonym
	p.CompanyID = pp.CompanyID
	return nil
}

// FromWorldState implements state.WorldStateManager.
func (p *PseudonymToCompanyID) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetPvtDataWithCompositeKey(stub, PSEUDONYM_TO_COMPANY_ID_PREFIX, keyAttributes, state.COMPANIES_PVT_DATA_COLLECTION, p)
}

// GetID implements state.WorldStateManager.
func (p *PseudonymToCompanyID) GetID() *[][]string {
	return &[][]string{{p.Pseudonym}}
}

// ToWorldState implements state.WorldStateManager.
func (p *PseudonymToCompanyID) ToWorldState(stub shim.ChaincodeStubInterface) error {
	firstID := (*p.GetID())[0]
	return state.PutPvtDataWithCompositeKey(stub, PSEUDONYM_TO_COMPANY_ID_PREFIX, firstID, state.COMPANIES_PVT_DATA_COLLECTION, p)
}

// CreatePseudonymToCompanyID creates a new mapping and saves it to the world state.
func CreatePseudonymToCompanyID(stub shim.ChaincodeStubInterface, pseudonym, companyID string) error {
	pseudonymToCompanyID := &PseudonymToCompanyID{
		Pseudonym: pseudonym,
		CompanyID: companyID,
	}
	return pseudonymToCompanyID.ToWorldState(stub)
}

// GetCompanyIDByPseudonym returns the company ID for a given pseudonym.
func GetCompanyIDByPseudonym(stub shim.ChaincodeStubInterface, pseudonym string) (string, error) {
	pseudonymToCompanyID := &PseudonymToCompanyID{}
	err := pseudonymToCompanyID.FromWorldState(stub, []string{pseudonym})
	if err != nil {
		return "", err
	}
	return pseudonymToCompanyID.CompanyID, nil
}
