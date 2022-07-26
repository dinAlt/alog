package alog

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// DefaultCapacity of Logger channel
const DefaultCapacity = 500

// Logger is a thread safe channel based logger implementation. It should only be
// created with constructor (see New).
type Logger struct {
	writer io.Writer
	logC   chan *logRecord
	doneC  chan struct{}
}

// New returns pointer to newly created Logger with given Options (see options.go)
func New(options ...Option) *Logger {
	result := &Logger{
		doneC: make(chan struct{}),
	}

	for _, o := range options {
		o(result)
	}

	if result.writer == nil {
		result.writer = bufio.NewWriter(os.Stderr)
	}

	if result.logC == nil {
		result.logC = make(chan *logRecord, DefaultCapacity)
	}

	return result
}

// Run receives messages from logging channel and writes them to configured
// writer. It blocks caller goroutine until Stop is called.
// It returns error if underlying writer.Writer call wasn't succeed.
func (l *Logger) Run() error {
	for rec := range l.logC {
		_, err := l.writer.Write(rec.bytes())
		if err != nil {
			return fmt.Errorf("Logger.Run(): %w", err)
		}
	}
	close(l.doneC)
	return nil
}

// Printf enqueues message for logging
func (l *Logger) Printf(format string, a ...any) {
	record := &logRecord{
		format:    format,
		args:      a,
		timestamp: time.Now(),
	}
	l.push(record)
}

func (l *Logger) push(record *logRecord) {
	l.logC <- record
}

// Stop closes logger channel, whaits while all messages are processed.
// It also calls Flush on underlying writer if it implements Flusher (see flusher).
// Returns error if Flush call wasn't succeed.
func (l *Logger) Stop(ctx context.Context) error {
	close(l.logC)
	select {
	case <-l.doneC:
	case <-ctx.Done():
		return fmt.Errorf("Logger.Stop(): %w", ctx.Err())
	}

	flusher, _ := l.writer.(interface{ Flush() error })
	if flusher == nil {
		return nil
	}

	err := flusher.Flush()

	if err != nil {
		return fmt.Errorf("Logger.Stop(): %w", err)
	}

	return nil
}

// Flusher may be implemented by Logger writer.
type Flusher interface {
	Flush() error
}

type logRecord struct {
	args      []any
	format    string
	timestamp time.Time
}

func (r *logRecord) bytes() []byte {
	format := fmt.Sprintf("%s %s\n", r.timestamp.Format("2006/01/02 15:04:05"), r.format)
	message := fmt.Sprintf(format, r.args...)
	return []byte(message)
}
