package companies

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/johannww/phd-impl/chaincodes/carbon/data"
	"github.com/johannww/phd-impl/chaincodes/common/pb"
	"github.com/johannww/phd-impl/chaincodes/common/state"
	"github.com/johannww/phd-impl/chaincodes/common/utils"
	"google.golang.org/protobuf/proto"
)

const COMPANY_PREFIX = "company"

// Company represent a company stored as public data in the world state.
type Company struct {
	ID         string                // ID might be the CNPJ (brazilian company national ID)
	Coordinate *utils.Coordinate     // Geographical coordinate in floating point format
	DataProps  *data.ValidationProps // How data from the company is validated
}

var _ state.WorldStateManager = (*Company)(nil)

// FromWorldState implements state.WorldStateManager.
func (c *Company) FromWorldState(stub shim.ChaincodeStubInterface, keyAttributes []string) error {
	return state.GetStateWithCompositeKey(stub, COMPANY_PREFIX, keyAttributes, c)
}

// GetID implements state.WorldStateManager.
func (c *Company) GetID() *[][]string {
	return &[][]string{{c.ID}}
}

// ToWorldState implements state.WorldStateManager.
func (c *Company) ToWorldState(stub shim.ChaincodeStubInterface) error {
	id := c.GetID()
	return state.PutStateWithCompositeKey(stub, COMPANY_PREFIX, id, c)
}

func RegisterCompany(stub shim.ChaincodeStubInterface, company *Company) error {
	if company == nil {
		return nil
	}
	if company.ID == "" {
		return fmt.Errorf("company ID cannot be empty")
	}
	if *company.Coordinate == (utils.Coordinate{}) {
		return fmt.Errorf("company coordinate cannot be empty")
	}
	if company.DataProps == nil || len(company.DataProps.Methods) == 0 {
		return fmt.Errorf("company data properties cannot be empty")
	}

	return company.ToWorldState(stub)
}

func (c *Company) ToProto() proto.Message {
	var coord *pb.Coordinate
	if c.Coordinate != nil {
		coord = &pb.Coordinate{Latitude: c.Coordinate.Latitude, Longitude: c.Coordinate.Longitude}
	}
	var valProps *pb.ValidationProps
	if c.DataProps != nil {
		methods := make([]pb.ValidationMethod, 0, len(c.DataProps.Methods))
		for _, m := range c.DataProps.Methods {
			methods = append(methods, pb.ValidationMethod(m))
		}
		valProps = &pb.ValidationProps{Methods: methods}
	}
	return &pb.Company{Id: c.ID, Coordinate: coord, DataProps: valProps}
}

func (c *Company) FromProto(m proto.Message) error {
	pc, ok := m.(*pb.Company)
	if !ok {
		return fmt.Errorf("unexpected proto message type for Company")
	}
	c.ID = pc.Id
	if pc.Coordinate != nil {
		c.Coordinate = &utils.Coordinate{Latitude: pc.Coordinate.Latitude, Longitude: pc.Coordinate.Longitude}
	} else {
		c.Coordinate = nil
	}
	if pc.DataProps != nil {
		methods := make([]data.ValidationMethod, 0, len(pc.DataProps.Methods))
		for _, m := range pc.DataProps.Methods {
			methods = append(methods, data.ValidationMethod(m))
		}
		c.DataProps = &data.ValidationProps{Methods: methods}
	} else {
		c.DataProps = nil
	}
	return nil
}
