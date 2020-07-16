package gossip

import (
	"bytes"

	"github.com/cjey/gbase/context"
)

const (
	_LOG_LEVEL_DEBUG = 1
	_LOG_LEVEL_INFO  = 2
	_LOG_LEVEL_WARN  = 3
	_LOG_LEVEL_ERROR = 4
)

type logWriter struct {
	ctx context.Context
	lvl int
}

func newLogWriter(ctx context.Context, lvl int) *logWriter {
	return &logWriter{
		ctx: ctx,
		lvl: lvl,
	}
}

func (l *logWriter) trimLevel(p []byte) ([]byte, int) {
	var ol = len(p)
	p = bytes.TrimPrefix(p, []byte("[DEBUG] "))
	if len(p) < ol {
		return p, _LOG_LEVEL_DEBUG
	}
	p = bytes.TrimPrefix(p, []byte("[INFO] "))
	if len(p) < ol {
		return p, _LOG_LEVEL_INFO
	}
	p = bytes.TrimPrefix(p, []byte("[WARN] "))
	if len(p) < ol {
		return p, _LOG_LEVEL_WARN
	}
	p = bytes.TrimPrefix(p, []byte("[ERR] "))
	if len(p) < ol {
		return p, _LOG_LEVEL_ERROR
	}
	p = bytes.TrimPrefix(p, []byte("[ERROR] "))
	if len(p) < ol {
		return p, _LOG_LEVEL_ERROR
	}
	return p, 0
}

func (l *logWriter) Write(p []byte) (int, error) {
	if bytes.Contains(p, []byte("ping")) {
		// ignore all ping log
		return len(p), nil
	}

	var ol = len(p)
	var lvl int
	p = bytes.TrimSpace(p)
	p, lvl = l.trimLevel(p)
	if lvl < l.lvl {
		return ol, nil
	}
	p = bytes.TrimPrefix(p, []byte("memberlist: "))
	switch lvl {
	case _LOG_LEVEL_ERROR:
		var lower bool
		// let encryption message lower level
		for _, kw := range [][]byte{[]byte("Encrypt"), []byte("encrypt"), []byte("Descrypt"), []byte("decrypt")} {
			if bytes.Contains(p, kw) {
				lower = true
				break
				l.ctx.Warn(string(p))
			}
		}
		if lower {
			l.ctx.Warn(string(p))
		} else {
			l.ctx.Error(string(p))
		}
	case _LOG_LEVEL_WARN:
		l.ctx.Warn(string(p))
	case _LOG_LEVEL_INFO:
		l.ctx.Info(string(p))
	case _LOG_LEVEL_DEBUG:
		l.ctx.Debug(string(p))
	}
	return ol, nil
}
