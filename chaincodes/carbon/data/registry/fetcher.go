package registry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
)

// FetchVerifiedData performs an HTTPS GET request to the provider's URL
// using the RootCA stored on the ledger to verify the server's identity.
func FetchVerifiedData(
	stub shim.ChaincodeStubInterface,
	providerName string,
	endpointPath string,
) ([]byte, error) {
	provider := &RegistryProvider{}
	if err := provider.FromWorldState(stub, []string{providerName}); err != nil {
		return nil, fmt.Errorf("could not find trusted provider %s: %v", providerName, err)
	}

	if provider.RootCA == nil {
		return nil, fmt.Errorf("provider %s has no RootCA registered", providerName)
	}

	// Setup TLS config with the pinned RootCA
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(provider.RootCA) {
		return nil, fmt.Errorf("failed to parse provider's RootCA PEM")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
	}

	url := fmt.Sprintf("%s%s", provider.BaseURL, endpointPath)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s failed with status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
