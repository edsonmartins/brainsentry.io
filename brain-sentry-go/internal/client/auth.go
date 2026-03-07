package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/integraltech/brainsentry/internal/dto"
)

// Login authenticates with email and password, stores the token.
func (c *Client) Login(email, password string) (*dto.LoginResponse, error) {
	req := dto.LoginRequest{
		Email:    email,
		Password: password,
	}
	var resp dto.LoginResponse
	if err := c.Post("/v1/auth/login", req, &resp); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	c.token = resp.Token
	return &resp, nil
}

// LoginDemo authenticates using the demo endpoint.
func (c *Client) LoginDemo() (*dto.LoginResponse, error) {
	var resp dto.LoginResponse
	if err := c.Post("/v1/auth/demo", nil, &resp); err != nil {
		return nil, fmt.Errorf("demo login: %w", err)
	}
	c.token = resp.Token
	return &resp, nil
}

// SaveToken persists the current token to a file.
func (c *Client) SaveToken(path string) error {
	if c.token == "" {
		return fmt.Errorf("no token to save")
	}
	path = expandHome(path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}
	return os.WriteFile(path, []byte(c.token), 0600)
}

// LoadToken reads a token from file and sets it on the client.
func (c *Client) LoadToken(path string) error {
	path = expandHome(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading token file: %w", err)
	}
	c.token = strings.TrimSpace(string(data))
	return nil
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
