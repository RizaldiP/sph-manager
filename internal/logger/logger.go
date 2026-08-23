package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func New(logDir string) (*slog.Logger, func(), error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(filepath.Join(logDir, "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}
	w := safeMultiWriter{sinks: []io.Writer{f, os.Stdout}}
	lg := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	closer := func() { _ = f.Close() }
	return lg, closer, nil
}

type safeMultiWriter struct {
	sinks []io.Writer
}

func (m safeMultiWriter) Write(p []byte) (int, error) {
	for _, s := range m.sinks {
		_, _ = s.Write(p)
	}
	return len(p), nil
}
