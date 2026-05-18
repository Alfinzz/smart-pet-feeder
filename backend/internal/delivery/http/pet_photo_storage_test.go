package http

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectImageExtension(t *testing.T) {
	tests := []struct {
		name      string
		content   []byte
		expect    string
		expectErr string
	}{
		{name: "jpeg", content: []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01}, expect: ".jpg"},
		{name: "png", content: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, expect: ".png"},
		{name: "webp", content: []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'}, expect: ".webp"},
		{name: "empty", content: nil, expectErr: "photo file is empty"},
		{name: "invalid", content: []byte("not an image"), expectErr: "photo must be a jpeg, png, or webp image"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := detectImageExtension(bytes.NewReader(test.content))
			if test.expectErr != "" {
				if err == nil || err.Error() != test.expectErr {
					t.Fatalf("detectImageExtension() error = %v, want %q", err, test.expectErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("detectImageExtension() unexpected error = %v", err)
			}
			if got != test.expect {
				t.Fatalf("detectImageExtension() = %q, want %q", got, test.expect)
			}
		})
	}
}

func TestRequestBodyTooLarge(t *testing.T) {
	if !requestBodyTooLarge(errors.New("http: request body too large")) {
		t.Fatal("requestBodyTooLarge() should detect MaxBytesReader errors")
	}
	if requestBodyTooLarge(errors.New("missing form file")) {
		t.Fatal("requestBodyTooLarge() should ignore unrelated errors")
	}
}

func TestLocalUploadPath(t *testing.T) {
	uploadDir := t.TempDir()

	got, ok := localUploadPath(uploadDir, "/uploads/pets/photo.jpg")
	if !ok {
		t.Fatal("localUploadPath() rejected valid pet upload path")
	}
	wantSuffix := filepath.Join("pets", "photo.jpg")
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("localUploadPath() = %q, want suffix %q", got, wantSuffix)
	}

	rejected := []string{
		"",
		"/uploads/other/photo.jpg",
		"/uploads/pets/../other.jpg",
		"https://example.com/uploads/pets/photo.jpg",
	}
	for _, publicPath := range rejected {
		t.Run(publicPath, func(t *testing.T) {
			if got, ok := localUploadPath(uploadDir, publicPath); ok {
				t.Fatalf("localUploadPath() accepted %q as %q", publicPath, got)
			}
		})
	}
}
