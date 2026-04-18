package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SignedUpload struct {
	Path      string `json:"path"`
	UploadURL string `json:"upload_url"`
	Token     string `json:"token,omitempty"`
	PublicURL string `json:"public_url"`
}

type Client interface {
	CreateSignedUpload(ctx context.Context, objectPath string) (*SignedUpload, error)
	PublicURL(objectPath string) string
}

type SupabaseClient struct {
	baseURL    string
	bucket     string
	serviceKey string
	publicURL  string
	httpClient *http.Client
}

type createSignedUploadResponse struct {
	SignedURL string `json:"signedURL"`
	URL       string `json:"url"`
	Token     string `json:"token"`
	Path      string `json:"path"`
	Key       string `json:"Key"`
}

func NewSupabaseClient(baseURL string, bucket string, serviceKey string) *SupabaseClient {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	trimmedBucket := strings.TrimSpace(bucket)
	return &SupabaseClient{
		baseURL:    trimmedBaseURL,
		bucket:     trimmedBucket,
		serviceKey: strings.TrimSpace(serviceKey),
		publicURL:  trimmedBaseURL + "/storage/v1/object/public/" + trimmedBucket,
		httpClient: &http.Client{},
	}
}

func (c *SupabaseClient) CreateSignedUpload(ctx context.Context, objectPath string) (*SignedUpload, error) {
	requestBody, err := json.Marshal(map[string]any{"upsert": true})
	if err != nil {
		return nil, fmt.Errorf("storage: marshal signed upload request: %w", err)
	}
	endpoint := fmt.Sprintf("%s/storage/v1/object/upload/sign/%s/%s", c.baseURL, c.bucket, strings.TrimLeft(objectPath, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("storage: create signed upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("storage: request signed upload url: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("storage: signed upload request failed with status %d", resp.StatusCode)
	}

	var payload createSignedUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("storage: decode signed upload response: %w", err)
	}
	rawSignedURL := firstNonEmpty(payload.SignedURL, payload.URL)
	if rawSignedURL == "" {
		return nil, fmt.Errorf("storage: signed upload url missing from response")
	}
	uploadURL := rawSignedURL
	if strings.HasPrefix(rawSignedURL, "/") {
		uploadURL = c.baseURL + rawSignedURL
	}
	path := firstNonEmpty(payload.Path, payload.Key, objectPath)
	return &SignedUpload{Path: strings.TrimLeft(path, "/"), UploadURL: uploadURL, Token: payload.Token, PublicURL: c.PublicURL(path)}, nil
}

func (c *SupabaseClient) PublicURL(objectPath string) string {
	return c.publicURL + "/" + strings.TrimLeft(objectPath, "/")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
