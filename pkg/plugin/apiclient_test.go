package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIClient(t *testing.T) {
	token := "test-token"
	client := NewAPIClient(token)
	assert.NotNil(t, client)
	assert.Equal(t, token, client.token)
}

func TestTokenHeader(t *testing.T) {
	client := NewAPIClient("abc123")
	expected := "Bearer abc123"
	assert.Equal(t, expected, client.tokenHeader())
}
