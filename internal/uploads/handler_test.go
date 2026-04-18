package uploads

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/storage"
)

type stubStorageClient struct{}

func (stubStorageClient) CreateSignedUpload(context.Context, string) (*storage.SignedUpload, error) {
	return &storage.SignedUpload{Path: "articles/test.png", UploadURL: "https://example.com/upload", PublicURL: "https://example.com/public/test.png"}, nil
}

func (stubStorageClient) PublicURL(objectPath string) string {
	return "https://example.com/public/" + objectPath
}

func TestHandlerPresignValidatesFolder(t *testing.T) {
	handler := NewHandler(NewService(stubStorageClient{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/uploads/presign", handler.Presign)
	req := httptest.NewRequest(http.MethodPost, "/v1/uploads/presign", strings.NewReader(`{"filename":"logo.png","content_type":"image/png","folder":"bad"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerConfirmRequiresPath(t *testing.T) {
	handler := NewHandler(NewService(stubStorageClient{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/uploads/confirm", handler.Confirm)
	req := httptest.NewRequest(http.MethodPost, "/v1/uploads/confirm", strings.NewReader(`{"path":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
