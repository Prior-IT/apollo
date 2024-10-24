package config

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

func (l LogLevel) ToSlog() slog.Level {
	switch l {
	case LogLevelDebug:
		return slog.LevelDebug
	case LogLevelInfo:
		return slog.LevelInfo
	case LogLevelWarn:
		return slog.LevelWarn
	case LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type LogFormat string

const (
	LogFormatPlaintext LogFormat = "plaintext"
	LogFormatJSON      LogFormat = "json"
)

type AppEnv string

const (
	AppEnvDev        AppEnv = "dev"
	AppEnvProduction AppEnv = "production"
)

type Config struct {
	App            AppConfig
	Sentry         SentryConfig
	Database       DatabaseConfig
	Log            LogConfig
	OAuthProviders map[string]OauthProviderConfig `mapstructure:"OAUTH"`
}

type AppConfig struct {
	Debug                  bool
	SSL                    bool   `default:"true"`
	Port                   uint32 `default:"3000"`
	ProxyPort              uint32
	BasePath               string `default:""`
	Host                   string
	URL                    string
	Name                   string
	Env                    AppEnv `default:"prod"`
	Version                string
	RequestTimeout         uint32 `default:"30"` // in seconds
	AuthenticationKey      string `mapstructure:"AUTHKEY"`
	EncryptionKey          string `mapstructure:"ENCKEY"`
	DefaultPermissionGroup int    `mapstructure:"DEFAULTPERMGROUP"`
}

type SentryConfig struct {
	Enabled      bool
	DSN          string
	SampleRate   float64
	TracesRate   float64
	ProfilesRate float64
	ReplayRate   float64
}

type DatabaseConfig struct {
	URL    string
	Schema string `default:"public"`
}

type LogConfig struct {
	Format  LogFormat `default:"json"`
	Level   LogLevel
	Verbose bool
}

type OauthProviderConfig struct {
	ID            string
	Secret        string
	Scope         []string
	AuthURL       string
	TokenURL      string
	DeviceAuthURL string `default:""`
	UserURL       string
}

func (c Config) BaseURL() string {
	url := c.App.URL
	// If no url was specified, build one from the host and port values
	if len(c.App.URL) == 0 {
		port := c.App.Port
		if c.App.ProxyPort > 0 {
			port = c.App.ProxyPort
		}
		url = fmt.Sprintf("%v:%v", c.App.Host, port)
	}
	protocol := "http"
	if c.App.SSL {
		protocol = "https"
	}
	return fmt.Sprintf(
		"%s://%s",
		protocol,
		url,
	)
}

func (c *Config) IsTest() bool {
	return flag.Lookup("test.v") != nil || strings.HasSuffix(os.Args[0], ".test") ||
		strings.Contains(os.Args[0], "/_test/")
}

func Load(configFS embed.FS, dotenvFiles ...string) (*Config, error) {
	configFile, err := configFS.ReadFile("config.toml")
	if err != nil {
		return nil, fmt.Errorf("could not read config.toml from your configFS: %w", err)
	}

	reader := viper.NewWithOptions(viper.KeyDelimiter("_"))
	reader.SetConfigType("toml")

	if err = reader.ReadConfig(bytes.NewBuffer(configFile)); err != nil {
		return nil, fmt.Errorf("could not load the app configuration: %w", err)
	}

	// Environment override
	err = godotenv.Load(dotenvFiles...)
	if errors.Is(err, os.ErrNotExist) {
		slog.Warn("No .env file found, continuing...")
	} else if err != nil {
		return nil, fmt.Errorf(".env file found, but could not load it: %w", err)
	}
	reader.AutomaticEnv()

	var config Config
	if err := reader.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("invalid config format: %w", err)
	}

	if config.App.Debug && !config.IsTest() {
		slog.Warn("APP_DEBUG is turned on, do not run this mode in production!")
	}

	return &config, nil
}
