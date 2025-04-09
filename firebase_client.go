package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

type FirebaseClient struct {
	HTTPClient *http.Client
}

func (c *FirebaseClient) DoRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
	}

	return c.HTTPClient.Do(req)
}

func NewFirebaseClient(credentials string) (*FirebaseClient, error) {
	opts := []option.ClientOption{
		option.WithScopes(
			"https://www.googleapis.com/auth/firebase",
			"https://www.googleapis.com/auth/firebase.storage",
			"https://www.googleapis.com/auth/cloud-platform",
		),
	}

	if credentials != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentials)))
	}

	client, _, err := transport.NewHTTPClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return &FirebaseClient{HTTPClient: client}, nil
}
