package plugin

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadManifest(t *testing.T) {
	sample := `apiVersion: v1
kind: Plugin
metadata:
  name: test-plugin
  version: "0.1.0"
  description: "demo plugin"
  author: "Author Name"
  website: "https://example.com"
  license: "MIT"
spec:
  runtime: go
  entry: main.go
  permissions:
    - repositories:read
  hooks:
    - name: pre-receive
      script: hooks/pre.sh
  webhooks:
    - event: push
      handler: HandlePush
  settings:
    - name: api_key
      type: string
      required: true
      secret: true
  dependencies:
    - name: github.com/example/dependency
      version: "^1.0.0"
`
	dir, err := ioutil.TempDir("", "plugintest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "plugin.yaml")
	err = ioutil.WriteFile(path, []byte(sample), 0644)
	assert.NoError(t, err)

	m, err := LoadManifest(path)
	assert.NoError(t, err)
	assert.Equal(t, "v1", m.APIVersion)
	assert.Equal(t, "Plugin", m.Kind)
	assert.Equal(t, "test-plugin", m.Metadata.Name)
	assert.Equal(t, "0.1.0", m.Metadata.Version)
	assert.Equal(t, "go", m.Spec.Runtime)
	assert.Len(t, m.Spec.Permissions, 1)
	assert.Len(t, m.Spec.Hooks, 1)
	assert.Len(t, m.Spec.Webhooks, 1)
	assert.Len(t, m.Spec.Settings, 1)
	assert.Len(t, m.Spec.Dependencies, 1)
}
