package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Database   DatabaseConfig `yaml:"database"`
	Logging    LoggingConfig  `yaml:"logging"`
	ConfigFile string
}

// const RedisAddr = "127.0.0.1:6378"
var DbHost string

var Database *DatabaseConfig
var Logging *LoggingConfig
var SSO *casdoorsdk.Client

func Setup() {
	var Http = &AppConfig{
		ConfigFile: "config.yaml",
	}

	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(err)
	}

	certificate_content, err1 := os.ReadFile("token_jwt_key.pem")
	if err1 != nil {
		slog.Error("failed to read sso certificate", "error", err1)
		os.Exit(2)
	}

	if err := cleanenv.ReadConfig("config.yaml", Http); err != nil {
		slog.Error("file not found", "file", "config.yaml")
		os.Exit(2)
	}

	DbHost = Config("DB_HOST")

	certificate_key := string(certificate_content)
	os.Setenv("SSO_CERTIFICATE", certificate_key)
	sso_client := casdoorsdk.NewClient(Config("SSO_END_POINT"), Config("SSO_CLIENT_ID"), Config("SSO_CLIENT_SECRET"), Config("SSO_CERTIFICATE"), Config("SSO_ORGANIZATION_NAME"), Config("SSO_APPLICATION_NAME"))

	SSO = sso_client

	InitLoggers()

	Http.Logging.Setup()
	Logging = &Http.Logging
	Http.Database.Setup()
	Database = &Http.Database
}
func Config(key string) string {
	return os.Getenv(key)
}
