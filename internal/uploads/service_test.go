package uploads

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/rnrnshn/oportunidades-api/pkg/storage"
)

type recordingStorageClient struct {
	objectPath string
}

func (r *recordingStorageClient) CreateSignedUpload(_ context.Context, objectPath string) (*storage.SignedUpload, error) {
	r.objectPath = objectPath
	return &storage.SignedUpload{Path: objectPath}, nil
}

func (r *recordingStorageClient) PublicURL(objectPath string) string {
	return "https://cdn.example/" + objectPath
}

var (
	editor  = Actor{ID: "11111111-1111-1111-1111-111111111111", Role: "cms_partner"}
	student = Actor{ID: "22222222-2222-2222-2222-222222222222", Role: "user"}
)

func TestPresignBuildsPathFromContentType(t *testing.T) {
	client := &recordingStorageClient{}
	service := NewService(client)

	if _, err := service.Presign(context.Background(), PresignInput{Filename: "../Logo UEM.svg", ContentType: "image/png", Folder: "universities", Actor: editor}); err != nil {
		t.Fatalf("presign returned error: %v", err)
	}
	if !strings.HasPrefix(client.objectPath, "universities/") || !strings.HasSuffix(client.objectPath, "-logo-uem.png") || strings.Contains(client.objectPath, "..") {
		t.Fatalf("unexpected object path: %s", client.objectPath)
	}
}

func TestPresignUserFolderIsScopedToCaller(t *testing.T) {
	client := &recordingStorageClient{}
	service := NewService(client)

	if _, err := service.Presign(context.Background(), PresignInput{Filename: "me.jpg", ContentType: "image/jpeg", Folder: "users", Actor: student}); err != nil {
		t.Fatalf("presign returned error: %v", err)
	}
	if !strings.HasPrefix(client.objectPath, "users/"+student.ID+"/") {
		t.Fatalf("expected path scoped to the caller, got %s", client.objectPath)
	}
}

func TestPresignRejects(t *testing.T) {
	cases := map[string]struct {
		input PresignInput
		want  error
	}{
		"svg":                    {PresignInput{Filename: "x.svg", ContentType: "image/svg+xml", Folder: "articles", Actor: editor}, ErrUnsupportedType},
		"html":                   {PresignInput{Filename: "x.html", ContentType: "text/html", Folder: "articles", Actor: editor}, ErrUnsupportedType},
		"student content folder": {PresignInput{Filename: "x.png", ContentType: "image/png", Folder: "articles", Actor: student}, ErrForbidden},
		"anonymous users folder": {PresignInput{Filename: "x.png", ContentType: "image/png", Folder: "users", Actor: Actor{}}, ErrForbidden},
		"unknown folder":         {PresignInput{Filename: "x.png", ContentType: "image/png", Folder: "secrets", Actor: editor}, ErrInvalidPath},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewService(&recordingStorageClient{}).Presign(context.Background(), tc.input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestConfirm(t *testing.T) {
	cases := map[string]struct {
		path  string
		actor Actor
		want  error
	}{
		"editor content":       {"articles/abc-cover.webp", editor, nil},
		"own user folder":      {"users/" + student.ID + "/abc-me.jpg", student, nil},
		"other user folder":    {"users/" + editor.ID + "/abc-me.jpg", student, ErrInvalidPath},
		"student content":      {"articles/abc-cover.png", student, ErrForbidden},
		"traversal":            {"articles/../users/x.png", editor, ErrInvalidPath},
		"leading slash":        {"/articles/abc.png", editor, ErrInvalidPath},
		"disallowed extension": {"articles/abc.svg", editor, ErrInvalidPath},
		"no folder":            {"abc.png", editor, ErrInvalidPath},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, err := NewService(&recordingStorageClient{}).Confirm(context.Background(), ConfirmInput{Path: tc.path, Actor: tc.actor})
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
			if tc.want == nil && result.Data.PublicURL != "https://cdn.example/"+tc.path {
				t.Fatalf("unexpected public url: %s", result.Data.PublicURL)
			}
		})
	}
}
