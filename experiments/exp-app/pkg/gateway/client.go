package gateway

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	pb_msp "github.com/hyperledger/fabric-protos-go-apiv2/msp"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/johannww/phd-impl/chaincodes/common/state/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GatewayConfig struct {
	PeerAddr      string // e.g., localhost:7051
	TLSCertPath   string // Path to peer TLS cert
	MspID         string // e.g., Org1MSP
	UserCertPath  string // Path to user certificate
	UserKeyPath   string // Path to user private key
	ChannelName   string
	ChaincodeName string
}

type ClientWrapper struct {
	gateway  *client.Gateway
	network  *client.Network
	contract *client.Contract
	config   *GatewayConfig
}

func NewClientWrapper(cfg *GatewayConfig) (*ClientWrapper, error) {
	// Load TLS certificate for peer
	tlsCertPEM, err := os.ReadFile(cfg.TLSCertPath)
	if err != nil {
		return nil, fmt.Errorf("read tls cert: %w", err)
	}

	tlsCert, err := identity.CertificateFromPEM(tlsCertPEM)
	if err != nil {
		return nil, fmt.Errorf("parse tls cert: %w", err)
	}

	// Create TLS credentials
	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)
	tlsConfig := credentials.NewClientTLSFromCert(certPool, "")

	// Create gRPC connection
	conn, err := grpc.NewClient(cfg.PeerAddr, grpc.WithTransportCredentials(tlsConfig))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	// Load user identity
	userID, err := newIdentity(cfg.MspID, cfg.UserCertPath)
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	// Load signing function
	signFunc, err := newSign(cfg.UserKeyPath)
	if err != nil {
		return nil, fmt.Errorf("create sign func: %w", err)
	}

	// Create gateway
	gw, err := client.Connect(
		userID,
		client.WithSign(signFunc),
		client.WithHash(hash.SHA256),
		client.WithClientConnection(conn),
	)
	if err != nil {
		return nil, fmt.Errorf("connect gateway: %w", err)
	}

	// Get network and contract
	network := gw.GetNetwork(cfg.ChannelName)
	contract := network.GetContract(cfg.ChaincodeName)

	return &ClientWrapper{
		gateway:  gw,
		network:  network,
		contract: contract,
		config:   cfg,
	}, nil
}

func (c *ClientWrapper) attachErrorInfo(res []byte, err error) (result []byte, finalErr error) {
	if err == nil {
		return res, nil
	}

	switch e := err.(type) {
	case *client.EndorseError:
		return res, fmt.Errorf("endorse error: %s. details: %s", e.Error(), e.GRPCStatus().Details())
	case *client.SubmitError:
		return res, fmt.Errorf("submit error: %s. details: %s", e.Error(), e.GRPCStatus().Details())
	case *client.TransactionError:
		return res, fmt.Errorf("transaction error: %s. details: %s", e.Error(), e.GRPCStatus().Details())
	case *client.CommitStatusError:
		return res, fmt.Errorf("commit status error: %s. details: %s", e.Error(), e.GRPCStatus().Details())
	case *client.CommitError:
		return res, fmt.Errorf("commit error: %s", e.Error())
	default:
		return res, fmt.Errorf("unexpected error type: %T: %w", err, err)
	}

}

func (c *ClientWrapper) GetIdentityID() string {
	// Create a SerializedIdentity that matches what the peer/orderer expects
	sid := &pb_msp.SerializedIdentity{
		Mspid:   c.gateway.Identity().MspID(),
		IdBytes: c.gateway.Identity().Credentials(),
	}
	sidBytes, _ := proto.Marshal(sid)

	// Create a mock stub and set the creator to our serialized identity
	mockStub := mocks.NewMockStub("identity-calc", nil)
	mockStub.Creator = sidBytes

	// Use cid.New with the mock stub as requested by the user
	cidObj, _ := cid.New(mockStub)
	id, _ := cidObj.GetID()
	return id
}

func (c *ClientWrapper) EvaluateTransaction(functionName string, args ...string) ([]byte, error) {
	res, err := c.contract.EvaluateTransaction(functionName, args...)
	return c.attachErrorInfo(res, err)
}

func (c *ClientWrapper) SubmitTransaction(functionName string, args ...string) ([]byte, error) {
	res, err := c.contract.SubmitTransaction(functionName, args...)
	return c.attachErrorInfo(res, err)
}

func (c *ClientWrapper) SubmitAsync(functionName string, args ...string) ([]byte, *client.Commit, error) {
	res, commit, err := c.contract.SubmitAsync(functionName, client.WithArguments(args...))
	res, err = c.attachErrorInfo(res, err)
	return res, commit, err
}

func (c *ClientWrapper) SubmitWithTransient(functionName string, transient map[string][]byte, args ...string) ([]byte, error) {
	proposal, err := c.contract.NewProposal(functionName, client.WithArguments(args...), client.WithTransient(transient))
	if err != nil {
		return nil, fmt.Errorf("create proposal: %w", err)
	}

	transaction, err := proposal.Endorse()
	if err != nil {
		return c.attachErrorInfo(nil, err)
	}

	result := transaction.Result()
	commit, err := transaction.Submit()
	if err != nil {
		return c.attachErrorInfo(result, err)
	}

	commitStatus, err := commit.Status()
	if err != nil {
		return c.attachErrorInfo(result, err)
	}

	if commitStatus.Code != peer.TxValidationCode_VALID {
		return c.attachErrorInfo(result, fmt.Errorf("transaction commit failed with status code: %s", commitStatus.Code.String()))
	}

	return result, nil
}

func (c *ClientWrapper) SubmitAsyncWithTransient(
	functionName string,
	transient map[string][]byte,
	args ...string,
) ([]byte, *client.Commit, error) {
	res, commit, err := c.contract.SubmitAsync(
		functionName,
		client.WithArguments(args...),
		client.WithTransient(transient),
	)
	res, err = c.attachErrorInfo(res, err)
	return res, commit, err
}

func (c *ClientWrapper) Close() error {
	return c.gateway.Close()
}

func newIdentity(mspID, certPath string) (*identity.X509Identity, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read cert: %w", err)
	}

	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, fmt.Errorf("parse cert: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}

	return id, nil
}

func newSign(keyPath string) (identity.Sign, error) {
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}

	key, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	sign, err := identity.NewPrivateKeySign(key)
	if err != nil {
		return nil, fmt.Errorf("create sign: %w", err)
	}

	return sign, nil
}
