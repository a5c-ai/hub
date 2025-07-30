package plugin

// Config captures runtime configuration and manifest for a plugin.
type Config struct {
	Manifest *Manifest
	APIToken string
	Settings map[string]interface{}
}

// NewConfig creates a plugin Config from a parsed manifest, API token, and settings map.
func NewConfig(manifest *Manifest, apiToken string, settings map[string]interface{}) *Config {
	return &Config{
		Manifest: manifest,
		APIToken: apiToken,
		Settings: settings,
	}
}

// GetSetting retrieves a setting value by name, or nil if not defined.
func (c *Config) GetSetting(name string) interface{} {
	if c.Settings == nil {
		return nil
	}
	return c.Settings[name]
}
