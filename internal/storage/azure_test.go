package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAzureBackend_MissingConfig(t *testing.T) {
	cases := []struct {
		name   string
		cfg    AzureConfig
		errMsg string
	}{
		{
			name:   "missing account name",
			cfg:    AzureConfig{AccountName: ""},
			errMsg: "azure account name is required",
		},
		{
			name:   "missing account key",
			cfg:    AzureConfig{AccountName: "acct", AccountKey: ""},
			errMsg: "azure account key is required",
		},
		{
			name:   "missing container name",
			cfg:    AzureConfig{AccountName: "acct", AccountKey: "key", ContainerName: ""},
			errMsg: "azure container name is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := NewAzureBackend(tc.cfg)
			require.Error(t, err)
			assert.Nil(t, b)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}
