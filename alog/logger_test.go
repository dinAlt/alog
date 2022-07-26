package alog

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestLogger_Stop(t *testing.T) {
	w := noopWriteFlusher{}
	logger := New(WithWriter(&w))

	go func() {
		err := logger.Run()
		if err != nil {
			t.Error(err)
		}
	}()

	err := logger.Stop(context.Background())
	if err != nil {
		t.Error(err)
	}

	if !w.flushed {
		t.Errorf("Logger.Stop() flush not called")
	}
}

func TestLogger_push(t *testing.T) {
	records := []logRecord{
		{
			format:    "test message",
			timestamp: time.Now(),
		},
		{
			format:    "test message %d %s",
			args:      []any{2, "b"},
			timestamp: time.Now().Add(20 * time.Second),
		},
	}

	loggerBuf := bytes.Buffer{}
	logger := New(WithWriter(&loggerBuf))

	go func() {
		err := logger.Run()
		if err != nil {
			t.Error(err)
		}
	}()
	for _, rec := range records {
		r := rec
		logger.push(&r)
	}
	err := logger.Stop(context.Background())
	if err != nil {
		t.Error(err)
	}

	recordsBuf := bytes.Buffer{}
	for _, r := range records {
		recordsBuf.Write(r.bytes())
	}
	recordBytes := recordsBuf.Bytes()
	loggerBytes := loggerBuf.Bytes()
	if !bytes.Equal(recordBytes, loggerBytes) {
		t.Errorf("Logger.push() want:\n %s got:\n %s", recordBytes, loggerBytes)
	}
}

type noopWriteFlusher struct {
	flushed bool
}

func (f *noopWriteFlusher) Write(b []byte) (int, error) {
	return len(b), nil
}

func (f *noopWriteFlusher) Flush() error {
	f.flushed = true
	return nil
}
