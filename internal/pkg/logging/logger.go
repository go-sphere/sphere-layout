package logging

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-sphere/sphere-layout/internal/config"
	"github.com/go-sphere/sphere/log"
	"github.com/go-sphere/sphere/log/logbuffer"
	"github.com/go-sphere/sphere/log/zapx"
)

const logBufferCapacity = 4096

// New builds the process log backend, installs it as the global logger, and
// returns it with an idempotent closer (Sync, then close the zap file handle).
// Use Buffer to reach the dash tail API.
func New(conf zapx.Config) (log.Backend, func() error) {
	zapBackend := zapx.NewBackend(conf, log.WithAttrs(map[string]any{
		"version": config.BuildVersion,
	}))
	buffer := logbuffer.New(logBufferCapacity, log.WithMinLevel(bufferMinLevel(conf.Level)))
	backend := log.NewMultiBackend(zapBackend, buffer)
	log.InitWithBackends(backend)
	slog.SetDefault(zapBackend.SlogLogger())
	return backend, sync.OnceValue(func() error {
		return errors.Join(backend.Sync(), zapBackend.Close())
	})
}

// Buffer returns the in-memory tail mounted in backend, or nil.
func Buffer(b log.Backend) *logbuffer.Buffer {
	switch t := b.(type) {
	case *logbuffer.Buffer:
		return t
	case *log.MultiBackend:
		for _, child := range t.Backends() {
			if buf := Buffer(child); buf != nil {
				return buf
			}
		}
	}
	return nil
}

func bufferMinLevel(level string) log.Level {
	name := strings.ToLower(strings.TrimSpace(level))
	switch name {
	case "dpanic", "panic", "fatal":
		return log.LevelError
	}
	minLevel, ok := logbuffer.ParseLevel(name)
	if !ok {
		return log.LevelInfo
	}
	return minLevel
}
