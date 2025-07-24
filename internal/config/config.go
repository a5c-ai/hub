package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment string     `mapstructure:"environment"`
	LogLevel    int        `mapstructure:"log_level"`
	Server      Server     `mapstructure:"server"`
	Database    Database   `mapstructure:"database"`
	JWT         JWT        `mapstructure:"jwt"`
	CORS        CORS       `mapstructure:"cors"`
	Storage     Storage    `mapstructure:"storage"`
	OAuth       OAuth      `mapstructure:"oauth"`
	SAML        SAML       `mapstructure:"saml"`
	LDAP        LDAP       `mapstructure:"ldap"`
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

type OAuth struct {
	GitHub    GitHubOAuth    `mapstructure:"github"`
	Google    GoogleOAuth    `mapstructure:"google"`
	Microsoft MicrosoftOAuth `mapstructure:"microsoft"`
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

type SAML struct {
	Enabled       bool   `mapstructure:"enabled"`
	EntityID      string `mapstructure:"entity_id"`
	SSOURL        string `mapstructure:"sso_url"`
	Certificate   string `mapstructure:"certificate"`
	PrivateKey    string `mapstructure:"private_key"`
	AttributeMap  map[string]string `mapstructure:"attribute_map"`
}

type LDAP struct {
	Enabled    bool   `mapstructure:"enabled"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	BaseDN     string `mapstructure:"base_dn"`
	BindDN     string `mapstructure:"bind_dn"`
	BindPassword string `mapstructure:"bind_password"`
	UserFilter string `mapstructure:"user_filter"`
	AttributeMap map[string]string `mapstructure:"attribute_map"`
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
	viper.BindEnv("storage.repository_path", "REPOSITORY_PATH")

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