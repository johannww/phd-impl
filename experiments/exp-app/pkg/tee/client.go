package tee

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/johannww/phd-impl/tee_auction/go/auction"
)

// Client handles communication with the TEE auction service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new TEE auction client
func NewClient(baseURL string, skipTLSVerify bool) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// Ping checks if the TEE service is reachable
func (c *Client) Ping() error {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/ping", c.baseURL))
	if err != nil {
		return fmt.Errorf("failed to ping TEE service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TEE service returned status %d", resp.StatusCode)
	}

	return nil
}

// GetReport retrieves the TEE attestation report
func (c *Client) GetReport() ([]byte, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/report", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TEE service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// AuctionResponse represents the response from the TEE auction service
type AuctionResponse struct {
	Public  *auction.SerializedAuctionResultTEE `json:"public"`
	Private *auction.SerializedAuctionResultTEE `json:"private"`
}

// RunAuction sends auction data to the TEE service and retrieves results
func (c *Client) RunAuction(auctionDataBytes []byte) (*AuctionResponse, error) {
	url := fmt.Sprintf("%s/auction", c.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewReader(auctionDataBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send auction request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TEE service returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var auctionResp AuctionResponse
	if err := json.Unmarshal(body, &auctionResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &auctionResp, nil
}

// SerializeAuctionResults converts the TEE auction results to the format expected by the chaincode
func SerializeAuctionResults(result *auction.SerializedAuctionResultTEE) ([]byte, error) {
	return json.Marshal(result)
}
