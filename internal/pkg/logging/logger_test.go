package logging

import (
	"log/slog"
	"testing"

	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
	"github.com/go-sphere/sphere/log/zapx"
)

func TestNewInstallsBufferAndCloseIsIdempotent(t *testing.T) {
	originalBackend := log.With().Backend()
	originalSlog := slog.Default()
	t.Cleanup(func() {
		log.InitWithBackends(originalBackend)
		slog.SetDefault(originalSlog)
	})

	conf := zapx.NewDefaultConfig()
	conf.Console.Disable = true
	backend, closer := New(conf)
	if backend == nil {
		t.Fatal("New returned a nil backend")
	}
	if closer == nil {
		t.Fatal("New returned a nil closer")
	}
	buf := Buffer(backend)
	if buf == nil {
		t.Fatal("Buffer(New()) returned nil")
	}

	log.Info("captured")
	entries, _ := buf.History(0, 1, log.LevelDebug)
	if len(entries) != 1 || entries[0].Message != "captured" {
		t.Fatalf("buffer entries = %+v, want captured log", entries)
	}

	if err := closer(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := closer(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestLoggerBufferHonorsZapLevelCase(t *testing.T) {
	originalBackend := log.With().Backend()
	originalSlog := slog.Default()
	t.Cleanup(func() {
		log.InitWithBackends(originalBackend)
		slog.SetDefault(originalSlog)
	})

	conf := zapx.NewDefaultConfig()
	conf.Console.Disable = true
	conf.Level = "DEBUG"
	backend, closer := New(conf)
	t.Cleanup(func() {
		_ = closer()
	})

	log.Debug("debug-line")
	entries, _ := Buffer(backend).History(0, 8, log.LevelDebug)
	if len(entries) != 1 || entries[0].Message != "debug-line" {
		t.Fatalf("buffer entries = %+v, want debug-line", entries)
	}
}

func TestBufferFindsRawRing(t *testing.T) {
	t.Parallel()
	buf := logbuffer.New(8)
	if got := Buffer(buf); got != buf {
		t.Fatalf("Buffer(raw) = %p, want %p", got, buf)
	}
	if Buffer(nil) != nil {
		t.Fatal("Buffer(nil) should be nil")
	}
}

func TestBufferFindsRingInMultiBackend(t *testing.T) {
	t.Parallel()
	buf := logbuffer.New(8)
	multi := log.NewMultiBackend(log.NewNopBackend(), buf)
	if got := Buffer(multi); got != buf {
		t.Fatalf("Buffer(multi) = %p, want %p", got, buf)
	}
}
