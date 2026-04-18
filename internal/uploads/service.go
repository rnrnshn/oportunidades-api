package uploads

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/rnrnshn/oportunidades-api/pkg/storage"
)

var filenameUnsafePattern = regexp.MustCompile(`[^a-z0-9._-]+`)

type Service struct {
	storageClient storage.Client
}

type PresignInput struct {
	Filename    string
	ContentType string
	Folder      string
}

type ConfirmInput struct {
	Path string
}

type PresignResult struct {
	Data storage.SignedUpload `json:"data"`
}

type ConfirmResult struct {
	Data struct {
		Path      string `json:"path"`
		PublicURL string `json:"public_url"`
	} `json:"data"`
}

func NewService(storageClient storage.Client) *Service {
	return &Service{storageClient: storageClient}
}

func (s *Service) Presign(ctx context.Context, input PresignInput) (*PresignResult, error) {
	objectPath := buildObjectPath(input.Folder, input.Filename)
	signedUpload, err := s.storageClient.CreateSignedUpload(ctx, objectPath)
	if err != nil {
		return nil, fmt.Errorf("uploads: create signed upload: %w", err)
	}
	return &PresignResult{Data: *signedUpload}, nil
}

func (s *Service) Confirm(_ context.Context, input ConfirmInput) (*ConfirmResult, error) {
	result := &ConfirmResult{}
	result.Data.Path = strings.TrimLeft(strings.TrimSpace(input.Path), "/")
	result.Data.PublicURL = s.storageClient.PublicURL(result.Data.Path)
	return result, nil
}

func buildObjectPath(folder string, filename string) string {
	base := strings.ToLower(strings.TrimSpace(filename))
	base = filenameUnsafePattern.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	ext := strings.ToLower(filepath.Ext(base))
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = "upload"
	}
	return strings.Trim(strings.TrimSpace(folder), "/") + "/" + uuid.NewString() + "-" + name + ext
}
