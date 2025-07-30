package plugin

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// Manifest represents the plugin manifest structure loaded from plugin.yaml.
type Manifest struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   ManifestMetadata `yaml:"metadata"`
	Spec       ManifestSpec     `yaml:"spec"`
}

// ManifestMetadata holds top-level metadata of the plugin.
type ManifestMetadata struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Website     string `yaml:"website"`
	License     string `yaml:"license"`
}

// ManifestSpec holds the specification details of the plugin.
type ManifestSpec struct {
	Runtime      string       `yaml:"runtime"`
	Entry        string       `yaml:"entry"`
	Permissions  []string     `yaml:"permissions"`
	Hooks        []Hook       `yaml:"hooks"`
	Webhooks     []Webhook    `yaml:"webhooks"`
	Settings     []Setting    `yaml:"settings"`
	Dependencies []Dependency `yaml:"dependencies"`
}

// Hook represents a git hook configuration in the plugin.
type Hook struct {
	Name   string `yaml:"name"`
	Script string `yaml:"script"`
}

// Webhook represents an event hook into platform webhook events.
type Webhook struct {
	Event   string `yaml:"event"`
	Handler string `yaml:"handler"`
}

// Setting defines a configurable setting for the plugin.
type Setting struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Required    bool        `yaml:"required,omitempty"`
	Secret      bool        `yaml:"secret,omitempty"`
	Default     interface{} `yaml:"default,omitempty"`
	Description string      `yaml:"description,omitempty"`
}

// Dependency lists external module dependencies of the plugin.
type Dependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// LoadManifest reads and parses a plugin.yaml manifest file.
func LoadManifest(path string) (*Manifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
