package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment string   `mapstructure:"environment"`
	LogLevel    int      `mapstructure:"log_level"`
	Server      Server   `mapstructure:"server"`
	Database    Database `mapstructure:"database"`
	Redis       Redis    `mapstructure:"redis"`
	JWT         JWT      `mapstructure:"jwt"`
	CORS        CORS     `mapstructure:"cors"`
	Storage     Storage  `mapstructure:"storage"`
	Security    Security `mapstructure:"security"`
	OAuth       OAuth    `mapstructure:"oauth"`
	// GitHub integration tokens configuration
	GitHub        GitHubIntegration `mapstructure:"github"`
	SAML          SAML              `mapstructure:"saml"`
	LDAP          LDAP              `mapstructure:"ldap"`
	SMTP          SMTP              `mapstructure:"smtp"`
	SSH           SSH               `mapstructure:"ssh"`
	Elasticsearch Elasticsearch     `mapstructure:"elasticsearch"`
	Application   Application       `mapstructure:"application"`
	// Git LFS configuration
	LFS LFS `mapstructure:"lfs"`
}

// LFS holds Git LFS storage configuration
type LFS struct {
	// Storage backend for Git LFS: "azure_blob", "s3", "filesystem"
	Backend string       `mapstructure:"backend"`
	Azure   AzureStorage `mapstructure:"azure"`
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

type Redis struct {
	Enabled    bool   `mapstructure:"enabled"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type JWT struct {
	Secret         string `mapstructure:"secret"`
	ExpirationHour int    `mapstructure:"expiration_hour"`
}

type CORS struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type Storage struct {
	RepositoryPath string          `mapstructure:"repository_path"`
	Artifacts      ArtifactStorage `mapstructure:"artifacts"`
}

type ArtifactStorage struct {
	Backend       string       `mapstructure:"backend"` // "azure", "s3", "filesystem"
	Azure         AzureStorage `mapstructure:"azure"`
	S3            S3Storage    `mapstructure:"s3"`
	MaxSizeMB     int64        `mapstructure:"max_size_mb"`    // Max artifact size in MB
	RetentionDays int          `mapstructure:"retention_days"` // Retention period in days
	BasePath      string       `mapstructure:"base_path"`      // For filesystem backend
}

type AzureStorage struct {
	AccountName   string `mapstructure:"account_name"`
	AccountKey    string `mapstructure:"account_key"`
	ContainerName string `mapstructure:"container_name"`
	EndpointURL   string `mapstructure:"endpoint_url"`
}

type S3Storage struct {
	Region          string `mapstructure:"region"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	EndpointURL     string `mapstructure:"endpoint_url"` // For S3-compatible services
	UseSSL          bool   `mapstructure:"use_ssl"`
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

// GitHubIntegration holds configuration for GitHub API integration tokens.
type GitHubIntegration struct {
	ClientID     string       `mapstructure:"client_id"`
	ClientSecret string       `mapstructure:"client_secret"`
	Tokens       GitHubTokens `mapstructure:"tokens"`
}

// GitHubTokens specifies per-organization and per-user personal access tokens.
type GitHubTokens struct {
	Organizations map[string]string `mapstructure:"organizations"`
	Users         map[string]string `mapstructure:"users"`
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
	Enabled     bool     `mapstructure:"enabled"`
	Addresses   []string `mapstructure:"addresses"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
	CloudID     string   `mapstructure:"cloud_id"`
	APIKey      string   `mapstructure:"api_key"`
	IndexPrefix string   `mapstructure:"index_prefix"`
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
	viper.SetDefault("redis.enabled", false)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration_hour", 24)
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("storage.repository_path", "/repositories")
	viper.SetDefault("storage.artifacts.backend", "filesystem")
	viper.SetDefault("storage.artifacts.base_path", "/var/lib/hub/artifacts")
	viper.SetDefault("storage.artifacts.max_size_mb", 1024)
	viper.SetDefault("storage.artifacts.retention_days", 90)
	viper.SetDefault("storage.artifacts.azure.container_name", "artifacts")
	viper.SetDefault("storage.artifacts.s3.use_ssl", true)
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
	// Git LFS defaults
	viper.SetDefault("lfs.backend", "filesystem")
	viper.SetDefault("lfs.azure.account_name", "")
	viper.SetDefault("lfs.azure.account_key", "")
	viper.SetDefault("lfs.azure.container_name", "lfs")

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
	viper.BindEnv("redis.enabled", "REDIS_ENABLED")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.max_retries", "REDIS_MAX_RETRIES")
	viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiration_hour", "JWT_EXPIRATION_HOUR")
	viper.BindEnv("oauth.github.client_id", "AUTH_GITHUB_CLIENT_ID")
	viper.BindEnv("oauth.github.client_secret", "AUTH_GITHUB_CLIENT_SECRET")
	viper.BindEnv("oauth.google.client_id", "GOOGLE_CLIENT_ID")
	viper.BindEnv("oauth.google.client_secret", "GOOGLE_CLIENT_SECRET")
	viper.BindEnv("oauth.microsoft.client_id", "MICROSOFT_CLIENT_ID")
	viper.BindEnv("oauth.microsoft.client_secret", "MICROSOFT_CLIENT_SECRET")
	viper.BindEnv("oauth.microsoft.tenant_id", "MICROSOFT_TENANT_ID")
	viper.BindEnv("oauth.gitlab.client_id", "GITLAB_CLIENT_ID")
	viper.BindEnv("oauth.gitlab.client_secret", "GITLAB_CLIENT_SECRET")
	viper.BindEnv("oauth.gitlab.base_url", "GITLAB_BASE_URL")
	viper.BindEnv("storage.repository_path", "REPOSITORY_PATH")
	viper.BindEnv("storage.artifacts.backend", "ARTIFACT_STORAGE_BACKEND")
	viper.BindEnv("storage.artifacts.base_path", "ARTIFACT_STORAGE_PATH")
	viper.BindEnv("storage.artifacts.max_size_mb", "ARTIFACT_MAX_SIZE_MB")
	viper.BindEnv("storage.artifacts.retention_days", "ARTIFACT_RETENTION_DAYS")
	viper.BindEnv("storage.artifacts.azure.account_name", "AZURE_STORAGE_ACCOUNT_NAME")
	viper.BindEnv("storage.artifacts.azure.account_key", "AZURE_STORAGE_ACCOUNT_KEY")
	viper.BindEnv("storage.artifacts.azure.container_name", "AZURE_STORAGE_CONTAINER_NAME")
	viper.BindEnv("storage.artifacts.azure.endpoint_url", "AZURE_STORAGE_ENDPOINT_URL")
	viper.BindEnv("storage.artifacts.s3.region", "AWS_REGION")
	viper.BindEnv("storage.artifacts.s3.bucket", "S3_BUCKET")
	viper.BindEnv("storage.artifacts.s3.access_key_id", "AWS_ACCESS_KEY_ID")
	viper.BindEnv("storage.artifacts.s3.secret_access_key", "AWS_SECRET_ACCESS_KEY")
	viper.BindEnv("storage.artifacts.s3.endpoint_url", "S3_ENDPOINT_URL")
	viper.BindEnv("storage.artifacts.s3.use_ssl", "S3_USE_SSL")
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
	// Git LFS env bindings
	viper.BindEnv("lfs.backend", "LFS_BACKEND")
	viper.BindEnv("lfs.azure.account_name", "LFS_AZURE_ACCOUNT_NAME")
	viper.BindEnv("lfs.azure.account_key", "LFS_AZURE_ACCOUNT_KEY")
	viper.BindEnv("lfs.azure.container_name", "LFS_AZURE_CONTAINER_NAME")

	// GitHub integration defaults and env bindings
	viper.SetDefault("github.client_id", "")
	viper.SetDefault("github.client_secret", "")
	viper.SetDefault("github.tokens.organizations", map[string]string{})
	viper.SetDefault("github.tokens.users", map[string]string{})
	viper.BindEnv("github.client_id", "GITHUB_CLIENT_ID")
	viper.BindEnv("github.client_secret", "GITHUB_CLIENT_SECRET")
	viper.BindEnv("github.tokens.organizations", "GITHUB_TOKENS_ORGANIZATIONS")
	viper.BindEnv("github.tokens.users", "GITHUB_TOKENS_USERS")

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
