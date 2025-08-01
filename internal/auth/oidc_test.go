package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvisionUser_RoleAssignment(t *testing.T) {
	// Setup services and database
	authSvc, db, cfg := setupTestServices(t)
	jwtMgr := NewJWTManager(cfg.JWT)
	oidcSvc, err := NewOIDCService(db, jwtMgr, cfg, authSvc)
	require.NoError(t, err)

	// Test extracting claims.Roles and mapping groups to roles
	claims := &OIDCClaims{
		Email:             "user@example.com",
		PreferredUsername: "user",
		Roles:             []string{"r1", "r2"},
		Groups:            []string{"g1", "g2"},
	}
	j := &JITProvisioningConfig{
		Enabled:     true,
		DefaultRole: "default",
		GroupMapping: map[string]string{
			"g2": "mapped",
		},
	}
	user, err := oidcSvc.ProvisionUser(claims, j)
	require.NoError(t, err)
	// expect claims roles and the mapped group role, default role not applied since roles exist
	assert.ElementsMatch(t, []string{"r1", "r2", "mapped"}, user.Roles)

	// Test default role is applied when no roles or mappings present
	claims2 := &OIDCClaims{Email: "u2@example.com", PreferredUsername: "u2"}
	j2 := &JITProvisioningConfig{Enabled: true, DefaultRole: "fallback"}
	user2, err := oidcSvc.ProvisionUser(claims2, j2)
	require.NoError(t, err)
	assert.Equal(t, []string{"fallback"}, user2.Roles)
}
