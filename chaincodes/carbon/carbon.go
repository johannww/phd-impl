package carbon

import (
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	pb "github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

type Carbon struct{}

// Init is called during Instantiate transaction after the chaincode container
// has been established for the first time, allowing the chaincode to
// initialize its internal data
func (carbon *Carbon) Init(stub shim.ChaincodeStubInterface) *pb.Response {
	panic("not implemented") // TODO: Implement
}

// Invoke is called to update or query the ledger in a proposal transaction.
// Updated state variables are not committed to the ledger until the
// transaction is committed.
func (carbon *Carbon) Invoke(stub shim.ChaincodeStubInterface) *pb.Response {
	// TODO: perhaps use the contract api here
	// it uses reflect to auto parse the parameters
	return shim.Success([]byte("Carbon chaincode is returning a success response"))
	// panic("not implemented") // TODO: Implement
}
