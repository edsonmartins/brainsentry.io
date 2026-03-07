package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the HTTP SDK client for the BrainSentry API.
type Client struct {
	baseURL    string
	tenantID   string
	token      string
	httpClient *http.Client
}

// New creates a new Client with the given base URL and tenant ID.
func New(baseURL, tenantID string) *Client {
	return &Client{
		baseURL:  baseURL,
		tenantID: tenantID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the JWT token for authenticated requests.
func (c *Client) SetToken(token string) {
	c.token = token
}

// Token returns the current JWT token.
func (c *Client) Token() string {
	return c.token
}

// IsAuthenticated returns true if the client has a token.
func (c *Client) IsAuthenticated() bool {
	return c.token != ""
}

// Get performs a GET request and decodes the JSON response into result.
func (c *Client) Get(path string, result any) error {
	return c.do(http.MethodGet, path, nil, result)
}

// Post performs a POST request with a JSON body and decodes the response.
func (c *Client) Post(path string, body, result any) error {
	return c.do(http.MethodPost, path, body, result)
}

// Put performs a PUT request with a JSON body and decodes the response.
func (c *Client) Put(path string, body, result any) error {
	return c.do(http.MethodPut, path, body, result)
}

// Patch performs a PATCH request with a JSON body and decodes the response.
func (c *Client) Patch(path string, body, result any) error {
	return c.do(http.MethodPatch, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	return c.do(http.MethodDelete, path, nil, nil)
}

// APIError represents an error returned by the API.
type APIError struct {
	StatusCode int
	Message    string
	ErrorCode  string
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" {
		return fmt.Sprintf("API error %d [%s]: %s", e.StatusCode, e.ErrorCode, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

func (c *Client) do(method, path string, body, result any) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", c.tenantID)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			break
		}
		resp.Body.Close()
		time.Sleep(time.Duration(attempt+1) * time.Second)

		// Recreate request body for retry
		if body != nil {
			data, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(data)
		}
		req, _ = http.NewRequest(method, url, bodyReader)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", c.tenantID)
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		apiErr := &APIError{StatusCode: resp.StatusCode}

		var errResp struct {
			Error     string `json:"error"`
			Message   string `json:"message"`
			ErrorCode string `json:"errorCode"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			apiErr.Message = errResp.Message
			if apiErr.Message == "" {
				apiErr.Message = errResp.Error
			}
			apiErr.ErrorCode = errResp.ErrorCode
		} else {
			apiErr.Message = string(respBody)
		}
		return apiErr
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}
