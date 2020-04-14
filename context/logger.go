package context

import (
	"go.uber.org/zap"
)

// NameJoiner will generate a full name after joining the origin and given.
// Default logger use it to join name when invoking Fork()
type NameJoiner func(origin, given string) (name string)

// LocationJoiner will generate a full location after joining the origin and given.
// Default logger use it to join location when invoking Fork()
type LocationJoiner func(origin, given string) (field, location string)

// Logger is a logger, but it has name and location.
// When print message, it simulates zap's 'with' style,
// encourage print message and data seprately
type Logger interface {
	// Fork return a new logger, with the given name and location,
	// but it's full name and full location should inhert from me.
	// It could just simplely join my name and location.
	// e.g. my name = ServiceMonitor, name = go1, then new name = ServiceMonitor.go1.
	// e.g. my location = CheckPort, location = CheckAlive, then new location = CheckPort/CheckAlive.
	// name will be print as zap logger's name, mostly it looks like a uuid
	// location will be print as a data field, field name is '@' mostly.
	Fork(name, location string) Logger

	// Name return my full name
	Name() string
	// Location return my full location
	Location() string

	// Debug("New user created", "name", "CJey", "sex", "male")
	Debug(msg string, kvs ...interface{})
	Debugf(template string, args ...interface{})
	Info(msg string, kvs ...interface{})
	Infof(template string, args ...interface{})
	Warn(msg string, kvs ...interface{})
	Warnf(template string, args ...interface{})
	Error(msg string, kvs ...interface{})
	Errorf(template string, args ...interface{})
	Panic(msg string, kvs ...interface{})
	Panicf(template string, args ...interface{})
	Fatal(msg string, kvs ...interface{})
	Fatalf(template string, args ...interface{})

	// With return a Logger with specified key/value pairs
	With(kvs ...interface{}) Logger
	// Sync flush log buffers
	Sync() error
}

type logger struct {
	zap0 *zap.SugaredLogger
	zap  *zap.SugaredLogger

	nameJoiner     NameJoiner
	locationJoiner LocationJoiner

	name     string
	location string
}

var _ Logger = &logger{}

// mostly return origin + . + given
func nameJoineroiner(origin, given string) string {
	const sep = "."
	if given != "" {
		if origin != "" {
			return origin + sep + given
		}
		return given
	}
	return origin
}

// mostly return @, origin + / + given
func locationJoiner(origin, given string) (string, string) {
	const sep = "/"
	const key = "@"
	if given != "" {
		if origin != "" {
			return key, origin + "/" + given
		}
		return key, given
	}
	return key, origin
}

// NewLogger return a Logger, with name and location.
// name and location are optional, empty string means no name, no location.
// z0, nj, lj are optional too, default z0 is zap.S(), default nj is nameJoiner, default lj is locationJoiner.
// nj and lj will be used when Fork() invoked
func NewLogger(name, location string, z0 *zap.SugaredLogger, nj NameJoiner, lj LocationJoiner) Logger {
	return (&logger{
		zap0: z0,

		nameJoiner:     nj,
		locationJoiner: lj,
	}).fork(name, location)
}

func (l *logger) fork(name, location string) *logger {
	var l2 = &logger{
		zap0: l.zap0,
		zap:  l.zap0,

		nameJoiner:     l.nameJoiner,
		locationJoiner: l.locationJoiner,
	}

	if l2.zap0 == nil {
		l2.zap0 = zap.S()
		l2.zap = l2.zap0
	}

	if l2.nameJoiner == nil {
		l2.name = nameJoineroiner(l.name, name)
	} else {
		l2.name = l2.nameJoiner(l.name, name)
	}

	var field string
	if l2.locationJoiner == nil {
		field, l2.location = locationJoiner(l.location, location)
	} else {
		field, l2.location = l2.locationJoiner(l.location, location)
	}

	if l2.name != "" {
		l2.zap = l2.zap.Named(l2.name)
	}
	if l2.location != "" && field != "" {
		l2.zap = l2.zap.With(field, l2.location)
	}

	return l2
}

func (l *logger) Fork(name, location string) Logger {
	if l == nil {
		return nil
	}
	return l.fork(name, location)
}

func (l *logger) Name() string {
	return l.name
}

func (l *logger) Location() string {
	return l.location
}

func (l *logger) Debug(msg string, kvs ...interface{}) {
	l.zap.Debugw(msg, kvs...)
}

func (l *logger) Debugf(template string, args ...interface{}) {
	l.zap.Debugf(template, args...)
}

func (l *logger) Info(msg string, kvs ...interface{}) {
	l.zap.Infow(msg, kvs...)
}

func (l *logger) Infof(template string, args ...interface{}) {
	l.zap.Infof(template, args...)
}

func (l *logger) Warn(msg string, kvs ...interface{}) {
	l.zap.Warnw(msg, kvs...)
}

func (l *logger) Warnf(template string, args ...interface{}) {
	l.zap.Warnf(template, args...)
}

func (l *logger) Error(msg string, kvs ...interface{}) {
	l.zap.Errorw(msg, kvs...)
}

func (l *logger) Errorf(template string, args ...interface{}) {
	l.zap.Errorf(template, args...)
}

func (l *logger) Panic(msg string, kvs ...interface{}) {
	l.zap.Panicw(msg, kvs...)
}

func (l *logger) Panicf(template string, args ...interface{}) {
	l.zap.Panicf(template, args...)
}

func (l *logger) Fatal(msg string, kvs ...interface{}) {
	l.zap.Fatalw(msg, kvs...)
}

func (l *logger) Fatalf(template string, args ...interface{}) {
	l.zap.Fatalf(template, args...)
}

func (l *logger) With(kvs ...interface{}) Logger {
	if len(kvs) == 0 {
		return l
	}
	var l2 = l.fork("", "")
	l2.zap = l2.zap.With(kvs...)
	return l2
}

func (l *logger) Sync() error {
	return l.zap.Sync()
}
