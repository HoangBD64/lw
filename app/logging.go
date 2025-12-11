package app

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lmittmann/tint"
	"github.com/samber/lo"
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

type LoggingConfig struct {
	Type       string `yaml:"type"`
	ServerName string `yaml:"server_name"`
}

func init() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
}

func (log *LoggingConfig) Setup() {
	logrus.SetOutput(os.Stdout)
	disableColor := false
	if IsInProduction() {
		logrus.SetLevel(logrus.InfoLevel)
		disableColor = true
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
	my_formater := &logrus.TextFormatter{
		DisableColors: disableColor,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			package_name := f.Func.Name()
			spliter_1 := strings.Split(package_name, ".")
			if len(spliter_1) > 0 {
				spliter_2 := strings.Split(spliter_1[0], "/")
				if len(spliter_2) > 1 {
					paths := spliter_2[1:]
					fullpath := strings.Join(paths, "/")
					return fmt.Sprintf("%s/%s:%d", fullpath, filename, f.Line), fmt.Sprintf("%s()", f.Function)
				}
			}
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
	}
	logrus.SetFormatter(my_formater)
	logrus.SetReportCaller(true)
}

func InitLoggers() {
	var stdoutLogger slog.Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	if !IsInProduction() {
		stdoutLogger = tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug, TimeFormat: time.DateTime, ReplaceAttr: colorReplaceAttrFunc})
	}

	slogFormatter := newLogFormatterForMonitors()
	if !IsInProduction() {
		slogFormatter = newLogFormatter(true, 50)
	}

	// handler logger output
	handler := slogmulti.Pipe(
		slogFormatter,
	).Handler(slogmulti.Fanout(stdoutLogger))

	// set global logger with custom options
	slog.SetDefault(slog.New(handler))
}

func IsInProduction() bool {
	return Config("DEV_ENV") != "true"
}

type formatterSetting struct {
	formatter slogformatter.Formatter
	enabled   bool
}

func newLogFormatter(isDev bool, maxLength int) slogmulti.Middleware {
	trimOK := isDev

	jsonFormatter := slogformatter.FormatByType(func(rawData datatypes.JSON) slog.Value {
		data := string(rawData)
		if trimOK && len(data) > maxLength {
			return slog.StringValue(data[:maxLength] + "...")
		}
		return slog.StringValue(data)
	})

	bytesFormatter := slogformatter.FormatByType(func(rawData []byte) slog.Value {
		data := string(rawData)
		if trimOK && len(data) > maxLength {
			return slog.StringValue(data[:maxLength] + "...")
		}
		return slog.StringValue(data)
	})

	uuidPtrFormatter := slogformatter.FormatByType(func(rawData *uuid.UUID) slog.Value { return slog.StringValue(rawData.String()) })
	uuidFormatter := slogformatter.FormatByType(func(rawData uuid.UUID) slog.Value { return slog.StringValue(rawData.String()) })

	formatters := []formatterSetting{
		{bytesFormatter, true},
		{jsonFormatter, true},
		{uuidFormatter, true},
		{uuidPtrFormatter, true},
	}

	enabledFormatters := lo.FilterMap(formatters, func(f formatterSetting, _ int) (slogformatter.Formatter, bool) { return f.formatter, f.enabled })
	return slogformatter.NewFormatterMiddleware(enabledFormatters...)
}

func newLogFormatterForMonitors() slogmulti.Middleware {
	bytesFormatter := slogformatter.FormatByType(func(rawData []byte) slog.Value { return slog.StringValue(string(rawData)) })
	jsonFormatter := slogformatter.FormatByType(func(rawData datatypes.JSON) slog.Value { return slog.StringValue(string(rawData)) })
	uuidPtrFormatter := slogformatter.FormatByType(func(rawData *uuid.UUID) slog.Value { return slog.StringValue(rawData.String()) })
	uuidFormatter := slogformatter.FormatByType(func(rawData uuid.UUID) slog.Value { return slog.StringValue(rawData.String()) })

	formatters := []formatterSetting{
		{jsonFormatter, true},
		{bytesFormatter, true},
		{uuidFormatter, true},
		{uuidPtrFormatter, true},
	}

	enabledFormatters := lo.FilterMap(formatters, func(f formatterSetting, _ int) (slogformatter.Formatter, bool) { return f.formatter, f.enabled })
	return slogformatter.NewFormatterMiddleware(enabledFormatters...)
}

func colorReplaceAttrFunc(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey && len(groups) == 0 {
		level, ok := a.Value.Any().(slog.Level)
		if !ok {
			return a
		}
		switch level.String() {
		case "DEBUG":
			return tint.Attr(225, slog.String(a.Key, "DEBUG"))
		case "INFO":
			return tint.Attr(10, slog.String(a.Key, "INFO"))
		case "WARN":
			return tint.Attr(226, slog.String(a.Key, "WARN"))
		case "ERROR":
			return tint.Attr(1, slog.String(a.Key, "ERROR"))
		}
	}

	return a
}
