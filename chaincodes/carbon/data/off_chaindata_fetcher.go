package data

import (
	"io"
	"net/http"
)

// GetBytesFromUri performs an HTTP request with the given method to the given URI
// and returns the response body bytes or an error if the request fails.
func GetBytesFromUri(uri, method string) ([]byte, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
