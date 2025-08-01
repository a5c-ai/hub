package auth

import (
	"strings"
	"testing"

	"github.com/a5c-ai/hub/internal/config"
)

// Test the mock LDAP connection for bind and search behavior
func TestMockLDAPConnection_BindAndSearch(t *testing.T) {
	m := NewMockLDAPConnection()
	// successful bind
	if err := m.Bind("testuser", "testpass"); err != nil {
		t.Fatalf("expected bind success, got error: %v", err)
	}
	// wrong password
	if err := m.Bind("testuser", "wrong"); err != ErrLDAPAuth {
		t.Errorf("expected ErrLDAPAuth, got %v", err)
	}
	// unknown user
	if err := m.Bind("unknown", "pass"); err != ErrLDAPUserNotFound {
		t.Errorf("expected ErrLDAPUserNotFound, got %v", err)
	}
	// search matching testuser
	results, err := m.Search("dc=example,dc=com", "testuser", []string{"cn", "mail"})
	if err != nil {
		t.Fatalf("expected search success, got error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	entry := results[0]
	if entry.DN == "" {
		t.Error("expected DN to be set")
	}
	if mail, ok := entry.Attributes["mail"]; !ok || len(mail) == 0 {
		t.Errorf("expected mail attribute, got %v", entry.Attributes)
	}
}

// Test createConnection returns an error for invalid host in production mode
func TestCreateConnection_InvalidHost(t *testing.T) {
	svc := &LDAPService{config: &config.Config{Environment: "production", LDAP: config.LDAP{Enabled: true, Host: "nonexistent.invalid", Port: 389}}}
	_, err := svc.createConnection()
	if err == nil || !strings.Contains(err.Error(), "failed to connect to LDAP") {
		t.Errorf("expected connection failure, got %v", err)
	}
}

// Test LDAPService.TestConnection in mock (development) mode succeeds
func TestLDAPService_TestConnection_Mock(t *testing.T) {
	cfg := &config.Config{Environment: "development", LDAP: config.LDAP{Enabled: true, BaseDN: "dc=example,dc=com"}}
	svc, err := NewLDAPService(nil, nil, cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error creating service: %v", err)
	}
	if err := svc.TestConnection(); err != nil {
		t.Errorf("expected TestConnection success in mock mode, got %v", err)
	}
}
