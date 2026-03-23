// internal/api/client.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/crewsbase/crewsbase-cli/internal/config"
)

// Version is set by the cmd package at init time
var Version = "dev"

// APIError represents an API error with an exit code
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

type Client struct {
	baseURL    string
	token      string
	account    string
	httpClient *http.Client
	debug      bool
}

func NewClient(accountOverride, tokenOverride string, debug bool) (*Client, error) {
	token := config.GetToken(tokenOverride)
	if token == "" {
		return nil, fmt.Errorf("not authenticated. Run `crewsbase auth login` to authenticate")
	}

	account := config.GetAccount(accountOverride)
	if account == "" {
		return nil, fmt.Errorf("no account set. Run `crewsbase config set default_account <slug>` or use --account flag")
	}

	return &Client{
		baseURL:    config.GetAPIURL(),
		token:      token,
		account:    account,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		debug:      debug,
	}, nil
}

func (c *Client) Get(path string, params map[string]string) ([]byte, http.Header, error) {
	return c.do("GET", path, params, nil)
}

func (c *Client) Post(path string, body interface{}) ([]byte, http.Header, error) {
	return c.do("POST", path, nil, body)
}

func (c *Client) Patch(path string, body interface{}) ([]byte, http.Header, error) {
	return c.do("PATCH", path, nil, body)
}

func (c *Client) Delete(path string) ([]byte, http.Header, error) {
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) do(method, path string, params map[string]string, body interface{}) ([]byte, http.Header, error) {
	url := fmt.Sprintf("%s/%s/api%s", c.baseURL, c.account, path)

	var jsonBody []byte
	var err error
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "crewsbase-cli/"+Version)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: %s %s\n", method, req.URL)
	}

	// Retry with backoff on 429
	var resp *http.Response
	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(delay)
			// Recreate request body for retry
			if jsonBody != nil {
				req.Body = io.NopCloser(bytes.NewReader(jsonBody))
			}
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("request failed: %w", err)
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			break
		}
		resp.Body.Close()
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, nil, &APIError{Code: 2, Message: "authentication failed. Run `crewsbase auth login` to re-authenticate"}
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil, &APIError{Code: 3, Message: "not found"}
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return respBody, resp.Header, nil
}
