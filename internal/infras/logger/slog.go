package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"hermes-ai/internal/infras/ctxkey"
	"hermes-ai/internal/infras/utils"
)

// Default 初始化默认logger对象
func Default(options ...Option) {
	opts := []Option{
		WithAddSource(true),
		WithEnableJSON(),
	}
	if len(options) > 0 {
		opts = append(opts, options...)
	}

	Init(opts...)
}

// Init golang slog package init
func Init(opts ...Option) {
	opt := Options{
		level:        slog.LevelInfo, // default:info level
		replaceAttr:  nil,
		outputToFile: false,
		logDir:       "./logs",
		fileName:     "hermes-ai.log",
	}

	for _, o := range opts {
		o(&opt)
	}

	var writer io.Writer
	if opt.outputToFile {
		if opt.fileMaxSize == 0 {
			opt.fileMaxSize = 100
		}

		if opt.fileName == "" {
			opt.fileName = "hermes-ai.log"
		}

		if opt.logDir == "" {
			opt.logDir = "./logs"
		}

		opt.fileName = filepath.Join(opt.logDir, opt.fileName)
		if opt.fileMaxBackups == 0 {
			opt.fileMaxBackups = 10
		}

		if opt.fileMaxAge == 0 {
			opt.fileMaxAge = 30
		}

		// set lumberjack
		w := &lumberjack.Logger{
			Filename:   opt.fileName,
			MaxSize:    opt.fileMaxSize,
			MaxBackups: opt.fileMaxBackups,
			MaxAge:     opt.fileMaxAge,
			Compress:   opt.compress,
		}

		writer = io.MultiWriter(os.Stdout, w)
	} else {
		writer = os.Stdout
	}

	var handler slog.Handler
	handlerOption := &slog.HandlerOptions{
		Level:       opt.level,
		AddSource:   opt.addSource,
		ReplaceAttr: opt.replaceAttr,
	}

	if opt.json {
		handler = slog.NewJSONHandler(writer, handlerOption)
	} else {
		handler = slog.NewTextHandler(writer, handlerOption)
	}

	slog.SetDefault(slog.New(handler))
}

// With returns a Logger that includes the given attributes
// in each output operation. Arguments are converted to
// attributes as if by slog [Logger.Log].
func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

// GetLevel 获取日志级别
func GetLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GetRequestID 从ctx上获取request_id
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return utils.UUID()
	}

	requestID, ok := ctx.Value(ctxkey.RequestIdKey).(string)
	if requestID == "" || !ok {
		return utils.UUID()
	}

	return requestID
}
