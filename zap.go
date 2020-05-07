package gbase

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ReplaceZapLogger use default log style to replace zap global logger
func ReplaceZapLogger(level, outpath, encoding string, showCaller bool) (func(), error) {
	var logger, err = NewZapLogger(level, outpath, encoding, showCaller)
	if err != nil {
		return nil, err
	}
	return zap.ReplaceGlobals(logger), nil
}

// NewZapLogger return default log style
func NewZapLogger(level, outpath, encoding string, showCaller bool) (*zap.Logger, error) {
	var zlevel zapcore.Level
	switch level {
	case "debug":
		zlevel = zap.DebugLevel
	case "info":
		zlevel = zap.InfoLevel
	case "warn":
		zlevel = zap.WarnLevel
	case "error":
		zlevel = zap.ErrorLevel
	case "panic":
		zlevel = zap.PanicLevel
	case "fatal":
		zlevel = zap.FatalLevel
	default:
		return nil, errors.New("Unexpected log level " + level)
	}

	if dir := filepath.Dir(outpath); dir != "." && dir != ".." && dir != "/" {
		if _, e := os.Stat(dir); errors.Is(e, os.ErrNotExist) {
			if e := os.MkdirAll(dir, 0755); e != nil {
				return nil, e
			}
		}
	}

	var zcfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zlevel),
		OutputPaths:      []string{outpath},
		ErrorOutputPaths: []string{outpath},
		Encoding:         encoding,
		DisableCaller:    !showCaller,

		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "T",
			LevelKey:   "L",
			NameKey:    "N",
			MessageKey: "M",
			LineEnding: "\n",

			CallerKey:     "Caller",
			StacktraceKey: "Stack",

			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeDuration: zapcore.NanosDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
			},
		},
	}

	return zcfg.Build()
}
