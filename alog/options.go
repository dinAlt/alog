package alog

import "io"

// Option is interface which should be implemented by Logger options
type Option func(l *Logger)

// WithWriter options allow to create Logger with custom writer
func WithWriter(w io.Writer) Option {
	return func(l *Logger) {
		l.writer = w
	}
}

// WithCapacity sets Logger channel capacity
func WithCapacity(c int) Option {
	return func(l *Logger) {
		l.logC = make(chan *logRecord, c)
	}
}

