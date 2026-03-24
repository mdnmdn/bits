package bitget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/coingecko/coingecko-cli/internal/auth"
)

// generateSignature generates HMAC-SHA256 signature for Bitget API.
// The message format is: timestamp + method + path + body.
func (c *Client) generateSignature(method, path, body, timestamp string) string {
	message := timestamp + method + path + body
	return auth.GenerateHMACSHA256Base64(message, c.config.Secret)
}

// signedRequest creates and executes a signed HTTP request to the Bitget API.
// It handles signature generation and sets all required authentication headers.
func (c *Client) signedRequest(method, endpoint string, queryString string, jsonBody interface{}) ([]byte, error) {
	var bodyStr string
	var bodyReader io.Reader

	if jsonBody != nil {
		jsonData, err := json.Marshal(jsonBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyStr = string(jsonData)
		bodyReader = bytes.NewBuffer(jsonData)
	}

	path := endpoint
	if queryString != "" {
		path = endpoint + "?" + queryString
	}

	url := c.config.BaseURL + path

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign := c.generateSignature(method, path, bodyStr, timestamp)

	req.Header.Set("ACCESS-KEY", c.config.Key)
	req.Header.Set("ACCESS-SIGN", sign)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", c.config.Passphrase)
	req.Header.Set("Content-Type", "application/json")

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}
