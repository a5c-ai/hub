package services

import (
	"context"
	"errors"
	"sync"

	"github.com/a5c-ai/hub/pkg/plugin"
)

// PluginService defines operations for plugin marketplace and installations.
type PluginService interface {
	ListMarketplace(ctx context.Context) ([]*plugin.Manifest, error)
	InstallOrgPlugin(ctx context.Context, org, name string, settings map[string]interface{}) error
	UninstallOrgPlugin(ctx context.Context, org, name string) error
	InstallRepoPlugin(ctx context.Context, owner, repo, name string, settings map[string]interface{}) error
	UninstallRepoPlugin(ctx context.Context, owner, repo, name string) error
}

// pluginService is an in-memory implementation of PluginService.
type pluginService struct {
	mu           sync.RWMutex
	marketplace  map[string]*plugin.Manifest
	orgInstalls  map[string]map[string]map[string]interface{}
	repoInstalls map[string]map[string]map[string]interface{}
}

// NewPluginService creates a new in-memory plugin service instance.
func NewPluginService() PluginService {
	return &pluginService{
		marketplace:  make(map[string]*plugin.Manifest),
		orgInstalls:  make(map[string]map[string]map[string]interface{}),
		repoInstalls: make(map[string]map[string]map[string]interface{}),
	}
}

// ListMarketplace returns all registered plugin manifests.
func (s *pluginService) ListMarketplace(ctx context.Context) ([]*plugin.Manifest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*plugin.Manifest, 0, len(s.marketplace))
	for _, m := range s.marketplace {
		list = append(list, m)
	}
	return list, nil
}

// InstallOrgPlugin installs or updates a plugin for an organization.
func (s *pluginService) InstallOrgPlugin(ctx context.Context, org, name string, settings map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.marketplace[name]; !ok {
		return errors.New("plugin not found")
	}
	if s.orgInstalls[org] == nil {
		s.orgInstalls[org] = make(map[string]map[string]interface{})
	}
	s.orgInstalls[org][name] = settings
	return nil
}

// UninstallOrgPlugin removes a plugin installation from an organization.
func (s *pluginService) UninstallOrgPlugin(ctx context.Context, org, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if installs, ok := s.orgInstalls[org]; ok {
		delete(installs, name)
	}
	return nil
}

// InstallRepoPlugin installs or updates a plugin for a repository.
func (s *pluginService) InstallRepoPlugin(ctx context.Context, owner, repo, name string, settings map[string]interface{}) error {
	key := owner + "/" + repo
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.marketplace[name]; !ok {
		return errors.New("plugin not found")
	}
	if s.repoInstalls[key] == nil {
		s.repoInstalls[key] = make(map[string]map[string]interface{})
	}
	s.repoInstalls[key][name] = settings
	return nil
}

// UninstallRepoPlugin removes a plugin installation from a repository.
func (s *pluginService) UninstallRepoPlugin(ctx context.Context, owner, repo, name string) error {
	key := owner + "/" + repo
	s.mu.Lock()
	defer s.mu.Unlock()
	if installs, ok := s.repoInstalls[key]; ok {
		delete(installs, name)
	}
	return nil
}
