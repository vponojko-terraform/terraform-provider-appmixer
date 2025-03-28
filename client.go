package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DoRequest executes a request with authentication
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var req *http.Request
	var err error

	url := fmt.Sprintf("%s%s", c.ApiURL, path)

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	// Set authorization header if token exists
	if c.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	}

	tflog.Debug(ctx, "Making API request", map[string]interface{}{
		"method": method,
		"url":    url,
	})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Log the error response
		tflog.Error(ctx, "API request failed", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(respBody),
			"method":      method,
			"url":         url,
		})

		// Try to parse error message if possible
		var errorMessage string
		var errorObj map[string]interface{}
		if err := json.Unmarshal(respBody, &errorObj); err == nil {
			if msg, ok := errorObj["message"]; ok {
				errorMessage = fmt.Sprintf("%v", msg)
			} else if msg, ok := errorObj["error"]; ok {
				errorMessage = fmt.Sprintf("%v", msg)
			}
		}

		if errorMessage != "" {
			return respBody, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorMessage)
		}

		return respBody, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
