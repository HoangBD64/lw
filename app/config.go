package app

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/sony/sonyflake"
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
var IDGenerator *sonyflake.Sonyflake

func Setup() {
	var Http = &AppConfig{
		ConfigFile: "config.yaml",
	}

	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(err)
	}

	envMachineID := Config("MACHINE_ID")

	machineID, err := strconv.Atoi(envMachineID)
	if err != nil {
		slog.Error("invalid machine id format", "error", err)
		os.Exit(2)
	}

	snf, err := sonyflake.New(sonyflake.Settings{
		StartTime:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID:      func() (uint16, error) { return uint16(machineID), nil },
		CheckMachineID: func(uint16) bool { return true },
	})

	if err != nil {
		slog.Error("failed to create sonyflake instance", "error", err)
		os.Exit(2)
	}

	IDGenerator = snf

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
