package app

import (
	"context"
	"core/internal/domain_model/model"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

type DatabaseConfig struct {
	*gorm.DB
	Driver      string `yaml:"driver" env:"DB_DRIVER"`
	Host        string `yaml:"host" env:"DB_HOST"`
	Username    string `yaml:"username" env:"DB_USER"`
	Password    string `yaml:"password" env:"DB_PASSWORD"`
	DBName      string `yaml:"db_name" env:"DB_NAME"`
	Port        int    `yaml:"port" env:"DB_PORT"`
	Connections int    `yaml:"connections" env:"DB_CONNECTIONS"`
	Debug       bool   `yaml:"debug"`
	MaxIdleConn int    `env:"MAX_IDLE_CONN"`
	MaxOpenConn int    `env:"MAX_OPEN_CONN"`
}

func (cf *DatabaseConfig) Setup() {
	mainDbDNS := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cf.Host, cf.Port, cf.Username, cf.Password, cf.DBName)

	logLevel := logger.Warn
	if IsInProduction() {
		logLevel = logger.Warn
	}

	DB, err := gorm.Open(postgres.Open(mainDbDNS), &gorm.Config{
		PrepareStmt: true,
		Logger: NewSlogLogger(slog.Default().With("module", "gorm"), logger.Config{
			SlowThreshold:             50 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
		}),
	})

	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}

	if cf.Debug {
		DB = DB.Debug()
	}
	sqlDB, _ := DB.DB()
	cf.DB = DB
	sqlDB.SetMaxIdleConns(10)  // MAX_IDLE_CONN
	sqlDB.SetMaxOpenConns(100) // MAX_OPEN_CONN
	tracePlugin := tracing.NewPlugin()
	DB.Use(tracePlugin)

	MigrateDatabase(DB)
}

func MigrateDatabase(DB *gorm.DB) {
	if tx := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"); tx.Error != nil {
		slog.Debug("failed to create uuid extension", "error", tx.Error)
	}

	//must be pointer
	migrations := []any{
		&model.User{},
		&model.Admin{},
		&model.SuperAdmin{},
		&model.Document{},
		&model.UserDocument{},
		&model.SystemSetting{},
	}

	if err := DB.AutoMigrate(migrations...); err != nil {
		slog.Error("Database migration failed", "error", err)
		panic(err)
	}

}

type slogLogger struct {
	Logger                    *slog.Logger
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	Parameterized             bool
	Colorful                  bool // Ignored in slog
	IgnoreRecordNotFoundError bool
}

func NewSlogLogger(logger *slog.Logger, config logger.Config) logger.Interface {
	return &slogLogger{
		Logger:                    logger,
		LogLevel:                  config.LogLevel,
		SlowThreshold:             config.SlowThreshold,
		Parameterized:             config.ParameterizedQueries,
		IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
	}
}

func (l *slogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *slogLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		l.Logger.InfoContext(ctx, msg, slog.Any("data", data))
	}
}

func (l *slogLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		l.Logger.WarnContext(ctx, msg, slog.Any("data", data))
	}
}

func (l *slogLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		l.Logger.ErrorContext(ctx, msg, slog.Any("data", data))
	}
}

func (l *slogLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []any{
		slog.String("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)),
		slog.String("query", sql),
	}

	if rows != -1 {
		fields = append(fields, slog.Int64("row_affected", rows))
	}

	switch {
	case err != nil && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		fields = append(fields, slog.String("error", err.Error()))
		l.Logger.ErrorContext(ctx, "SQL trace", fields...)

	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold:
		l.Logger.WarnContext(ctx, "SQL trace", fields...)

	case l.LogLevel >= logger.Info:
		l.Logger.InfoContext(ctx, "SQL trace", fields...)
	}
}
