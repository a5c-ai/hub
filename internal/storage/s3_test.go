package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewS3Backend_MissingConfig verifies required configuration fields.
func TestNewS3Backend_MissingConfig(t *testing.T) {
	cases := []struct {
		name   string
		cfg    S3Config
		errMsg string
	}{
		{"missing bucket", S3Config{Region: "us-east-1", AccessKeyID: "id", SecretAccessKey: "key"}, "s3 bucket name is required"},
		{"missing access key", S3Config{Region: "us-east-1", Bucket: "b", SecretAccessKey: "key"}, "s3 access key ID is required"},
		{"missing secret key", S3Config{Region: "us-east-1", Bucket: "b", AccessKeyID: "id"}, "s3 secret access key is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := NewS3Backend(tc.cfg)
			require.Error(t, err)
			assert.Nil(t, b)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

// TestS3Backend_BasicOperations runs Upload, Exists, Download, GetSize, GetLastModified, Delete and GetURL against a mock S3 HTTP server.
func TestS3Backend_BasicOperations(t *testing.T) {
	const bucket = "test-bucket"
	store := make(map[string]struct {
		data    []byte
		modTime time.Time
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 || parts[0] != bucket {
			http.NotFound(w, r)
			return
		}
		key := parts[1]
		switch r.Method {
		case http.MethodPut:
			data, _ := io.ReadAll(r.Body)
			store[key] = struct {
				data    []byte
				modTime time.Time
			}{data: data, modTime: time.Now().UTC()}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			obj, ok := store[key]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Last-Modified", obj.modTime.Format(http.TimeFormat))
			w.Write(obj.data)
		case http.MethodHead:
			obj, ok := store[key]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprint(len(obj.data)))
			w.Header().Set("Last-Modified", obj.modTime.Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			delete(store, key)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	cfg := S3Config{
		Region:          "us-east-1",
		Bucket:          bucket,
		AccessKeyID:     "id",
		SecretAccessKey: "key",
		EndpointURL:     srv.URL,
		UseSSL:          false,
	}
	backend, err := NewS3Backend(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	key := "folder/test.txt"
	content := "hello world"

	// Upload
	err = backend.Upload(ctx, key, strings.NewReader(content), int64(len(content)))
	require.NoError(t, err)

	// Exists
	exists, err := backend.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Download
	rc, err := backend.Download(ctx, key)
	require.NoError(t, err)
	data, err := io.ReadAll(rc)
	rc.Close()
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	// GetSize and GetLastModified
	size, err := backend.GetSize(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
	modTime, err := backend.GetLastModified(ctx, key)
	require.NoError(t, err)
	assert.WithinDuration(t, store[key].modTime, modTime, time.Second)

	// GetURL (presign)
	urlstr, err := backend.GetURL(ctx, key, time.Minute)
	require.NoError(t, err)
	assert.Contains(t, urlstr, key)
	assert.Contains(t, urlstr, "X-Amz-Expires=")

	// Delete and ensure removal
	err = backend.Delete(ctx, key)
	require.NoError(t, err)
	exists, err = backend.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)
}
