package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIClient(t *testing.T) {
	token := "test-token"
	client := NewAPIClient(token)
	assert.NotNil(t, client)
	assert.Equal(t, token, client.token)
}

func TestGetRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		assert.Equal(t, "/api/v1/repositories/foo/bar", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer server.Close()

	os.Setenv("HUB_API_URL", server.URL+"/api/v1")
	defer os.Unsetenv("HUB_API_URL")
	client := NewAPIClient("token")
	client.client = server.Client()

	res, err := client.GetRepository(context.Background(), "foo", "bar")
	assert.NoError(t, err)
	m, ok := res.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", m["key"])
}

func TestListRepositories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		assert.Equal(t, "/api/v1/repositories", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"r1"},{"name":"r2"}]`)
	}))
	defer server.Close()

	os.Setenv("HUB_API_URL", server.URL+"/api/v1")
	defer os.Unsetenv("HUB_API_URL")
	client := NewAPIClient("tok")
	client.client = server.Client()

	list, err := client.ListRepositories(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	m0, ok := list[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "r1", m0["name"])
}

func TestCreatePRComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer abc", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/repositories/x/y/pulls/7/comments", r.URL.Path)
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		assert.Equal(t, "hi there", payload["body"])
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	os.Setenv("HUB_API_URL", server.URL+"/api/v1")
	defer os.Unsetenv("HUB_API_URL")
	client := NewAPIClient("abc")
	client.client = server.Client()

	err := client.CreatePRComment(context.Background(), "x", "y", 7, "hi there")
	assert.NoError(t, err)
}

func TestCreateStatusCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tk", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/repositories/a/b/statuses/abcd1234", r.URL.Path)
		var chk StatusCheck
		_ = json.NewDecoder(r.Body).Decode(&chk)
		assert.Equal(t, "failure", chk.State)
		assert.Equal(t, "ctx", chk.Context)
		assert.Equal(t, "desc", chk.Description)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	os.Setenv("HUB_API_URL", server.URL+"/api/v1")
	defer os.Unsetenv("HUB_API_URL")
	client := NewAPIClient("tk")
	client.client = server.Client()

	check := &StatusCheck{State: "failure", Context: "ctx", Description: "desc"}
	err := client.CreateStatusCheck(context.Background(), "a", "b", "abcd1234", check)
	assert.NoError(t, err)
}

func TestTokenHeader(t *testing.T) {
	client := NewAPIClient("abc123")
	expected := "Bearer abc123"
	assert.Equal(t, expected, client.tokenHeader())
}
