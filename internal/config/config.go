package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment   string        `mapstructure:"environment"`
	LogLevel      int           `mapstructure:"log_level"`
	Server        Server        `mapstructure:"server"`
	Database      Database      `mapstructure:"database"`
	JWT           JWT           `mapstructure:"jwt"`
	CORS          CORS          `mapstructure:"cors"`
	Storage       Storage       `mapstructure:"storage"`
	Security      Security      `mapstructure:"security"`
	OAuth         OAuth         `mapstructure:"oauth"`
	SAML          SAML          `mapstructure:"saml"`
	LDAP          LDAP          `mapstructure:"ldap"`
	SMTP          SMTP          `mapstructure:"smtp"`
	SSH           SSH           `mapstructure:"ssh"`
	Elasticsearch Elasticsearch `mapstructure:"elasticsearch"`
	Application   Application   `mapstructure:"application"`
}

type Server struct {
	Port int `mapstructure:"port"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWT struct {
	Secret         string `mapstructure:"secret"`
	ExpirationHour int    `mapstructure:"expiration_hour"`
}

type CORS struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type Storage struct {
	RepositoryPath string `mapstructure:"repository_path"`
}

type Security struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

type OAuth struct {
	GitHub    GitHubOAuth    `mapstructure:"github"`
	Google    GoogleOAuth    `mapstructure:"google"`
	Microsoft MicrosoftOAuth `mapstructure:"microsoft"`
	GitLab    GitLabOAuth    `mapstructure:"gitlab"`
}

type GitHubOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type GoogleOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type MicrosoftOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
	TenantID     string `mapstructure:"tenant_id"`
}

type GitLabOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
	BaseURL      string `mapstructure:"base_url"`
}

type SAML struct {
	Enabled      bool              `mapstructure:"enabled"`
	EntityID     string            `mapstructure:"entity_id"`
	SSOURL       string            `mapstructure:"sso_url"`
	Certificate  string            `mapstructure:"certificate"`
	PrivateKey   string            `mapstructure:"private_key"`
	AttributeMap map[string]string `mapstructure:"attribute_map"`
}

type LDAP struct {
	Enabled      bool              `mapstructure:"enabled"`
	Host         string            `mapstructure:"host"`
	Port         int               `mapstructure:"port"`
	BaseDN       string            `mapstructure:"base_dn"`
	BindDN       string            `mapstructure:"bind_dn"`
	BindPassword string            `mapstructure:"bind_password"`
	UserFilter   string            `mapstructure:"user_filter"`
	AttributeMap map[string]string `mapstructure:"attribute_map"`
}

type SSH struct {
	Enabled     bool   `mapstructure:"enabled"`
	Port        int    `mapstructure:"port"`
	HostKeyPath string `mapstructure:"host_key_path"`
}

type SMTP struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

type Elasticsearch struct {
	Enabled    bool     `mapstructure:"enabled"`
	Addresses  []string `mapstructure:"addresses"`
	Username   string   `mapstructure:"username"`
	Password   string   `mapstructure:"password"`
	CloudID    string   `mapstructure:"cloud_id"`
	APIKey     string   `mapstructure:"api_key"`
	IndexPrefix string  `mapstructure:"index_prefix"`
}

type Application struct {
	BaseURL string `mapstructure:"base_url"`
	Name    string `mapstructure:"name"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", 4)
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "hub")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "hub")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration_hour", 24)
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("storage.repository_path", "/var/lib/hub/repositories")
	viper.SetDefault("security.encryption_key", "default-32-byte-key-for-secrets")
	viper.SetDefault("ssh.enabled", true)
	viper.SetDefault("ssh.port", 2222)
	viper.SetDefault("ssh.host_key_path", "./ssh_host_key")
	viper.SetDefault("smtp.host", "")
	viper.SetDefault("smtp.port", "587")
	viper.SetDefault("smtp.username", "")
	viper.SetDefault("smtp.password", "")
	viper.SetDefault("smtp.from", "noreply@localhost")
	viper.SetDefault("smtp.use_tls", true)
	viper.SetDefault("elasticsearch.enabled", false)
	viper.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.index_prefix", "hub")
	viper.SetDefault("application.base_url", "http://localhost:3000")
	viper.SetDefault("application.name", "A5C Hub")

	viper.AutomaticEnv()

	viper.BindEnv("environment", "ENVIRONMENT")
	viper.BindEnv("log_level", "LOG_LEVEL")
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiration_hour", "JWT_EXPIRATION_HOUR")
	viper.BindEnv("oauth.github.client_id", "GITHUB_CLIENT_ID")
	viper.BindEnv("oauth.github.client_secret", "GITHUB_CLIENT_SECRET")
	viper.BindEnv("oauth.google.client_id", "GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.google.client_secret", "GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.microsoft.client_id", "MICROSOFT_CLIENT_ID")
	viper.BindEnv("oauth.microsoft.client_secret", "MICROSOFT_CLIENT_SECRET")
	viper.BindEnv("oauth.microsoft.tenant_id", "MICROSOFT_TENANT_ID")
	viper.BindEnv("oauth.gitlab.client_id", "GITLAB_CLIENT_ID")
	viper.BindEnv("oauth.gitlab.client_secret", "GITLAB_CLIENT_SECRET")
	viper.BindEnv("oauth.gitlab.base_url", "GITLAB_BASE_URL")
	viper.BindEnv("storage.repository_path", "REPOSITORY_PATH")
	viper.BindEnv("security.encryption_key", "ENCRYPTION_KEY")
	viper.BindEnv("ssh.enabled", "SSH_ENABLED")
	viper.BindEnv("ssh.port", "SSH_PORT")
	viper.BindEnv("ssh.host_key_path", "SSH_HOST_KEY_PATH")
	viper.BindEnv("smtp.host", "SMTP_HOST")
	viper.BindEnv("smtp.port", "SMTP_PORT")
	viper.BindEnv("smtp.username", "SMTP_USERNAME")
	viper.BindEnv("smtp.password", "SMTP_PASSWORD")
	viper.BindEnv("smtp.from", "SMTP_FROM")
	viper.BindEnv("smtp.use_tls", "SMTP_USE_TLS")
	viper.BindEnv("elasticsearch.enabled", "ELASTICSEARCH_ENABLED")
	viper.BindEnv("elasticsearch.addresses", "ELASTICSEARCH_ADDRESSES")
	viper.BindEnv("elasticsearch.username", "ELASTICSEARCH_USERNAME")
	viper.BindEnv("elasticsearch.password", "ELASTICSEARCH_PASSWORD")
	viper.BindEnv("elasticsearch.cloud_id", "ELASTICSEARCH_CLOUD_ID")
	viper.BindEnv("elasticsearch.api_key", "ELASTICSEARCH_API_KEY")
	viper.BindEnv("elasticsearch.index_prefix", "ELASTICSEARCH_INDEX_PREFIX")
	viper.BindEnv("application.base_url", "BASE_URL")
	viper.BindEnv("application.name", "APPLICATION_NAME")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
