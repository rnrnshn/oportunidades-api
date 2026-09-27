package uploads

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/rnrnshn/oportunidades-api/pkg/storage"
)

var (
	ErrForbidden          = errors.New("uploads: forbidden")
	ErrUnsupportedType    = errors.New("uploads: unsupported content type")
	ErrInvalidPath        = errors.New("uploads: invalid path")
	filenameUnsafePattern = regexp.MustCompile(`[^a-z0-9._-]+`)
)

// SVG is excluded on purpose: it can carry scripts.
var extensionByContentType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

var contentFolders = map[string]bool{"articles": true, "opportunities": true, "universities": true}

const usersFolder = "users"

type Actor struct {
	ID   string
	Role string
}

type Service struct {
	storageClient storage.Client
}

type PresignInput struct {
	Filename    string
	ContentType string
	Folder      string
	Actor       Actor
}

type ConfirmInput struct {
	Path  string
	Actor Actor
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
	ext, ok := extensionByContentType[strings.ToLower(strings.TrimSpace(input.ContentType))]
	if !ok {
		return nil, ErrUnsupportedType
	}
	prefix, err := folderPrefix(input.Folder, input.Actor)
	if err != nil {
		return nil, err
	}
	signedUpload, err := s.storageClient.CreateSignedUpload(ctx, prefix+uuid.NewString()+"-"+safeName(input.Filename)+ext)
	if err != nil {
		return nil, fmt.Errorf("uploads: create signed upload: %w", err)
	}
	return &PresignResult{Data: *signedUpload}, nil
}

func (s *Service) Confirm(_ context.Context, input ConfirmInput) (*ConfirmResult, error) {
	objectPath := strings.TrimSpace(input.Path)
	if objectPath == "" || strings.HasPrefix(objectPath, "/") || path.Clean(objectPath) != objectPath {
		return nil, ErrInvalidPath
	}
	folder, _, found := strings.Cut(objectPath, "/")
	if !found {
		return nil, ErrInvalidPath
	}
	prefix, err := folderPrefix(folder, input.Actor)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(objectPath, prefix) || !allowedExtension(filepath.Ext(objectPath)) {
		return nil, ErrInvalidPath
	}

	result := &ConfirmResult{}
	result.Data.Path = objectPath
	result.Data.PublicURL = s.storageClient.PublicURL(objectPath)
	return result, nil
}

func folderPrefix(folder string, actor Actor) (string, error) {
	folder = strings.TrimSpace(folder)
	switch {
	case contentFolders[folder]:
		if actor.Role != "admin" && actor.Role != "cms_partner" {
			return "", ErrForbidden
		}
		return folder + "/", nil
	case folder == usersFolder:
		if strings.TrimSpace(actor.ID) == "" {
			return "", ErrForbidden
		}
		return usersFolder + "/" + actor.ID + "/", nil
	default:
		return "", ErrInvalidPath
	}
}

func allowedExtension(ext string) bool {
	for _, allowed := range extensionByContentType {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}

func safeName(filename string) string {
	base := strings.ToLower(strings.TrimSpace(filepath.Base(filename)))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.Trim(filenameUnsafePattern.ReplaceAllString(base, "-"), "-.")
	if base == "" {
		return "upload"
	}
	return base
}
